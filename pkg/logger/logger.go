package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	logger     *log.Logger
	level      LogLevel = INFO
	mu         sync.RWMutex
	logFile    *os.File
	logEntries []LogEntry
	maxEntries = 1000
)

func Init(logPath string) error {
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	logger = log.New(logFile, "", log.LstdFlags)
	logEntries = make([]LogEntry, 0)

	logMessageInternal(INFO, "日志系统初始化完成")
	return nil
}

func SetLevel(l LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	level = l
}

func SetMaxEntries(max int) {
	mu.Lock()
	defer mu.Unlock()
	maxEntries = max
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
	logMessageInternal(l, format, args...)
}

func logMessageInternal(l LogLevel, format string, args ...interface{}) {
	if l < level {
		return
	}

	prefix := ""
	levelStr := ""
	switch l {
	case DEBUG:
		prefix = "[DEBUG] "
		levelStr = "DEBUG"
	case INFO:
		prefix = "[INFO] "
		levelStr = "INFO"
	case WARN:
		prefix = "[WARN] "
		levelStr = "WARN"
	case ERROR:
		prefix = "[ERROR] "
		levelStr = "ERROR"
	}

	message := fmt.Sprintf(format, args...)
	fullMessage := prefix + message

	if logger != nil {
		logger.Println(fullMessage)
	} else {
		log.Println(fullMessage)
	}

	entry := LogEntry{
		Level:     levelStr,
		Message:   message,
		Timestamp: time.Now(),
	}

	logEntries = append(logEntries, entry)

	if len(logEntries) > maxEntries {
		logEntries = logEntries[len(logEntries)-maxEntries:]
	}
}

func GetLogs() []LogEntry {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]LogEntry, len(logEntries))
	copy(result, logEntries)
	return result
}

func GetLogsSince(since time.Time) []LogEntry {
	mu.RLock()
	defer mu.RUnlock()

	var result []LogEntry
	for _, entry := range logEntries {
		if entry.Timestamp.After(since) {
			result = append(result, entry)
		}
	}
	return result
}

func ClearLogs() {
	mu.Lock()
	defer mu.Unlock()
	logEntries = make([]LogEntry, 0)
	logMessageInternal(INFO, "日志已清除")
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	logMessageInternal(INFO, "日志系统关闭")
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	logger = nil
}
