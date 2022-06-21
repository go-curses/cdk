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

package strings

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

func BasicWordWrap(input string, maxChars int) (output string) {
	var words []string
	single := strings.Replace(input, "\n", " ", -1)
	if words = strings.Fields(single); len(words) == 0 {
		return input
	}
	output = ""
	count := 0
	for idx, word := range words {
		if count+1+len(word) >= maxChars {
			output += "\n"
			count = 0
		} else if idx > 0 {
			output += " "
		}
		output += word
		count += 1 + len(word)
	}
	return
}

func PadLeft(src, pad string, length int) string {
	for {
		if len(src) > length {
			return src[0 : length+1]
		}
		src = pad + src
	}
}

func PadRight(src, pad string, length int) string {
	for {
		if len(src) > length {
			return src[0 : length+1]
		}
		src += pad
	}
}

func CleanCRLF(s string) string {
	length := len(s)
	var last int
	for last = length - 1; last >= 0; last-- {
		if s[last] != '\r' && s[last] != '\n' {
			break
		}
	}
	return s[:last+1]
}

func NLSprintf(format string, argv ...interface{}) string {
	return CleanCRLF(fmt.Sprintf(format, argv...))
}

var _rxIsEmpty = regexp.MustCompile(`^\s*$`)

func IsEmpty(text string) bool {
	return len(text) == 0 || _rxIsEmpty.MatchString(text)
}

func HasSpace(text string) bool {
	for _, c := range text {
		if unicode.IsSpace(c) {
			return true
		}
	}
	return false
}

func IsBoolean(text string) bool {
	switch strings.ToLower(text) {
	case "1", "on", "yes", "y", "true":
		fallthrough
	case "0", "nil", "off", "no", "n", "false":
		return true
	}
	return false
}

func IsTrue(text string) bool {
	switch strings.ToLower(text) {
	case "1", "on", "yes", "y", "true":
		return true
	}
	return false
}

func IsFalse(text string) bool {
	switch strings.ToLower(text) {
	case "0", "nil", "off", "no", "n", "false":
		return true
	}
	return false
}

func IsUrl(str string) (isUrl bool) {
	if u, err := url.Parse(str); err == nil && u.Scheme != "" {
		return true
	}
	return
}

func StringSliceHasValue(slice []string, value string) (has bool) {
	for _, str := range slice {
		if str == value {
			return true
		}
	}
	return
}

func EqualStringSlices(a, b []string) (same bool) {
	lenA := len(a)
	if lenA != len(b) {
		return false
	}
	for i := 0; i < lenA; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

var _rxStripTags = regexp.MustCompile(`<[^>]*>`)

func StripTags(input string) (output string) {
	output = _rxStripTags.ReplaceAllString(input, "")
	return
}

func MakeObjectName(tag, user, address string) (name string) {
	name = fmt.Sprintf("[%s]%s@%s", tag, user, address)
	return
}
