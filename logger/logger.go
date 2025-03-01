package logger

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Unit int

const (
	B  Unit = 1
	KB      = 1024 * B
	MB      = 1024 * KB
	GB      = 1024 * MB
)

var (
	proxyLogger   *asyncLogger
	proxyLogLevel = DebugLevel

	systemLogger   *asyncLogger
	systemLogLevel = DebugLevel
)

type asyncLogger struct {
	logCh       chan string
	done        chan struct{}
	file        *os.File
	writer      *bufio.Writer
	maxSize     Unit
	currentSize int
	logFile     string
}

func InitProxyLog(filename string, maxSize Unit, level Level) {

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		dir := path.Dir(filename)
		if err = os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Errorf("init log file error: %v \n", err))
		}
	}

	fout, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(fmt.Errorf("init log file error: %v \n", err))
	}

	proxyLogger = &asyncLogger{
		logCh:   make(chan string, 30),
		done:    make(chan struct{}),
		file:    fout,
		maxSize: maxSize,
		logFile: filename,
	}
	proxyLogLevel = level
	go proxyLogger.startWriter()
}

func InitSystemLog(filename string, maxSize Unit, level Level) {

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		dir := path.Dir(filename)
		if err = os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Errorf("init log file error: %v \n", err))
		}
	}

	fout, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(fmt.Errorf("init log file error: %v \n", err))
	}

	systemLogger = &asyncLogger{
		logCh:   make(chan string, 30),
		done:    make(chan struct{}),
		file:    fout,
		writer:  bufio.NewWriter(fout),
		maxSize: maxSize,
		logFile: filename,
	}
	systemLogLevel = level
	go systemLogger.startWriter()
}

func (a *asyncLogger) rotate() error {

	if a.writer != nil {
		_ = a.writer.Flush()
	}

	if a.file != nil {
		_ = a.file.Close()
	}

	fout, err := os.OpenFile(a.logFile, os.O_WRONLY|os.O_TRUNC|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	a.currentSize = 0

	a.file = fout
	if a.writer != nil {
		a.writer = bufio.NewWriter(fout)
	}

	return nil
}

func (a *asyncLogger) startWriter() {
	for {
		select {
		case msg := <-a.logCh:
			size := len(msg)
			if a.currentSize+size > int(a.maxSize) {
				if err := a.rotate(); err != nil {
					_, _ = os.Stderr.WriteString(fmt.Sprintf("[ERROR] log rotate wrong : %s\n", err))
				}
			}
			if a.writer != nil {
				if n, err := a.writer.WriteString(msg); err != nil {
					_, _ = os.Stderr.WriteString(fmt.Sprintf("[ERROR] log write wrong : %s\n", err))
				} else {
					a.currentSize += n
				}
			} else {
				if n, err := a.file.WriteString(msg); err != nil {
					_, _ = os.Stderr.WriteString(fmt.Sprintf("[ERROR] log write wrong : %s\n", err))
				} else {
					a.currentSize += n
				}
			}

		case <-a.done:
			for len(a.logCh) > 0 {
				msg := <-a.logCh
				if a.writer != nil {
					_, _ = a.writer.WriteString(msg)
				} else {
					_, _ = a.file.WriteString(msg)
				}
			}
			if a.writer != nil {
				_ = a.writer.Flush()
			}
			_ = a.file.Close()
			a.writer = nil
			close(a.logCh)
			return
		}
	}
}

func Close() {
	if proxyLogger != nil {
		proxyLogger.done <- struct{}{}
	}
	if systemLogger != nil {
		systemLogger.done <- struct{}{}
	}
}

func timePrefix() string {
	t := time.Now()
	return t.Format("2006-01-02 15:04:05")
}

func LogLevel(level string) Level {
	level = strings.ToUpper(level)
	switch level {
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	default:
		panic(fmt.Sprintf("invalid log level: %s", level))
	}
}

func PDebug(format string, v ...any) {
	if proxyLogLevel > DebugLevel {
		return
	}
	if proxyLogger == nil {
		_, _ = os.Stdout.WriteString(fmt.Sprintf(timePrefix()+"[Debug]"+format+"\n", v...))
		return
	}
	proxyLogger.logCh <- fmt.Sprintf(timePrefix()+"[proxy]-[Debug]"+format+"\n", v...)
}

func PInfo(format string, v ...any) {
	if proxyLogLevel > InfoLevel {
		return
	}
	if proxyLogger == nil {
		_, _ = os.Stdout.WriteString(fmt.Sprintf(timePrefix()+"[proxy]-[Info]"+format+"\n", v...))
		return
	}
	proxyLogger.logCh <- fmt.Sprintf(timePrefix()+"[Info]"+format+"\n", v...)
}

func PWarn(format string, v ...any) {
	if proxyLogLevel > WarnLevel {
		return
	}
	if proxyLogger == nil {
		_, _ = os.Stdout.WriteString(fmt.Sprintf(timePrefix()+"[proxy]-[Warn]"+format+"\n", v...))
		return
	}
	proxyLogger.logCh <- fmt.Sprintf(timePrefix()+"[Warn]"+format+"\n", v...)
}

func PError(format string, v ...any) {
	if proxyLogLevel > ErrorLevel {
		return
	}
	if proxyLogger == nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf(timePrefix()+"[proxy]-[Error]"+format+"\n", v...))
		return
	}
	proxyLogger.logCh <- fmt.Sprintf(timePrefix()+"[Error]"+format+"\n", v...)
}

func SDebug(format string, v ...any) {
	if systemLogLevel > DebugLevel {
		return
	}
	if systemLogger == nil {
		_, _ = os.Stdout.WriteString(fmt.Sprintf(timePrefix()+"[system]-[Debug]"+format+"\n", v...))
		return
	}
	systemLogger.logCh <- fmt.Sprintf(timePrefix()+"[Debug]"+format+"\n", v...)
}

func SInfo(format string, v ...any) {
	if systemLogLevel > InfoLevel {
		return
	}
	if systemLogger == nil {
		_, _ = os.Stdout.WriteString(fmt.Sprintf(timePrefix()+"[system]-[Info]"+format+"\n", v...))
		return
	}
	systemLogger.logCh <- fmt.Sprintf(timePrefix()+"[Info]"+format+"\n", v...)
}

func SWarn(format string, v ...any) {
	if systemLogLevel > WarnLevel {
		return
	}
	if systemLogger == nil {
		_, _ = os.Stdout.WriteString(fmt.Sprintf(timePrefix()+"[system]-[Warn]"+format+"\n", v...))
		return
	}
	systemLogger.logCh <- fmt.Sprintf(timePrefix()+"[Warn]"+format+"\n", v...)
}

func SError(format string, v ...any) {
	if systemLogLevel > ErrorLevel {
		return
	}
	if systemLogger == nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf(timePrefix()+"[system]-[Error]"+format+"\n", v...))
		return
	}
	systemLogger.logCh <- fmt.Sprintf(timePrefix()+"[Error]"+format+"\n", v...)
}
