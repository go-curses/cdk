// Copyright (c) 2022-2023  The Go-Curses Authors
// Copyright 2019 The TCell Authors
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
	"sync"

	"github.com/go-curses/cdk/lib/paint"
)

type cell struct {
	currMain  rune
	currComb  []rune
	currStyle paint.Style
	lastMain  rune
	lastStyle paint.Style
	lastComb  []rune
	width     int
	valid     bool

	sync.Mutex
}

func newCell() *cell {
	c := &cell{}
	_ = c.init()
	return c
}

func (c *cell) init() bool {
	if !c.valid {
		c.currComb = make([]rune, 0)
		c.lastComb = make([]rune, 0)
		c.valid = true
		return true
	}
	return false
}
