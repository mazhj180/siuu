package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"siuu/pkg/logger/splitter"
	"sync"
	"time"
)

// LogLevel log level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String return the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func LevelString(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	}
	return INFO
}

// Color return the color code of the log level
func (l LogLevel) Color() string {
	switch l {
	case DEBUG:
		return "\033[36m" // cyan
	case INFO:
		return "\033[32m" // green
	case WARN:
		return "\033[33m" // yellow
	case ERROR:
		return "\033[31m" // red
	case FATAL:
		return "\033[35m" // purple
	default:
		return "\033[0m" // default
	}
}

// Logger logger
type Logger struct {
	level      LogLevel
	console    bool
	file       bool
	logDir     string
	dateFormat string
	mu         sync.Mutex
	fileWriter *os.File
	async      bool
	asyncChan  chan *logEntry
	wg         sync.WaitGroup
	closed     bool
	splitter   splitter.Splitter // splitter
	baseName   string            // base file name
}

// logEntry log entry
type logEntry struct {
	level   LogLevel
	message string
	time    time.Time
}

// Config log configuration
type Config struct {
	Level      LogLevel
	Console    bool
	File       bool
	LogDir     string
	DateFormat string
	Async      bool
	BaseName   string
	Splitter   splitter.Splitter // splitter
}

// DefaultConfig default configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      INFO,
		Console:    true,
		File:       true,
		LogDir:     "logs",
		DateFormat: "2006-01-02",
		Async:      false,
		BaseName:   "siuu",
		Splitter:   nil, // default no splitter, no split
	}
}

// New create new logger
func New(config *Config) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	logger := &Logger{
		level:      config.Level,
		console:    config.Console,
		file:       config.File,
		logDir:     config.LogDir,
		dateFormat: config.DateFormat,
		async:      config.Async,
		splitter:   config.Splitter,
		baseName:   config.BaseName,
	}

	if config.File {
		if err := logger.ensureLogDir(); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		if err := logger.openLogFileWithLock(); err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	if config.Async {
		logger.asyncChan = make(chan *logEntry, 1000)
		logger.wg.Add(1)
		go logger.asyncWorker()
	}

	return logger, nil
}

// ensureLogDir ensure log directory exists
func (l *Logger) ensureLogDir() error {
	return os.MkdirAll(l.logDir, 0755)
}

// openLogFile open log file (internal method, assume already holding lock)
func (l *Logger) openLogFile() error {
	// only check if splitter is not nil when it is not nil
	if l.fileWriter != nil && l.splitter != nil {
		// here we pass an empty string as message, because we only need to check if we need to split
		if !l.splitter.ShouldSplit(l.fileWriter, "") {
			return nil
		}
	}

	// close old file
	if l.fileWriter != nil {
		l.fileWriter.Close()
	}

	// get new file name
	var filename string
	if l.splitter != nil {
		filename = l.splitter.GetNextFileName(l.logDir, l.baseName)
	} else {
		// when splitter is nil, use fixed file name, no split
		filename = filepath.Join(l.logDir, l.baseName)
	}

	// create new log file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.fileWriter = file
	return nil
}

// openLogFileWithLock file open method with lock (external call)
func (l *Logger) openLogFileWithLock() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.openLogFile()
}

// asyncWorker asynchronous worker
func (l *Logger) asyncWorker() {
	defer l.wg.Done()

	for entry := range l.asyncChan {
		if l.closed {
			break
		}
		l.writeLog(entry.level, entry.message, entry.time)
	}
}

// writeLog write log
func (l *Logger) writeLog(level LogLevel, message string, t time.Time) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// check if need to switch log file
	if l.file {
		// only check if splitter is not nil when it is not nil
		if l.splitter != nil && l.fileWriter != nil {
			if l.splitter.ShouldSplit(l.fileWriter, message) {
				l.openLogFile()
				// execute cleanup operation
				l.splitter.Cleanup(l.logDir, l.baseName)
			}
		}
		// removed old date split logic, no split when splitter is nil
	}

	timestamp := t.Format("2006-01-02 15:04:05.000")
	logMessage := fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), message)

	// output to console
	if l.console {
		colorReset := "\033[0m"
		fmt.Printf("%s%s%s\n", level.Color(), logMessage, colorReset)
	}

	// output to file
	if l.file && l.fileWriter != nil {
		fmt.Fprintln(l.fileWriter, logMessage)
	}
}

// log record log
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if l.closed {
		return
	}

	message := fmt.Sprintf(format, args...)
	t := time.Now()

	if l.async {
		select {
		case l.asyncChan <- &logEntry{level: level, message: message, time: t}:
		default:
			// channel is full, write directly
			l.writeLog(level, message, t)
		}
	} else {
		l.writeLog(level, message, t)
	}
}

// Debug record debug log
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info record info log
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn record warning log
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error record error log
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal record fatal error log and exit program
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

// SetLevel set log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Close close logger
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	l.closed = true
	if l.async {
		close(l.asyncChan)
		l.wg.Wait()
	}

	if l.fileWriter != nil {
		return l.fileWriter.Close()
	}

	return nil
}

func (l *Logger) IsClosed() bool {
	return l.closed
}

// Sync sync all pending logs
func (l *Logger) Sync() {
	if l.async {
		// wait for all messages in the channel to be processed
		// note: here we cannot close the channel, because there may be new logs later
		// we just wait for the messages in the current queue to be processed
		for len(l.asyncChan) > 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// global logger
var globalLogger *Logger
var globalMu sync.Mutex

// Init initialize global logger
func Init(config *Config) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalLogger != nil {
		globalLogger.Close()
	}

	logger, err := New(config)
	if err != nil {
		return err
	}

	globalLogger = logger
	return nil
}

// GetLogger get global logger
func GetLogger() *Logger {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalLogger == nil {
		globalLogger, _ = New(DefaultConfig())
	}

	return globalLogger
}

// global convenience methods
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GetLogger().Fatal(format, args...)
}
