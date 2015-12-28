package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	LogDebug = iota
	LogInfo
	LogWarning
	LogError
	LogFatal
)

var logLevelStrings = [...]string{
	"DEBUG",
	"INFO",
	"WARNING",
	"ERROR",
	"FATAL",
}

var logLevel = LogWarning

// Early logging setup.
// Because we setup logging here, no other init functions should use loggers
func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func ResetLog(settings *ScanSettings) {
	if len(settings.LogfilePath) > 0 {
		if fp, err := os.Create(settings.LogfilePath); err == nil {
			log.SetOutput(fp)
		} else {
			Logf(LogError, "Unable to open logfile %s.", settings.LogfilePath)
		}
	}
	if len(settings.LogLevel) > 0 {
		SetLogLevel(settings.LogLevel)
	}
}

func Logf(level int, format string, args ...interface{}) {
	if level < logLevel {
		return
	}
	msg := fmt.Sprintf(format, args...)
	msg = fmt.Sprintf("[%s] %s", logLevelStrings[level], msg)
	log.Output(2, msg)
}

func SetLogLevel(level string) {
	level = strings.ToLower(level)
	for i, ll := range logLevelStrings {
		if strings.ToLower(ll) == level {
			logLevel = i
			return
		}
	}
}
