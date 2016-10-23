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

package logging

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
	logLevelMax
)

var LogLevelStrings = [...]string{
	"DEBUG",
	"INFO",
	"WARNING",
	"ERROR",
	"FATAL",
}

var logLevel = LogWarning
var defaultLogger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

func ResetLog(logfilePath, logLevel string) {
	if len(logfilePath) > 0 {
		if fp, err := os.Create(logfilePath); err == nil {
			defaultLogger = log.New(fp, "", log.Ldate|log.Ltime|log.Lshortfile)
		} else {
			Logf(LogError, "Unable to open logfile %s.", logfilePath)
		}
	}
	if len(logLevel) > 0 {
		SetLogLevel(logLevel)
	}
}

func Logf(level int, format string, args ...interface{}) {
	if level < logLevel {
		return
	}
	msg := fmt.Sprintf(format, args...)
	msg = fmt.Sprintf("[%s] %s", LogLevelStrings[level], msg)
	defaultLogger.Output(2, msg)
}

func SetLogLevel(level string) {
	level = strings.ToLower(level)
	for i, ll := range LogLevelStrings {
		if strings.ToLower(ll) == level {
			logLevel = i
			return
		}
	}
}
