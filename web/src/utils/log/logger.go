/**
 * Licensed Materials - Property of PEG TECH INC
 *
 * (C) Copyright PEG TECH INC. 2024 All Rights Reserved
 * SPDX-License-Identifier: Apache-2.0

 * Contributors:
 *    bryan@raksmart.com - Initial implementation
 *
 *
 * Purpose: logging utilities
 *
**/

package log

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

const (
	pkgLogID      = "utils/log"
	defaultFormat = "%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfile} %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}"
	defaultLevel  = logging.INFO
)

var (
	logger *logging.Logger

	defaultOutput *os.File

	modules map[string]string // Holds the map of all modules and their respective log level

	lock sync.RWMutex
	once sync.Once
)

func init() {
	logger = logging.MustGetLogger(pkgLogID)
	Reset()
}

// InitLogger sets up the logging backend based on the provided log file.
// which will read following configurations from the viper instance.
// logging.log_dir: log directory
// logging.log_level: log level: overall log level or module specific log level
//
//	(e.g. "debug", "info", "warn", "error", "fatal", "panic")
//	(e.g. "[<module>[,<module>...]=]<level>[:[<module>[,<module>...]=]<level>...]")
//
// logging.max_size: maximum size of log file
// logging.max_backups: maximum number of old log files to retain
// logging.max_age: maximum number of days to retain old log files
func InitLogger(log_file string) {
	logger.Debugf("Initializing logger with log file: %s", log_file)
	if log_file == "" {
		log_file = "cl.log"
	}
	log_dir := viper.GetString("logging.log_dir")
	if log_dir == "" {
		log_dir = "/opt/cloudland/log"
	}
	log_file = fmt.Sprintf("%s/%s", viper.GetString("logging.log_dir"), log_file)
	log_level := viper.GetString("logging.log_level")
	if log_level == "" {
		log_level = "debug"
	}
	max_size := viper.GetInt("logging.max_size")
	if max_size == 0 {
		max_size = 100
	}
	max_backups := viper.GetInt("logging.max_backups")
	if max_backups == 0 {
		max_backups = 10
	}
	max_age := viper.GetInt("logging.max_age")
	if max_age == 0 {
		max_age = 30
	}
	format := viper.GetString("logging.format")
	if format == "" {
		format = defaultFormat
	}
	logger.Debugf("initializing logger with log file: %s, log level: %s, max size: %d, max backups: %d, max age: %d, format: '%s'",
		log_file, log_level, max_size, max_backups, max_age, format)

	initRollingBackend(log_file, max_size, max_backups, max_age, format)
	InitLogLevelFromSpec(log_level)
}

// Reset sets to logging to the defaults defined in this package.
func Reset() {
	modules = make(map[string]string)
	lock = sync.RWMutex{}

	defaultOutput = os.Stderr
	initBackend(SetFormat(defaultFormat), defaultOutput)
	InitLogLevelFromSpec("")
}

// SetFormat sets the logging format.
func SetFormat(formatSpec string) logging.Formatter {
	if formatSpec == "" {
		formatSpec = defaultFormat
	}
	return logging.MustStringFormatter(formatSpec)
}

// InitBackend sets up the logging backend based on
// the provided logging formatter and I/O writer.
func initBackend(formatter logging.Formatter, output io.Writer) {
	backend := logging.NewLogBackend(output, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, formatter)
	logging.SetBackend(backendFormatter).SetLevel(defaultLevel, "")
}

// initRollingBackend set rolling log backend
// maxSize is the maximum size in megabytes
// maxBackups is the maximum number of old log files to retain
// maxAge is the maximum number of days to retain old log files
func initRollingBackend(logfile string, maxSize int, maxBackups int, maxAge int, format string) {
	if logfile != "" {
		output := &lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge, //days
		}
		initBackend(SetFormat(defaultFormat), output)
	} else {
		initBackend(SetFormat(defaultFormat), defaultOutput)
	}
}

// DefaultLevel returns the fallback value for loggers to use if parsing fails.
func DefaultLevel() string {
	return defaultLevel.String()
}

// GetModuleLevel gets the current logging level for the specified module.
func GetModuleLevel(module string) string {
	// logging.GetLevel() returns the logging level for the module, if defined.
	// Otherwise, it returns the default logging level, as set by
	// `blogging/logging.go`.
	level := logging.GetLevel(module).String()
	return level
}

// SetModuleLevel sets the logging level for the modules that match the supplied
// regular expression. Can be used to dynamically change the log level for the
// module.
func SetModuleLevel(moduleRegExp string, level string) (string, error) {
	return setModuleLevel(moduleRegExp, level, true, false)
}

func setModuleLevel(moduleRegExp string, level string, isRegExp bool, revert bool) (string, error) {
	var re *regexp.Regexp
	logLevel, err := logging.LogLevel(level)
	if err != nil {
		logger.Warningf("Invalid logging level '%s' - ignored", level)
	} else {
		if !isRegExp || revert {
			logging.SetLevel(logLevel, moduleRegExp)
			logger.Debugf("Module '%s' logger enabled for log level '%s'", moduleRegExp, level)
		} else {
			re, err = regexp.Compile(moduleRegExp)
			if err != nil {
				logger.Warningf("Invalid regular expression: %s", moduleRegExp)
				return "", err
			}
			lock.Lock()
			defer lock.Unlock()
			for module := range modules {
				if re.MatchString(module) {
					logging.SetLevel(logging.Level(logLevel), module)
					modules[module] = logLevel.String()
					logger.Debugf("Module '%s' logger enabled for log level '%s'", module, logLevel)
				}
			}
		}
	}
	return logLevel.String(), err
}

// MustGetLogger is used in place of `logging.MustGetLogger` to allow us to
// store a map of all modules and submodules that have loggers in the system.
func MustGetLogger(module string) *logging.Logger {
	l := logging.MustGetLogger(module)
	lock.Lock()
	defer lock.Unlock()
	modules[module] = GetModuleLevel(module)
	return l
}

// InitLogLevelFromSpec initializes the logging based on the supplied spec. It is
// exposed externally so that consumers of the blogging package may parse their
// own logging specification. The logging specification has the following form:
//
//	[<module>[,<module>...]=]<level>[:[<module>[,<module>...]=]<level>...]
func InitLogLevelFromSpec(spec string) string {
	levelAll := defaultLevel
	var err error

	if spec != "" {
		fields := strings.Split(spec, ":")
		for _, field := range fields {
			split := strings.Split(field, "=")
			switch len(split) {
			case 1:
				if levelAll, err = logging.LogLevel(field); err != nil {
					logger.Warningf("Logging level '%s' not recognized, defaulting to '%s': %s", field, defaultLevel, err)
					levelAll = defaultLevel // need to reset cause original value was overwritten
				}
			case 2:
				// <module>[,<module>...]=<level>
				levelSingle, err := logging.LogLevel(split[1])
				if err != nil {
					logger.Warningf("Invalid logging level in '%s' ignored", field)
					continue
				}

				if split[0] == "" {
					logger.Warningf("Invalid logging override specification '%s' ignored - no module specified", field)
				} else {
					modules := strings.Split(split[0], ",")
					for _, module := range modules {
						logger.Debugf("Setting logging level for module '%s' to '%s'", module, levelSingle)
						logging.SetLevel(levelSingle, module)
					}
				}
			default:
				logger.Warningf("Invalid logging override '%s' ignored - missing ':'?", field)
			}
		}
	}

	logging.SetLevel(levelAll, "") // set the logging level for all modules

	// iterate through modules to reload their level in the modules map based on
	// the new default level
	for k := range modules {
		MustGetLogger(k)
	}
	// register blogging logger in the modules map
	MustGetLogger(pkgLogID)

	return levelAll.String()
}
