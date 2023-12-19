//go:build !windows && !nacl && !plan9
// +build !windows,!nacl,!plan9

// Copyright (c) 2022-2023  The Go-Curses Authors
// Copyright 2016 The TCell Authors
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

package charset

import (
	"os"
	"regexp"
)

var (
	rxLocaleAtDot = regexp.MustCompile(`^.*\.(.+?)\@.*$`)
	rxLocaleAt    = regexp.MustCompile(`^(.+)@.*$`)
	rxLocaleDot   = regexp.MustCompile(`^.*\.(.+)$`)
)

func Get() string {
	// Determine the character set.  This can help us later.
	// Per POSIX, we search for LC_ALL first, then LC_CTYPE, and
	// finally LANG.  First one set wins.
	locale := ""
	if locale = os.Getenv("LC_ALL"); locale == "" {
		if locale = os.Getenv("LC_CTYPE"); locale == "" {
			locale = os.Getenv("LANG")
		}
	}
	if locale == "POSIX" || locale == "C" || locale == "US-ASCII" {
		return "US-ASCII"
	}
	if rxLocaleAtDot.MatchString(locale) {
		m := rxLocaleAtDot.FindAllStringSubmatch(locale, -1)
		return m[0][1]
	}
	if rxLocaleAt.MatchString(locale) {
		m := rxLocaleAt.FindAllStringSubmatch(locale, -1)
		return m[0][1]
	}
	if rxLocaleDot.MatchString(locale) {
		m := rxLocaleDot.FindAllStringSubmatch(locale, -1)
		return m[0][1]
	}
	// Default assumption, and on Linux we can see LC_ALL
	// without a character set, which we assume implies UTF-8.
	return "UTF-8"
}
