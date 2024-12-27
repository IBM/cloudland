/*
Copyright PEG Tech Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package log

import (
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	pkgLogID      = "utils/log"
	defaultFormat = "%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}"
)

var (
	logger        *logrus.Logger
	defaultOutput *os.File

	moduleLevels map[string]string
	lock         sync.RWMutex

	defaultLevel = logrus.DebugLevel
)

// Reset sets to logging to the defaults defined in this package.
func Reset() {
	moduleLevels = make(map[string]string)
	lock = sync.RWMutex{}

	defaultOutput = os.Stderr
	InitBackend(GetDefaultFormater(defaultFormat), defaultOutput)
}

// InitBackend sets up the logging backend based on
// the provided logging formatter and I/O writer.
func InitBackend(formatter logrus.Formatter, output io.Writer) {
	logrus.SetOutput(output)
	logrus.SetFormatter(formatter)
	logrus.SetLevel(defaultLevel)
}

// InitRollingBackend set rolling log backend
// maxSize is the maximum size in megabytes
// maxBackups is the maximum number of old log files to retain
// maxAge is the maximum number of days to retain old log files
func InitRollingBackend(logfile string, maxSize int, maxBackups int, maxAge int) {
	if logfile != "" {
		output := &lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge, //days
		}
		InitBackend(GetDefaultFormater(defaultFormat), output)
	} else {
		InitBackend(GetDefaultFormater(defaultFormat), defaultOutput)
	}
}

// GetDefaultFormater returns the default logging format.
func GetDefaultFormater(formatSpec string) logrus.Formatter {
	if formatSpec == "" {
		formatSpec = defaultFormat
	}
	return &logrus.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
		DisableColors:          false,
		DisableSorting:         false,
		DisableLevelTruncation: false,
		QuoteEmptyFields:       true,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "@level",
			logrus.FieldKeyFunc:  "@caller",
			logrus.FieldKeyMsg:   "@message",
		},
	}
}

// DefaultLevel returns the fallback value for loggers to use if parsing fails.
func DefaultLevel() string {
	return defaultLevel.String()
}

func GetLogger(module string) *logrus.Logger {
	return logrus.WithField("module", module).Logger
}

// GetModuleLevel gets the current logging level for the specified module.
func GetModuleLevel(module string) string {
	return moduleLevels[module]
}

// SetModuleLevel sets the logging level for the modules that match the supplied
// regular expression. Can be used to dynamically change the log level for the
// module.
func SetModuleLevel(moduleRegExp string, level string) error {
	return setModuleLevel(moduleRegExp, level, true, false)
}

func setModuleLevel(moduleRegExp string, level string, isRegExp bool, revert bool) error {
	if !isRegExp || revert {
		moduleLevels[moduleRegExp] = level
		logger.Debugf("Module '%s' logger enabled for log level '%s'", moduleRegExp, level)
	} else {
		re, err := regexp.Compile(moduleRegExp)
		if err != nil {
			logger.Warningf("Invalid regular expression: %s", moduleRegExp)
			return err
		}
		lock.Lock()
		defer lock.Unlock()
		for module := range moduleLevels {
			if re.MatchString(module) {
				moduleLevels[module] = level
				logger.Debugf("Module '%s' logger enabled for log level '%s'", module, level)
			}
		}
	}
	return nil
}

func setLevel(module string, l *logrus.Logger) {
	level := GetModuleLevel(module)
	if level != "" {
		parsedLevel, err := logrus.ParseLevel(level)
		if err != nil {
			l.SetLevel(defaultLevel) // Default to Info level if parsing fails
		} else {
			l.SetLevel(parsedLevel)
		}
	} else {
		l.SetLevel(defaultLevel)
	}
}

// MustGetLogger is used in place of `logging.MustGetLogger` to allow us to
// store a map of all modules and submodules that have loggers in the system.
func MustGetLogger(module string) *logrus.Logger {
	l := logrus.WithField("module", module).Logger
	lock.Lock()
	defer lock.Unlock()

	setLevel(module, l)
	return l
}

// InitFromSpec initializes the logging based on the supplied spec. It is
// exposed externally so that consumers of the blogging package may parse their
// own logging specification. The logging specification has the following form:
//
//	[<module>[,<module>...]=]<level>[:[<module>[,<module>...]=]<level>...]
func InitModuleLevelsFromSpec(spec string) {
	levelAll := defaultLevel

	if spec != "" {
		if strings.Index(spec, ":") == -1 && strings.Index(spec, "=") == -1 {
			parsedLevel, err := logrus.ParseLevel(spec)
			if err != nil {
				logger.Warningf("Invalid logging override specification '%s' ignored - invalid level", spec)
			} else {
				defaultLevel = parsedLevel
			}
		}
		fields := strings.Split(spec, ":")
		for _, field := range fields {
			split := strings.Split(field, "=")
			switch len(split) {
			case 1:
				SetModuleLevel(split[0], levelAll.String())
			case 2:
				// <module>[,<module>...]=<level>
				levelSingle := split[1]
				if split[0] == "" {
					logger.Warningf("Invalid logging override specification '%s' ignored - no module specified", field)
				} else {
					modules := strings.Split(split[0], ",")
					for _, module := range modules {
						logger.Debugf("Setting logging level for module '%s' to '%s'", module, levelSingle)
						SetModuleLevel(module, levelSingle)
					}
				}
			default:
				logger.Warningf("Invalid logging override '%s' ignored - missing ':'?", field)
			}
		}
	}
}

func init() {
	logger = logrus.New()
	Reset()
}
