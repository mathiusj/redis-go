package logger

import (
	"fmt"
	"log"
	"os"
)

// Level represents the logging level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger provides structured logging
type Logger struct {
	level  Level
	logger *log.Logger
}

var defaultLogger = &Logger{
	level:  LevelInfo,
	logger: log.New(os.Stdout, "", log.LstdFlags),
}

// SetLevel sets the global log level
func SetLevel(level Level) {
	defaultLogger.level = level
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	defaultLogger.log(LevelDebug, format, args...)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	defaultLogger.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	defaultLogger.log(LevelWarn, format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	defaultLogger.log(LevelError, format, args...)
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	prefix := ""
	switch level {
	case LevelDebug:
		prefix = "[DEBUG] "
	case LevelInfo:
		prefix = "[INFO] "
	case LevelWarn:
		prefix = "[WARN] "
	case LevelError:
		prefix = "[ERROR] "
	}

	msg := fmt.Sprintf(format, args...)
	l.logger.Printf("%s%s", prefix, msg)
}
