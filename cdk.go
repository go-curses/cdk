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
)

var (
	cdkHasInitialized = false
)

func Init() {
	if cdkHasInitialized {
		log.DebugF("called more than once")
		return
	}
	cdkHasInitialized = true
	_ = log.StartRestart()
	log.DebugF("CDK has been initialized")
}
