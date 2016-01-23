// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
var defaultLogger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

func ResetLog(settings *ScanSettings) {
	if len(settings.LogfilePath) > 0 {
		if fp, err := os.Create(settings.LogfilePath); err == nil {
			defaultLogger = log.New(fp, "", log.Ldate|log.Ltime|log.Lshortfile)
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
	defaultLogger.Output(2, msg)
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
