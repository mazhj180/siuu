package logger

import (
	"bufio"
	"fmt"
	"os"
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
	proxyLogLevel = InfoLevel

	systemLogger   *asyncLogger
	systemLogLevel = InfoLevel
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
	fout, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(fmt.Errorf("init log file error: %v \n", err))
	}

	proxyLogger = &asyncLogger{
		logCh:   make(chan string, 30),
		done:    make(chan struct{}),
		file:    fout,
		writer:  bufio.NewWriter(fout),
		maxSize: maxSize,
		logFile: filename,
	}
	proxyLogLevel = level
	go proxyLogger.startWriter()
}

func InitSystemLog(filename string, maxSize Unit, level Level) {
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
	if a.file != nil {
		_ = a.writer.Flush()
		_ = a.file.Close()
	}

	fout, err := os.OpenFile(a.logFile, os.O_WRONLY|os.O_TRUNC|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	a.currentSize = 0

	a.file = fout
	a.writer = bufio.NewWriter(fout)

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
			if n, err := a.writer.WriteString(msg + "\n"); err != nil {
				_, _ = os.Stderr.WriteString(fmt.Sprintf("[ERROR] log write wrong : %s\n", err))
			} else {
				a.currentSize += n
			}
		case <-a.done:
			for len(a.logCh) > 0 {
				msg := <-a.logCh
				_, _ = a.writer.WriteString(msg + "\n")
			}
			_ = a.writer.Flush()
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

func PDebug(format string, v ...any) {
	if proxyLogLevel > DebugLevel {
		return
	}
	proxyLogger.logCh <- fmt.Sprintf("[Debug]"+format, v...)
}

func PInfo(format string, v ...any) {
	if proxyLogLevel > InfoLevel {
		return
	}
	proxyLogger.logCh <- fmt.Sprintf("[Info]"+format, v...)
}

func PWarn(format string, v ...any) {
	if proxyLogLevel > WarnLevel {
		return
	}
	proxyLogger.logCh <- fmt.Sprintf("[Warn]"+format, v...)
}

func PError(format string, v ...any) {
	if proxyLogLevel > ErrorLevel {
		return
	}
	proxyLogger.logCh <- fmt.Sprintf("[Error]"+format, v...)
}

func SDebug(format string, v ...any) {
	if systemLogLevel > DebugLevel {
		return
	}
	systemLogger.logCh <- fmt.Sprintf("[Debug]"+format, v...)
}

func SInfo(format string, v ...any) {
	if systemLogLevel > InfoLevel {
		return
	}
	systemLogger.logCh <- fmt.Sprintf("[Info]"+format, v...)
}

func SWarn(format string, v ...any) {
	if systemLogLevel > WarnLevel {
		return
	}
	systemLogger.logCh <- fmt.Sprintf("[Warn]"+format, v...)
}

func SError(format string, v ...any) {
	if systemLogLevel > ErrorLevel {
		return
	}
	systemLogger.logCh <- fmt.Sprintf("[Error]"+format, v...)
}
