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

//go:build lockStack
// +build lockStack

package sync

import (
	"fmt"
	"runtime"
	"strings"
)

func (m *Mutex) makeTag(write bool, depth int) (tag string) {
	depth += 1
	if pc, _, line, ok := runtime.Caller(depth); ok {
		details := runtime.FuncForPC(pc)
		name := details.Name()
		if strings.Contains(name, "(*CWidget).LockDraw") {
			if pc, _, line, ok = runtime.Caller(depth + 1); ok {
				details = runtime.FuncForPC(pc)
				name = details.Name()
			}
		}
		tag = fmt.Sprintf("[write:%v] %v:%d", write, name, line)
	} else {
		tag = fmt.Sprintf("invalid depth: %d", depth)
	}
	return
}

func (m *Mutex) Lock() {
	m.Mutex.Lock()
	m.lockStack = append(m.lockStack, m.makeTag(true, 1))
}

func (m *Mutex) Unlock() {
	m.Mutex.Unlock()
	if len(m.lockStack) > 1 {
		m.lockStack = append([]string{}, m.lockStack[:len(m.lockStack)-1]...)
	} else {
		m.lockStack = []string{}
	}
}
