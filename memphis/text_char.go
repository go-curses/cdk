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

package memphis

import (
	"fmt"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

type TextChar interface {
	Set(r rune)
	SetByte(b []byte)
	Width() int
	Count() int
	Value() rune
	String() string
	IsSpace() bool
	IsNewline() bool
}

type CTextChar struct {
	value rune
	width int
	count int

	sync.RWMutex
}

func NewTextChar(b []byte) *CTextChar {
	t := &CTextChar{}
	t.SetByte(b)
	return t
}

func (c *CTextChar) Set(r rune) {
	c.SetByte([]byte(string(r)))
}

func (c *CTextChar) SetByte(b []byte) {
	c.value, c.width = utf8.DecodeRune(b)
	c.count = runewidth.RuneWidth(rune(string(b)[0]))
}

func (c *CTextChar) Width() int {
	return c.width
}

func (c *CTextChar) Count() int {
	return c.count
}

func (c *CTextChar) Value() rune {
	return c.value
}

func (c *CTextChar) String() string {
	if c.value == 0 {
		return " "
	}
	return fmt.Sprintf("%c", c.value)
}

func (c *CTextChar) IsSpace() bool {
	v := c.Value()
	return v == 0 || unicode.IsSpace(v)
}

func (c *CTextChar) IsNewline() bool {
	v := c.Value()
	return v == 10 || v == 13
}
