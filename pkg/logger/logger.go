package logger

import (
	"log"
	"os"
	"sync"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	logger      *log.Logger
	level       LogLevel = INFO
	mu          sync.Mutex
	logFile     *os.File
)

func Init(logPath string) error {
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	logger = log.New(logFile, "", log.LstdFlags)
	return nil
}

func SetLevel(l LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	level = l
}

func Debug(format string, args ...interface{}) {
	logMessage(DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	logMessage(INFO, format, args...)
}

func Warn(format string, args ...interface{}) {
	logMessage(WARN, format, args...)
}

func Error(format string, args ...interface{}) {
	logMessage(ERROR, format, args...)
}

func logMessage(l LogLevel, format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if l < level {
		return
	}

	prefix := ""
	switch l {
	case DEBUG:
		prefix = "[DEBUG] "
	case INFO:
		prefix = "[INFO] "
	case WARN:
		prefix = "[WARN] "
	case ERROR:
		prefix = "[ERROR] "
	}

	message := prefix + format

	if logger != nil {
		logger.Printf(message, args...)
	} else {
		log.Printf(message, args...)
	}
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
