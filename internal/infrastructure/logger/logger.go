package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	return log
}

// SetLevel sets the logging level
func SetLevel(level string) {
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

// Fields type, used to pass to `WithFields`
type Fields logrus.Fields

// Debug logs a message at level Debug
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Info logs a message at level Info
func Info(args ...interface{}) {
	log.Info(args...)
}

// Warn logs a message at level Warn
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Error logs a message at level Error
func Error(args ...interface{}) {
	log.Error(args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// WithFields adds fields to the logging context
func WithFields(fields Fields) *logrus.Entry {
	return log.WithFields(logrus.Fields(fields))
}
