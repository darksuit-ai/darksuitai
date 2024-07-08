package exceptions

import (
	"fmt"
	"github.com/darksuit-ai/darksuitai/types"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Logger represents a logging interface.
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
}

// FileLogger implements the Logger interface using a file.
type FileLogger struct {
	file *os.File
}

// NewFileLogger creates a new FileLogger instance.
func NewFileLogger(filename string) (*FileLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &FileLogger{file: file}, nil
}

// Debug logs a debug message.
func (l *FileLogger) Debug(v ...interface{}) {
	l.log("DEBUG", v...)
}

// Info logs an info message.
func (l *FileLogger) Info(v ...interface{}) {
	l.log("INFO", v...)
}

// Warn logs a warning message.
func (l *FileLogger) Warn(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	scriptName := filepath.Base(funcName)
	l.log("WARNING", fmt.Sprintf("%v\n%s - %s:%d\n%s", v, funcName, file, line, scriptName))
}

// Error logs an error message.
func (l *FileLogger) Error(v ...interface{}) {
	l.log("ERROR", v...)
}

// Fatal logs a fatal message and exits the program.
func (l *FileLogger) Fatal(v ...interface{}) {
	l.log("FATAL", v...)
	os.Exit(1)
}

// log logs a message with the specified level.
func (l *FileLogger) log(level string, v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.999999-0700 MST")
	message := fmt.Sprintf("%s - %s - %s\n", timestamp, level, fmt.Sprint(v...))
	if _, err := l.file.WriteString(message); err != nil {
		log.Fatal(err)
	}
}

// Close closes the underlying file.
func (l *FileLogger) Close() error {
	return l.file.Close()
}

type CustomLoggers struct {
	System  *FileLogger
	UserOps *FileLogger
	ToolOps *FileLogger
	AiOps   *FileLogger
}

var Loggers CustomLoggers

func init() {
	// Get the current working directory
	cd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Create the logs folder if it doesn't exist
	logsFolder := filepath.Join(cd, "logs")
	if err := os.MkdirAll(logsFolder, 0755); err != nil {
		log.Fatal(err)
	}

	// Create log files in the root directory
	logFiles := []string{"system.log", "aiops.log", "int_aiops.log", "toolops.log", "usersops.log"}
	for _, filename := range logFiles {
		logFile := filepath.Join(logsFolder, filename)
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			if err := os.WriteFile(logFile, []byte(""), 0644); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Setup loggers
	systemLogger, err := NewFileLogger(filepath.Join(logsFolder, "system.log"))
	if err != nil {
		log.Fatal(err)
	}

	aipipelineLogger, err := NewFileLogger(filepath.Join(logsFolder, "aiops.log"))
	if err != nil {
		log.Fatal(err)
	}

	// intAipipelineLogger, err := NewFileLogger(filepath.Join(logsFolder, "int_aiops.log"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	toolLogger, err := NewFileLogger(filepath.Join(logsFolder, "toolops.log"))
	if err != nil {
		log.Fatal(err)
	}

	usersopsLogger, err := NewFileLogger(filepath.Join(logsFolder, "usersops.log"))
	if err != nil {
		log.Fatal(err)
	}

	Loggers.System = systemLogger
	Loggers.ToolOps = toolLogger
	Loggers.UserOps = usersopsLogger
	Loggers.AiOps = aipipelineLogger

}

func IOLogger(rc int, detail, ext_ref string) types.Error {
	var error types.Error
	error.ResponseCode = rc
	error.Message = "Failed"
	error.Detail = detail
	error.ExternalReference = ext_ref

	return error
}
