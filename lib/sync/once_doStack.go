// Copyright (c) 2023  The Go-Curses Authors
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

//go:build doStack
// +build doStack

package sync

import (
	"fmt"
	"runtime"
)

func (o *Once) makeTag(depth int) (tag string) {
	depth += 1
	if pc, _, line, ok := runtime.Caller(depth); ok {
		details := runtime.FuncForPC(pc)
		tag = fmt.Sprintf("%v:%d", details.Name(), line)
	} else {
		tag = fmt.Sprintf("invalid depth: %d", depth)
	}
	return
}

func (o *Once) Do(fn func()) {
	o.doStack = o.makeTag(1)
	o.Once.Do(fn)
}
