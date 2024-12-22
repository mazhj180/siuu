package util

import (
	"evil-gopher/logger"
	"fmt"
	"strings"
)

func LogLevel(level string) logger.Level {
	level = strings.ToUpper(level)
	switch level {
	case "DEBUG":
		return logger.DebugLevel
	case "INFO":
		return logger.InfoLevel
	case "WARN":
		return logger.WarnLevel
	case "ERROR":
		return logger.ErrorLevel
	default:
		panic(fmt.Sprintf("invalid log level: %s", level))
	}
}
