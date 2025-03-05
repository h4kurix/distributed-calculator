package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	debugMode bool
	logger    *log.Logger
)

// Init initializes the logger with the given log level
func Init(level string) {
	logger = log.New(os.Stdout, "", log.LstdFlags)

	switch level {
	case "debug":
		debugMode = true
	case "info":
		debugMode = false
	default:
		debugMode = false
	}

	Info("Logger initialized with level: %s", level)
}

// Debug logs debug messages (only in debug mode)
func Debug(format string, v ...interface{}) {
	if debugMode {
		logMessage("DEBUG", format, v...)
	}
}

// Info logs informational messages
func Info(format string, v ...interface{}) {
	logMessage("INFO", format, v...)
}

// Error logs error messages
func Error(format string, v ...interface{}) {
	logMessage("ERROR", format, v...)
}

// logMessage formats and logs a message with the given level
func logMessage(level, format string, v ...interface{}) {
	// If logger is not initialized, initialize with defaults
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}
	
	message := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logger.Printf("[%s] %s: %s", timestamp, level, message)
}