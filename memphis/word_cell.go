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

package memphis

import (
	"github.com/go-curses/cdk/lib/paint"
)

// WordCell holds a list of characters making up a word or a gap (space)

type WordCell interface {
	Characters() []TextCell
	Set(word string, style paint.Style)
	GetCharacter(index int) (char TextCell)
	AppendRune(r rune, style paint.Style)
	IsNil() bool
	IsSpace() bool
	HasSpace() bool
	NewlineCount() (newlineCount int)
	Len() (count int)
	CompactLen() (count int)
	Value() (word string)
	String() (s string)
}

type CWordCell struct {
	characters []TextCell
}

func NewEmptyWordCell() WordCell {
	return &CWordCell{
		characters: make([]TextCell, 0),
	}
}

func NewNilWordCell(style paint.Style) WordCell {
	return &CWordCell{
		characters: []TextCell{NewTextCellFromRune(rune(0), style)},
	}
}

func NewWordCell(word string, style paint.Style) WordCell {
	w := &CWordCell{}
	w.Set(word, style)
	return w
}

func (w *CWordCell) Characters() []TextCell {
	return w.characters
}

func (w *CWordCell) Set(word string, style paint.Style) {
	w.characters = make([]TextCell, len(word))
	for i, c := range word {
		w.characters[i] = NewTextCellFromRune(c, style)
	}
	return
}

func (w *CWordCell) GetCharacter(index int) (char TextCell) {
	if index < len(w.characters) {
		char = w.characters[index]
	}
	return
}

func (w *CWordCell) AppendRune(r rune, style paint.Style) {
	w.characters = append(
		w.characters,
		NewTextCellFromRune(r, style),
	)
}

func (w *CWordCell) IsNil() bool {
	for _, c := range w.characters {
		if !c.IsNil() {
			return false
		}
	}
	return true
}

func (w *CWordCell) IsSpace() bool {
	for _, c := range w.characters {
		if !c.IsSpace() {
			return false
		}
	}
	return true
}

func (w *CWordCell) HasSpace() bool {
	for _, c := range w.characters {
		if c.IsSpace() {
			return true
		}
	}
	return false
}

func (w *CWordCell) NewlineCount() (newlineCount int) {
	for _, c := range w.characters {
		if c.IsNewline() {
			newlineCount += 1
		}
	}
	return
}

// the total number of characters in this word
func (w *CWordCell) Len() (count int) {
	count = 0
	for _, c := range w.characters {
		count += c.Count()
	}
	return
}

// same as `Len()` with space-words being treated as 1 character wide rather
// than the literal number of spaces from the input string
func (w *CWordCell) CompactLen() (count int) {
	if w.IsSpace() {
		count = 1
		return
	}
	count = w.Len()
	return
}

// returns the literal string value of the word
func (w *CWordCell) Value() (word string) {
	word = ""
	for _, c := range w.characters {
		word += c.StringValue()
	}
	return
}

// returns the debuggable value of the word
func (w *CWordCell) String() (s string) {
	s = ""
	for _, c := range w.characters {
		s += c.String()
	}
	return
}