// Copyright (c) 2022-2023  The Go-Curses Authors
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

	"github.com/go-curses/cdk/lib/paint"
)

type TextCell interface {
	Dirty() bool
	Set(r rune)
	SetByte(b []byte)
	SetStyle(style paint.Style)
	Equals(mc rune, style paint.Style, width int) bool
	Width() int
	Count() int
	Value() rune
	StringValue() string
	String() string
	Style() paint.Style
	IsNil() bool
	IsSpace() bool
	IsNewline() bool

	// sync.Locker
}

type CTextCell struct {
	char  *CTextChar
	style paint.Style
	dirty bool

	// sync.RWMutex
}

func NewTextCellFromRune(char rune, style paint.Style) *CTextCell {
	return NewTextCell(NewTextChar([]byte(string(char))), style)
}

func NewTextCell(char *CTextChar, style paint.Style) *CTextCell {
	return &CTextCell{
		char:  char,
		style: style,
		dirty: true,
	}
}

func (t *CTextCell) Equals(mc rune, style paint.Style, width int) bool {
	// t.RLock()
	// defer t.RUnlock()
	if !t.style.Equals(style) {
		return false
	}
	if t.char.Width() != width {
		return false
	}
	if t.char.Value() != mc {
		return false
	}
	return true
}

func (t *CTextCell) Dirty() bool {
	// t.RLock()
	// defer t.RUnlock()
	return t.dirty
}

func (t *CTextCell) Set(r rune) {
	// t.Lock()
	// defer t.Unlock()
	t.char.Set(r)
	t.dirty = true
}

func (t *CTextCell) SetByte(b []byte) {
	// t.Lock()
	// defer t.Unlock()
	t.char.SetByte(b)
	t.dirty = true
}

func (t *CTextCell) SetStyle(style paint.Style) {
	// t.Lock()
	// defer t.Unlock()
	t.style = style
	t.dirty = true
}

func (t *CTextCell) Width() int {
	// t.RLock()
	// defer t.RUnlock()
	return t.char.Width()
}

func (t *CTextCell) Count() int {
	// t.RLock()
	// defer t.RUnlock()
	return t.char.Count()
}

func (t *CTextCell) Value() rune {
	// t.RLock()
	// defer t.RUnlock()
	return t.char.Value()
}

func (t *CTextCell) StringValue() string {
	// t.RLock()
	// defer t.RUnlock()
	return t.char.String()
}

func (t *CTextCell) String() string {
	// t.RLock()
	// defer t.RUnlock()
	return fmt.Sprintf(
		"{Char=%s,Style=%s}",
		t.char.String(),
		t.style.String(),
	)
}

func (t *CTextCell) Style() paint.Style {
	// t.RLock()
	// defer t.RUnlock()
	return t.style
}

func (t *CTextCell) IsNil() bool {
	// t.RLock()
	// defer t.RUnlock()
	return t.char.Value() == rune(0)
}

func (t *CTextCell) IsSpace() bool {
	// t.RLock()
	// defer t.RUnlock()
	return t.char.IsSpace()
}

func (t *CTextCell) IsNewline() bool {
	// t.RLock()
	// defer t.RUnlock()
	return t.char.IsNewline()
}
