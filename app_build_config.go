// Copyright 2021  The CDK Authors
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
	"github.com/urfave/cli/v2"
)

type Config struct {
	Profiling          bool
	LogFile            bool
	LogFormat          bool
	LogFullPaths       bool
	LogLevel           bool
	LogLevels          bool
	LogTimestamps      bool
	LogTimestampFormat bool
	LogOutput          bool
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
}

func getCdkCliFlags() (flags []cli.Flag) {
	if Build.Profiling {
		flags = append(flags, cdkProfileFlag, cdkProfilePathFlag)
	}
	if Build.LogFile {
		flags = append(flags, cdkLogFileFlag)
	}
	if Build.LogFormat {
		flags = append(flags, cdkLogFormatFlag)
	}
	if Build.LogFullPaths {
		flags = append(flags, cdkLogFullPathsFlag)
	}
	if Build.LogLevel {
		flags = append(flags, cdkLogLevel)
	}
	if Build.LogLevels {
		flags = append(flags, cdkLogLevelsFlag)
	}
	if Build.LogTimestampFormat {
		flags = append(flags, cdkLogTimestampFormatFlag)
	}
	if Build.LogTimestamps {
		flags = append(flags, cdkLogTimestampsFlag)
	}
	if Build.LogOutput {
		flags = append(flags, cdkLogOutputFlag)
	}
	return
}
