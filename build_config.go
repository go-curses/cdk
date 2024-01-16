// Copyright (c) 2022-2023  The Go-Curses Authors
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

package cdk

import (
	cstrings "github.com/go-curses/cdk/lib/strings"
)

// Setting these globals will enable command line flags and their
// corresponding features. To set these, use the go build -ldflags:
//
//	go build -v -ldflags="-X 'github.com/go-curses/cdk.IncludeTtyFlag=true'"
var (
	IncludeTtyFlag            = "false"
	IncludeProfiling          = "false"
	IncludeLogFile            = "false"
	IncludeLogFormat          = "false"
	IncludeLogFullPaths       = "false"
	IncludeLogLevel           = "false"
	IncludeLogLevels          = "false"
	IncludeLogTimestamps      = "false"
	IncludeLogTimestampFormat = "false"
	IncludeLogOutput          = "false"
)

type Config struct {
	TtyFlag            bool
	Profiling          bool
	LogFile            bool
	LogFormat          bool
	LogFullPaths       bool
	LogLevel           bool
	LogLevels          bool
	LogTimestamps      bool
	LogTimestampFormat bool
	LogOutput          bool
	DisableLocalCall   bool
	DisableRemoteCall  bool
}

var Build = Config{
	Profiling:          false,
	LogFile:            false,
	LogFormat:          false,
	LogFullPaths:       false,
	LogLevel:           false,
	LogLevels:          false,
	LogTimestamps:      false,
	LogTimestampFormat: false,
	LogOutput:          false,
	DisableLocalCall:   true,
	DisableRemoteCall:  true,
}

func init() {
	Build.TtyFlag = cstrings.IsTrue(IncludeTtyFlag)
	Build.Profiling = cstrings.IsTrue(IncludeProfiling)
	Build.LogFile = cstrings.IsTrue(IncludeLogFile)
	Build.LogFormat = cstrings.IsTrue(IncludeLogFormat)
	Build.LogFullPaths = cstrings.IsTrue(IncludeLogFullPaths)
	Build.LogLevel = cstrings.IsTrue(IncludeLogLevel)
	Build.LogLevels = cstrings.IsTrue(IncludeLogLevels)
	Build.LogTimestamps = cstrings.IsTrue(IncludeLogTimestamps)
	Build.LogTimestampFormat = cstrings.IsTrue(IncludeLogTimestampFormat)
	Build.LogOutput = cstrings.IsTrue(IncludeLogOutput)
}
