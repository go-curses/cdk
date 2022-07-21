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

import (
	ejson "encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/go-curses/cdk/env"
	cpaths "github.com/go-curses/cdk/lib/paths"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func TestLoggingInit(t *testing.T) {
	Convey("Logging initialization checks", t, func() {
		// check for system event
		// test output methods: stdout, stderr, file, filepath, /dev/null
		logged, _, err := DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			_ = StartRestart()
			ErrorF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(logger.Formatter, ShouldHaveSameTypeAs, &prefixed.TextFormatter{})
		So(logged, ShouldStartWith, "ERROR")
		So(logged, ShouldEndWith, "testing\n")
		logged, _, err = DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stderr")
			_ = StartRestart()
			ErrorF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(logger.Formatter, ShouldHaveSameTypeAs, &prefixed.TextFormatter{})
		So(logged, ShouldStartWith, "ERROR")
		So(logged, ShouldEndWith, "testing\n")
	})
}

func TestLoggingTimestamps(t *testing.T) {
	Convey("Logging timestamp checks", t, func() {
		logged, _, err := DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			env.Set("GO_CDK_LOG_TIMESTAMPS", "true")
			env.Set("GO_CDK_LOG_TIMESTAMP_FORMAT", "2006-01-02")
			_ = StartRestart()
			ErrorF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		datestamp := time.Now().Format("2006-01-02")
		So(logged, ShouldStartWith, "["+datestamp+"]")
		So(logged, ShouldEndWith, "testing\n")
		env.Set("GO_CDK_LOG_TIMESTAMPS", "")
		env.Set("GO_CDK_LOG_TIMESTAMP_FORMAT", "")
	})
}

func TestLoggingFormatter(t *testing.T) {
	Convey("Logging json formatter checks", t, func() {
		// test formatter settings
		logged, _, err := DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			env.Set("GO_CDK_LOG_FORMAT", "json")
			_ = StartRestart()
			ErrorF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(logger.Formatter, ShouldHaveSameTypeAs, &log.JSONFormatter{})
		decoded := make(map[string]interface{})
		err = ejson.Unmarshal([]byte(logged), &decoded)
		So(err, ShouldBeNil)
		So(decoded, ShouldNotBeEmpty)
		So(decoded["level"], ShouldHaveSameTypeAs, "")
		So(decoded["level"].(string), ShouldEqual, "error")
		So(decoded["msg"], ShouldHaveSameTypeAs, "")
		So(decoded["msg"].(string), ShouldEndWith, "testing")
	})
	Convey("Logging text formatter checks", t, func() {
		logged, _, err := DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			env.Set("GO_CDK_LOG_FORMAT", "text")
			_ = StartRestart()
			ErrorF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(logger.Formatter, ShouldHaveSameTypeAs, &log.TextFormatter{})
		So(logged, ShouldStartWith, "level=error")
		So(logged, ShouldEndWith, "testing\"\n")
	})
}

func TestLoggingLevel(t *testing.T) {
	Convey("Logging level checks", t, func() {
		// trace
		logged, _, err := DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			env.Set("GO_CDK_LOG_FORMAT", "pretty")
			env.Set("GO_CDK_LOG_LEVEL", "trace")
			_ = StartRestart()
			TraceF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(logged, ShouldStartWith, "TRACE")
		So(logged, ShouldEndWith, "testing\n")
		// debug
		logged, _, err = DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			env.Set("GO_CDK_LOG_FORMAT", "pretty")
			env.Set("GO_CDK_LOG_LEVEL", "debug")
			_ = StartRestart()
			TraceF("testing")
			DebugF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(logged, ShouldStartWith, "DEBUG")
		So(logged, ShouldEndWith, "testing\n")
		// trace, debug, info
		logged, _, err = DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			env.Set("GO_CDK_LOG_FORMAT", "pretty")
			env.Set("GO_CDK_LOG_LEVEL", "info")
			_ = StartRestart()
			TraceF("testing")
			DebugF("testing")
			InfoF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(logged, ShouldStartWith, " INFO")
		So(logged, ShouldEndWith, "testing\n")
		// trace, debug, info, warn
		logged, _, err = DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			env.Set("GO_CDK_LOG_FORMAT", "pretty")
			env.Set("GO_CDK_LOG_LEVEL", "warn")
			_ = StartRestart()
			TraceF("testing")
			DebugF("testing")
			InfoF("testing")
			WarnF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(logged, ShouldStartWith, " WARN")
		So(logged, ShouldEndWith, "testing\n")
		// trace, debug, info, warn, error
		logged, _, err = DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			env.Set("GO_CDK_LOG_FORMAT", "pretty")
			env.Set("GO_CDK_LOG_LEVEL", "error")
			_ = StartRestart()
			TraceF("testing")
			DebugF("testing")
			InfoF("testing")
			WarnF("testing")
			ErrorF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(logged, ShouldStartWith, "ERROR")
		So(logged, ShouldEndWith, "testing\n")
		// fatal
		var fatal bool = false
		logged, _, err = DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			env.Set("GO_CDK_LOG_FORMAT", "pretty")
			env.Set("GO_CDK_LOG_LEVEL", "error")
			_ = StartRestart()
			logger.ExitFunc = func(int) { fatal = true }
			FatalF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(fatal, ShouldEqual, true)
		So(logged, ShouldStartWith, "FATAL")
		So(logged, ShouldEndWith, "testing\n")
		logger.ExitFunc = nil
		// panic
		var panicking bool = false
		logged, _, err = DoWithFakeIO(func() error {
			env.Set("GO_CDK_LOG_OUTPUT", "stdout")
			env.Set("GO_CDK_LOG_FORMAT", "pretty")
			env.Set("GO_CDK_LOG_LEVEL", "error")
			_ = StartRestart()
			defer func() {
				recover()
				panicking = true
			}()
			PanicF("testing")
			return nil
		})
		So(err, ShouldBeNil)
		So(panicking, ShouldEqual, true)
		So(logged, ShouldStartWith, "ERROR")
		So(logged, ShouldEndWith, "testing\n")
		// log prefix testing
		prefix := getLogPrefix(99)
		So(prefix, ShouldEqual, "(missing caller metadata)")
	})
}

func TestLoggingToFiles(t *testing.T) {
	Convey("Logging file checks", t, func() {
		So(logFileHandle, ShouldBeNil)
		if _, err := os.Stat(DefaultLogPath); err == nil {
			_ = os.Remove(DefaultLogPath)
		}
		env.Set("GO_CDK_LOG_OUTPUT", "file")
		env.Set("GO_CDK_LOG_FORMAT", "pretty")
		env.Set("GO_CDK_LOG_LEVEL", "error")
		env.Set("GO_CDK_LOG_FILE", DefaultLogPath)
		_ = StartRestart()
		So(logFileHandle, ShouldNotBeNil)
		ErrorF("testing")
		found_file := false
		if _, err := os.Stat(DefaultLogPath); err == nil {
			found_file = true
		}
		So(found_file, ShouldEqual, true)
		logged, err := ioutil.ReadFile(DefaultLogPath)
		So(err, ShouldBeNil)
		So(string(logged), ShouldEndWith, "testing\n")
		So(logFileHandle.Close(), ShouldBeNil)
		So(os.Remove(DefaultLogPath), ShouldBeNil)
		env.Set("GO_CDK_LOG_FILE", "/dev/null")
		err = StartRestart()
		So(err, ShouldBeNil)
		ErrorF("testing")
		found_file = false
		if _, err := os.Stat(DefaultLogPath); err == nil {
			found_file = true
		}
		So(found_file, ShouldEqual, false)
		tmpLog := os.TempDir() + string(os.PathSeparator) + "cdk.not.log"
		if cpaths.IsFile(tmpLog) {
			So(os.Remove(tmpLog), ShouldBeNil)
		}
		env.Set("GO_CDK_LOG_FILE", tmpLog)
		_ = StartRestart()
		ErrorF("testing")
		found_file = false
		if _, err := os.Stat(tmpLog); err == nil {
			found_file = true
		}
		So(found_file, ShouldEqual, true)
		logged, err = ioutil.ReadFile(tmpLog)
		So(err, ShouldBeNil)
		So(string(logged), ShouldEndWith, "testing\n")
		So(logFileHandle, ShouldNotBeNil)
		So(Stop(), ShouldBeNil)
		So(logFileHandle, ShouldBeNil)
		So(os.Chmod(tmpLog, 0000), ShouldBeNil)
		err = StartRestart()
		So(err, ShouldNotBeNil)
		So(os.Chmod(tmpLog, 0660), ShouldBeNil)
		So(os.Remove(tmpLog), ShouldBeNil)
		// restore default logging
		env.Set("GO_CDK_LOG_FILE", DefaultLogPath)
		_ = StartRestart()
	})
}
