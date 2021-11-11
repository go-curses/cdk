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

package env

// TODO: implement rc key=val files?
// TODO: decide upon auto-init versus manual

import (
	"os"
	"strings"
)

var (
	cache map[string]string
)

func init() {
	Reload()
}

func Reload() {
	cache = make(map[string]string)
	environ := os.Environ()
	for _, line := range environ {
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			cache[parts[0]] = parts[1]
		} else {
			// logging warning? os.Environ() gave bad output?
		}
	}
}

func Get(key, def string) string {
	if value, exists := cache[key]; exists {
		return value
	}
	return def
}

func Set(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		// logging error
	} else {
		cache[key] = value
	}
}
