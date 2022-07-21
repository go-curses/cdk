// Copyright 2022  The CDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

// TODO: Panic and Fatal logging methods need to release the gdk display

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	"github.com/go-curses/cdk/env"
	cstrings "github.com/go-curses/cdk/lib/strings"
)

const (
	LevelError string = "error"
	LevelWarn  string = "warn"
	LevelInfo  string = "info"
	LevelDebug string = "debug"
	LevelTrace string = "trace"
)

var LogLevels = []string{
	LevelError,
	LevelWarn,
	LevelInfo,
	LevelDebug,
	LevelTrace,
}

const (
	FormatPretty string = "pretty"
	FormatText   string = "text"
	FormatJson   string = "json"
)

const (
	OutputStderr string = "stderr"
	OutputStdout string = "stdout"
	OutputFile   string = "file"
)

var (
	logger        = log.New()
	logFileHandle *os.File
	logFullPaths  = false
	logBuffer     = bytes.NewBufferString("")

	DefaultLogPath = os.TempDir() + string(os.PathSeparator) + "cdk.log"
)

const (
	StandardTimestampFormat = "2006-01-02T15:04:05.000"
	DefaultTimestampFormat  = "20060102-150405.00"
)

func GetExitFunc() func(int) {
	if logger == nil {
		return func(int) {}
	}
	return logger.ExitFunc
}

func SetExitFunc(fn func(int)) {
	logger.ExitFunc = fn
}

func StartRestart() error {
	disableTimestamp := true
	if v := env.Get("GO_CDK_LOG_TIMESTAMPS", "false"); v == "true" {
		disableTimestamp = false
	}
	timestampFormat := DefaultTimestampFormat
	if v := env.Get("GO_CDK_LOG_TIMESTAMP_FORMAT", ""); v != "" {
		if v == "standard" {
			timestampFormat = StandardTimestampFormat
		} else if v == "default" {
			timestampFormat = DefaultTimestampFormat
		} else {
			timestampFormat = v
		}
	}
	switch env.Get("GO_CDK_LOG_FULL_PATHS", "false") {
	case "true":
		logFullPaths = true
	default:
		logFullPaths = false
	}
	switch env.Get("GO_CDK_LOG_FORMAT", "pretty") {
	case FormatJson:
		logger.SetFormatter(&log.JSONFormatter{
			TimestampFormat:  timestampFormat,
			DisableTimestamp: disableTimestamp,
		})
	case FormatText:
		logger.SetFormatter(&log.TextFormatter{
			TimestampFormat:  timestampFormat,
			DisableTimestamp: disableTimestamp,
			DisableSorting:   true,
			DisableColors:    true,
			FullTimestamp:    true,
		})
	case FormatPretty:
		fallthrough
	default:
		logger.SetFormatter(&prefixed.TextFormatter{
			DisableTimestamp: disableTimestamp,
			TimestampFormat:  timestampFormat,
			ForceFormatting:  true,
			FullTimestamp:    true,
			DisableSorting:   true,
			DisableColors:    true,
		})
	}
	switch env.Get("GO_CDK_LOG_LEVEL", LevelError) {
	case LevelTrace:
		logger.SetLevel(log.TraceLevel)
	case LevelDebug:
		logger.SetLevel(log.DebugLevel)
	case LevelInfo:
		logger.SetLevel(log.InfoLevel)
	case LevelWarn:
		logger.SetLevel(log.WarnLevel)
	case LevelError:
		fallthrough
	default:
		logger.SetLevel(log.ErrorLevel)
	}
	switch env.Get("GO_CDK_LOG_OUTPUT", OutputFile) {
	case OutputStdout:
		logger.SetOutput(os.Stdout)
	case OutputStderr:
		logger.SetOutput(os.Stderr)
	case OutputFile:
		fallthrough
	default:
		_ = Stop()
		if logfile := env.Get("GO_CDK_LOG_FILE", DefaultLogPath); !cstrings.IsEmpty(logfile) && logfile != "/dev/null" {
			logFH, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
			logFileHandle = logFH
			_, _ = logFileHandle.WriteString(logBuffer.String())
			logBuffer.Reset()
			logger.SetOutput(logFileHandle)
		} else {
			logger.SetOutput(ioutil.Discard)
		}
	}
	return nil
}

func Stop() error {
	if logFileHandle != nil {
		logBuffer.Reset()
		logger.SetOutput(logBuffer)
		if err := logFileHandle.Close(); err != nil {
			return err
		}
		logFileHandle = nil
	}
	return nil
}

func GetLevel() string {
	switch logger.GetLevel() {
	case log.TraceLevel:
		return LevelTrace
	case log.DebugLevel:
		return LevelDebug
	case log.WarnLevel:
		return LevelWarn
	case log.InfoLevel:
		return LevelInfo
	default:
		return LevelError
	}
}

func getLogPrefix(depth int) string {
	depth += 1
	if function, file, line, ok := runtime.Caller(depth); ok {
		fullName := runtime.FuncForPC(function).Name()
		funcName := fullName
		if i := strings.LastIndex(fullName, "."); i > -1 {
			funcName = fullName[i+1:]
		}
		packName := fullName
		if i := strings.Index(fullName, "."); i > -1 {
			packName = fullName[:i+1]
		}
		filepath := file
		if !logFullPaths {
			if packName == "main." {
				filepath = path.Base(filepath)
			} else {
				if i := strings.Index(filepath, packName); i > -1 {
					filepath = file[i:]
				}
			}
		}
		return fmt.Sprintf("%s:%d	%s", filepath, line, funcName)
	}
	return "(missing caller metadata)"
}

func TraceF(format string, argv ...interface{}) { TraceDF(1, format, argv...) }
func TraceDF(depth int, format string, argv ...interface{}) {
	logger.Tracef(cstrings.NLSprintf("%s	%s", getLogPrefix(depth+1), format), argv...)
}

func DebugF(format string, argv ...interface{}) { DebugDF(1, format, argv...) }
func DebugDF(depth int, format string, argv ...interface{}) {
	logger.Debugf(cstrings.NLSprintf("%s	%s", getLogPrefix(depth+1), format), argv...)
}

func InfoF(format string, argv ...interface{}) { InfoDF(1, format, argv...) }
func InfoDF(depth int, format string, argv ...interface{}) {
	logger.Infof(cstrings.NLSprintf("%s	%s", getLogPrefix(depth+1), format), argv...)
}

func WarnF(format string, argv ...interface{}) { WarnDF(1, format, argv...) }
func WarnDF(depth int, format string, argv ...interface{}) {
	logger.Warnf(cstrings.NLSprintf("%s	%s", getLogPrefix(depth+1), format), argv...)
}

func Error(err error)                           { ErrorDF(1, err.Error()) }
func ErrorF(format string, argv ...interface{}) { ErrorDF(1, format, argv...) }
func ErrorDF(depth int, format string, argv ...interface{}) {
	logger.Errorf(cstrings.NLSprintf("%s	%s", getLogPrefix(depth+1), format), argv...)
}

func Fatal(err error)                           { FatalDF(1, err.Error()) }
func FatalF(format string, argv ...interface{}) { FatalDF(1, format, argv...) }
func FatalDF(depth int, format string, argv ...interface{}) {
	// if dm := GetDisplay(); dm != nil {
	// 	dm.ReleaseDisplay()
	// }
	message := fmt.Sprintf(cstrings.NLSprintf("%s\t%s", getLogPrefix(depth+1), format), argv...)
	logger.Fatalf(message)
}

func Panic(err error)                           { PanicDF(1, err.Error()) }
func PanicF(format string, argv ...interface{}) { PanicDF(1, format, argv...) }
func PanicDF(depth int, format string, argv ...interface{}) {
	// if dm := GetDisplay(); dm != nil {
	// 	dm.ReleaseDisplay()
	// }
	message := fmt.Sprintf(cstrings.NLSprintf("%s\t%s", getLogPrefix(depth+1), format), argv...)
	logger.Errorf(message)
	_ = Stop()
	panic(message)
}

func Exit(code int) {
	InfoDF(1, "exiting with code: %d", code)
	logger.Exit(code)
}
