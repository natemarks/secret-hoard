package tools

import (
	"log"
	"os"
)

// Logger provides simple logging for CLI applications
type Logger struct {
	debug   *log.Logger
	info    *log.Logger
	error   *log.Logger
	isDebug bool
}

// NewLogger creates a new Logger
func NewLogger(debugMode bool) *Logger {
	return &Logger{
		debug:   log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime),
		info:    log.New(os.Stdout, "", 0), // No prefix for info
		error:   log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime),
		isDebug: debugMode,
	}
}

// Debug logs a debug message (only if debug mode enabled)
func (l *Logger) Debug(format string, v ...any) {
	if l.isDebug {
		l.debug.Printf(format, v...)
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...any) {
	l.info.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...any) {
	l.error.Printf(format, v...)
}

// Fatal logs an error and exits with code 1
func (l *Logger) Fatal(format string, v ...any) {
	l.error.Printf(format, v...)
	os.Exit(1)
}
