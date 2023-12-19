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

package memphis

import (
	"fmt"

	"github.com/go-curses/cdk/log"
)

// TODO: Surface caching!

type CWordPage []WordLine

type WordLineCacheFn = func() []WordLine

type WordPageCache interface {
	Hit(tag string, fn WordLineCacheFn) []WordLine
}

type CWordLineCache struct {
	cache map[string]CWordPage
}

func NewWordPageCache() (wpc *CWordLineCache) {
	return &CWordLineCache{
		cache: make(map[string]CWordPage),
	}
}

func (c *CWordLineCache) Clear() {
	if len(c.cache) > 0 {
		c.cache = make(map[string]CWordPage)
		log.TraceF("WordLineCache.Clear(): cache purged")
	}
}

func (c *CWordLineCache) Hit(tag string, fn WordLineCacheFn) (lines []WordLine) {
	// if v, ok := c.cache[tag]; ok {
	// 	// TraceF("WordLineCache.Hit(): returning cached value for \"%v\"", tag)
	// 	return v
	// }
	// TraceF("WordLineCache.Hit(): caching new value for \"%v\"", tag)
	// c.cache[tag] = fn()
	// return c.cache[tag]
	return fn()
}

func MakeTag(argv ...interface{}) (tag string) {
	tag += "{"
	for i, v := range argv {
		if len(tag) > 1 {
			tag += ";"
		}
		tag += fmt.Sprintf("%d=%v", i, v)
	}
	tag += "}"
	return
}
