package logging

import (
	"log"
	"os"
)

// LogLevel defines the level of logging.
type LogLevel int

const (
	// LevelError is the lowest level, only errors will be logged.
	LevelError LogLevel = iota
	// LevelInfo enables info logging.
	LevelInfo
	// LevelDebug enables debug logging.
	LevelDebug
)

// Logger is a configurable logger.
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger.
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs messages only if the 'Verbose' or 'Debug' flag is set.
func (l *Logger) Info(v ...interface{}) {
	if l.level >= LevelInfo {
		l.logger.SetPrefix("[INFO] ")
		l.logger.Println(v...)
	}
}

// Infof logs messages only if the 'Verbose' or 'Debug' flag is set.
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level >= LevelInfo {
		l.logger.SetPrefix("[INFO] ")
		l.logger.Printf(format, v...)
	}
}

// Debug logs messages only if the 'Debug' flag is set.
func (l *Logger) Debug(v ...interface{}) {
	if l.level >= LevelDebug {
		l.logger.SetPrefix("[DEBUG] ")
		l.logger.Println(v...)
	}
}

// Debugf logs messages only if the 'DebugFlag' flag is set.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level >= LevelDebug {
		l.logger.SetPrefix("[DEBUG] ")
		l.logger.Printf(format, v...)
	}
}

// Fatal is equivalent to l.logger.Fatal, but it respects the log level.
func (l *Logger) Fatal(v ...interface{}) {
	l.logger.SetPrefix("[FATAL] ")
	l.logger.Fatal(v...)
}

// Fatalf is equivalent to l.logger.Fatalf, but it respects the log level.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.SetPrefix("[FATAL] ")
	l.logger.Fatalf(format, v...)
}

// Warn logs messages at the INFO level with a WARN prefix.
func (l *Logger) Warn(v ...interface{}) {
	if l.level >= LevelInfo {
		l.logger.SetPrefix("[WARN] ")
		l.logger.Println(v...)
	}
}

// Warnf logs messages at the INFO level with a WARN prefix.
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level >= LevelInfo {
		l.logger.SetPrefix("[WARN] ")
		l.logger.Printf(format, v...)
	}
}
