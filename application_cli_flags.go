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

package cdk

import (
	"github.com/go-curses/cdk/log"
	"github.com/urfave/cli/v2"
)

var (
	AppCliProfileFlag = &cli.StringFlag{
		Name:        "cdk-profile",
		EnvVars:     []string{"GO_CDK_PROFILE"},
		Value:       "",
		Usage:       "profile one of: none, block, cpu, goroutine, mem, mutex, thread or trace",
		DefaultText: "none",
	}
	AppCliProfilePathFlag = &cli.StringFlag{
		Name:        "cdk-profile-path",
		EnvVars:     []string{"GO_CDK_PROFILE_PATH"},
		Value:       "",
		Usage:       "specify the directory path to store the profile data",
		DefaultText: DefaultGoProfilePath,
	}
	AppCliLogFileFlag = &cli.StringFlag{
		Name:        "cdk-log-file",
		EnvVars:     []string{"GO_CDK_LOG_FILE"},
		Value:       "",
		Usage:       "path to log file",
		DefaultText: log.DefaultLogPath,
	}
	AppCliLogLevel = &cli.StringFlag{
		Name:        "cdk-log-level",
		EnvVars:     []string{"GO_CDK_LOG_LEVEL"},
		Value:       "error",
		Usage:       "highest level of verbosity",
		DefaultText: "error",
	}
	AppCliLogFormatFlag = &cli.StringFlag{
		Name:        "cdk-log-format",
		EnvVars:     []string{"GO_CDK_LOG_FORMAT"},
		Value:       "pretty",
		Usage:       "json, text or pretty",
		DefaultText: "pretty",
	}
	AppCliLogTimestampsFlag = &cli.BoolFlag{
		Name:        "cdk-log-timestamps",
		EnvVars:     []string{"GO_CDK_LOG_TIMESTAMPS"},
		Value:       false,
		Usage:       "enable timestamps",
		DefaultText: "false",
	}
	AppCliLogTimestampFormatFlag = &cli.StringFlag{
		Name:        "cdk-log-timestamp-format",
		EnvVars:     []string{"GO_CDK_LOG_TIMESTAMP_FORMAT"},
		Value:       log.DefaultTimestampFormat,
		Usage:       "timestamp format",
		DefaultText: log.DefaultTimestampFormat,
	}
	AppCliLogFullPathsFlag = &cli.BoolFlag{
		Name:        "cdk-log-full-paths",
		EnvVars:     []string{"GO_CDK_LOG_FULL_PATHS"},
		Value:       false,
		Usage:       "log the full paths of source files",
		DefaultText: "false",
	}
	AppCliLogOutputFlag = &cli.StringFlag{
		Name:        "cdk-log-output",
		EnvVars:     []string{"GO_CDK_LOG_OUTPUT"},
		Value:       "file",
		Usage:       "logging output type: stdout, stderr or file",
		DefaultText: "file",
	}
	AppCliLogLevelsFlag = &cli.BoolFlag{
		Name:  "cdk-log-levels",
		Value: false,
		Usage: "list the levels of logging verbosity",
	}
)

func GetApplicationCliFlags() (flags []cli.Flag) {
	if Build.Profiling {
		flags = append(flags, AppCliProfileFlag, AppCliProfilePathFlag)
	}
	if Build.LogFile {
		flags = append(flags, AppCliLogFileFlag)
	}
	if Build.LogFormat {
		flags = append(flags, AppCliLogFormatFlag)
	}
	if Build.LogFullPaths {
		flags = append(flags, AppCliLogFullPathsFlag)
	}
	if Build.LogLevel {
		flags = append(flags, AppCliLogLevel)
	}
	if Build.LogLevels {
		flags = append(flags, AppCliLogLevelsFlag)
	}
	if Build.LogTimestampFormat {
		flags = append(flags, AppCliLogTimestampFormatFlag)
	}
	if Build.LogTimestamps {
		flags = append(flags, AppCliLogTimestampsFlag)
	}
	if Build.LogOutput {
		flags = append(flags, AppCliLogOutputFlag)
	}
	return
}
