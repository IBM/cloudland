/*
Copyright PEG Tech Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package log

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	pkgLogID      = "utils/log"
	defaultFormat = "%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}"
)

var (
	sys_logger, logger *logrus.Logger
	defaultOutput      *os.File

	lock sync.RWMutex

	defaultLevel = logrus.DebugLevel
)

type UTCFormatter struct {
	logrus.TextFormatter // Embed the default TextFormatter
}

// Format formats the log entry with UTC timestamp.
func (f *UTCFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Time = entry.Time.UTC()        // Convert time to UTC
	return f.TextFormatter.Format(entry) // Use the embedded formatter
}

// Reset sets to logging to the defaults defined in this package.
func Reset() {
	defaultOutput = os.Stderr
	initBackend(GetDefaultFormater(defaultFormat), defaultOutput)
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
	logger.Debugf("initializing logger with log file: %s, log level: %s, max size: %d, max backups: %d, max age: %d",
		log_file, log_level, max_size, max_backups, max_age)

	var err error
	defaultLevel, err = logrus.ParseLevel(log_level)
	if err != nil {
		logger.Fatalf("Failed to parse log level: %s", log_level)
	}
	InitRollingBackend(log_file, max_size, max_backups, max_age)
}

// InitBackend sets up the logging backend based on
// the provided logging formatter and I/O writer.
func initBackend(formatter logrus.Formatter, output io.Writer) {
	if sys_logger == nil {
		sys_logger = logrus.New()
	}
	sys_logger.SetOutput(output)
	sys_logger.SetFormatter(formatter)
	sys_logger.SetLevel(defaultLevel)
}

// InitRollingBackend set rolling log backend
// maxSize is the maximum size in megabytes
// maxBackups is the maximum number of old log files to retain
// maxAge is the maximum number of days to retain old log files
func InitRollingBackend(logfile string, maxSize int, maxBackups int, maxAge int) {
	logger.Debugf("Initializing rolling log backend with log file: %s, max size: %d, max backups: %d, max age: %d",
		logfile, maxSize, maxBackups, maxAge)
	if logfile != "" {
		output := &lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge, //days
		}
		initBackend(GetDefaultFormater(defaultFormat), output)
	} else {
		initBackend(GetDefaultFormater(defaultFormat), defaultOutput)
	}
}

// GetDefaultFormater returns the default logging format.
func GetDefaultFormater(formatSpec string) logrus.Formatter {
	if formatSpec == "" {
		formatSpec = defaultFormat
	}
	return &UTCFormatter{
		logrus.TextFormatter{
			FullTimestamp:          true,
			DisableColors:          false,
			DisableSorting:         false,
			DisableLevelTruncation: false,
			QuoteEmptyFields:       true,
			ForceColors:            true,
			DisableTimestamp:       false,
			TimestampFormat:        "2006-01-02 15:04:05.000 UTC",
		},
	}
}

// DefaultLevel returns the fallback value for loggers to use if parsing fails.
func DefaultLevel() string {
	return defaultLevel.String()
}

// MustGetLogger is used in place of `logging.MustGetLogger` to allow us to
// store a map of all modules and submodules that have loggers in the system.
func MustGetLogger(module string) *logrus.Logger {
	lock.Lock()
	defer lock.Unlock()
	if sys_logger == nil {
		InitRollingBackend("", 0, 0, 0)
	}
	l := sys_logger.WithField("module", module).Logger
	return l
}

func init() {
	lock = sync.RWMutex{}
	logger = logrus.New()
	logger.SetLevel(defaultLevel)
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(GetDefaultFormater(defaultFormat))
}
