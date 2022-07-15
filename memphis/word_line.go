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
	"fmt"
	"unicode"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/sync"
)

type WordLine interface {
	SetLine(line string, style paint.Style)
	AppendWord(word string, style paint.Style)
	AppendWordCell(word WordCell)
	AppendWordRune(wordIndex int, char rune, style paint.Style) error
	GetWord(index int) WordCell
	RemoveWord(index int)
	GetCharacter(index int) TextCell
	SetCharacter(index int, r rune)
	GetCharacterStyle(index int) (style paint.Style, ok bool)
	SetCharacterStyle(index int, style paint.Style)
	Words() []WordCell
	Len() (wordSpaceCount int)
	CharacterCount() (count int)
	WordCount() (wordCount int)
	LineCount() (lineCount int)
	HasSpace() bool
	Value() (s string)
	String() (s string)
	Make(mnemonic bool, wrap enums.WrapMode, ellipsize bool, justify enums.Justification, maxChars int, fillerStyle paint.Style) (formatted []WordLine)
}

type CWordLine struct {
	words []WordCell
	cache *CWordLineCache

	sync.RWMutex
}

func NewEmptyWordLine() WordLine {
	return &CWordLine{
		words: make([]WordCell, 0),
		cache: NewWordPageCache(),
	}
}

func NewWordLine(line string, style paint.Style) WordLine {
	wl := &CWordLine{
		cache: NewWordPageCache(),
	}
	wl.SetLine(line, style)
	return wl
}

func (w *CWordLine) SetLine(line string, style paint.Style) {
	w.Lock()
	defer w.Unlock()
	w.cache.Clear()
	w.words = make([]WordCell, 0)
	isWord, wasNL := false, false
	wid := 0
	for _, c := range line {
		if unicode.IsSpace(c) {
			if c == '\n' {
				isWord = false
				wasNL = true
				w.words = append(w.words, NewEmptyWordCell())
				wid = len(w.words) - 1
			} else if isWord || wasNL || len(w.words) == 0 {
				isWord = false
				w.words = append(w.words, NewEmptyWordCell())
				wid = len(w.words) - 1
			}
			// appending to the "space" word
			w.words[wid].AppendRune(c, style)
		} else {
			if !isWord || len(w.words) == 0 {
				isWord = true
				w.words = append(w.words, NewEmptyWordCell())
				wid = len(w.words) - 1
			}
			w.words[wid].AppendRune(c, style)
		}
	}
}

func (w *CWordLine) AppendWord(word string, style paint.Style) {
	w.Lock()
	defer w.Unlock()
	w.cache.Clear()
	w.words = append(w.words, NewWordCell(word, style))
}

func (w *CWordLine) AppendWordCell(word WordCell) {
	w.Lock()
	defer w.Unlock()
	w.cache.Clear()
	w.words = append(w.words, word)
}

func (w *CWordLine) AppendWordRune(wordIndex int, char rune, style paint.Style) error {
	w.Lock()
	defer w.Unlock()
	if wordIndex < len(w.words) {
		w.cache.Clear()
		w.words[wordIndex].AppendRune(char, style)
		return nil
	}
	return fmt.Errorf("word at index %d not found", wordIndex)
}

func (w *CWordLine) GetWord(index int) WordCell {
	w.RLock()
	defer w.RUnlock()
	if index < len(w.words) {
		return w.words[index]
	}
	return nil
}

func (w *CWordLine) RemoveWord(index int) {
	w.Lock()
	defer w.Unlock()
	if index < len(w.words) {
		w.cache.Clear()
		w.words = append(
			w.words[:index],
			w.words[index+1:]...,
		)
	}
}

func (w *CWordLine) GetCharacter(index int) TextCell {
	if index < w.CharacterCount() {
		w.RLock()
		defer w.RUnlock()
		count := 0
		for _, word := range w.words {
			for _, c := range word.Characters() {
				if count == index {
					return c
				}
				count++
			}
		}
	}
	return nil
}

func (w *CWordLine) SetCharacter(index int, r rune) {
	if index < w.CharacterCount() {
		w.Lock()
		defer w.Unlock()
		count := 0
		for _, word := range w.words {
			for _, c := range word.Characters() {
				if count == index {
					c.Set(r)
					return
				}
				count++
			}
		}
	}
}

func (w *CWordLine) GetCharacterStyle(index int) (style paint.Style, ok bool) {
	if index < w.CharacterCount() {
		w.RLock()
		defer w.RUnlock()
		count := 0
		for _, word := range w.words {
			for _, c := range word.Characters() {
				if count == index {
					return c.Style(), true
				}
				count++
			}
		}
	}
	return paint.Style{}, false
}

func (w *CWordLine) SetCharacterStyle(index int, style paint.Style) {
	if index < w.CharacterCount() {
		w.Lock()
		defer w.Unlock()
		count := 0
		for _, word := range w.words {
			for _, c := range word.Characters() {
				if count == index {
					c.SetStyle(style)
					return
				}
				count++
			}
		}
	}
}

func (w *CWordLine) Words() []WordCell {
	w.RLock()
	defer w.RUnlock()
	return w.words
}

func (w *CWordLine) Len() (wordSpaceCount int) {
	w.RLock()
	defer w.RUnlock()
	return len(w.words)
}

func (w *CWordLine) CharacterCount() (count int) {
	w.RLock()
	defer w.RUnlock()
	for _, word := range w.words {
		count += word.Len()
	}
	return
}

func (w *CWordLine) WordCount() (wordCount int) {
	w.RLock()
	defer w.RUnlock()
	for _, word := range w.words {
		if !word.IsSpace() {
			wordCount++
		}
	}
	return
}

func (w *CWordLine) LineCount() (lineCount int) {
	w.RLock()
	defer w.RUnlock()
	for _, word := range w.words {
		lineCount += word.NewlineCount()
	}
	return
}

func (w *CWordLine) HasSpace() bool {
	w.RLock()
	defer w.RUnlock()
	for _, word := range w.words {
		if word.IsSpace() {
			return true
		}
	}
	return false
}

func (w *CWordLine) Value() (s string) {
	w.RLock()
	defer w.RUnlock()
	for _, c := range w.words {
		s += c.Value()
	}
	return
}

func (w *CWordLine) String() (s string) {
	w.RLock()
	defer w.RUnlock()
	s = "{"
	for i, c := range w.words {
		if i > 0 {
			s += ","
		}
		s += c.String()
	}
	s += "}"
	return
}

// wrap, justify and align the set input, with filler style
func (w *CWordLine) Make(mnemonic bool, wrap enums.WrapMode, ellipsize bool, justify enums.Justification, maxChars int, fillerStyle paint.Style) (formatted []WordLine) {
	tag := MakeTag(mnemonic, wrap, justify, maxChars, fillerStyle)
	return w.cache.Hit(tag, func() []WordLine {
		var lines []WordLine
		lines = append(lines, NewEmptyWordLine())
		cid, wid, lid := 0, 0, 0
		mnemonicFound := false
		w.RLock()
		defer w.RUnlock()
		for _, word := range w.words {
			for wcId, c := range word.Characters() {
				switch c.Value() {
				case '\n':
					lines = append(lines, NewEmptyWordLine())
					lid = len(lines) - 1
					wid = -1
				case '_':
					if mnemonic {
						nextId := wcId + 1
						if l := word.Len(); l > nextId {
							ch := word.GetCharacter(nextId)
							switch ch.Value() {
							case '_', ' ':
							default:
								mnemonicFound = true
								continue
							}
						}
					}
					fallthrough
				default:
					if wid >= lines[lid].Len() {
						lines[lid].AppendWordCell(NewEmptyWordCell())
					}
					if mnemonicFound {
						_ = lines[lid].AppendWordRune(wid, c.Value(), c.Style().Underline(true))
						mnemonicFound = false
					} else {
						_ = lines[lid].AppendWordRune(wid, c.Value(), c.Style())
					}
				}
				cid++
			}
			wid++
		}
		lines = w.applyTypography(wrap, ellipsize, justify, maxChars, fillerStyle, lines)
		return lines
	})
}

func (w *CWordLine) applyTypography(wrap enums.WrapMode, ellipsize bool, justify enums.Justification, maxChars int, fillerStyle paint.Style, input []WordLine) (output []WordLine) {
	output = w.applyTypographicWrap(wrap, ellipsize, maxChars, input)
	output = w.applyTypographicJustify(justify, maxChars, fillerStyle, output)
	return
}

func (w *CWordLine) applyTypographicWrap(wrap enums.WrapMode, ellipsize bool, maxChars int, input []WordLine) (output []WordLine) {
	// all space-words must be applied as 1 width
	switch wrap {
	case enums.WRAP_WORD:
		// break onto inserted/new line at end gap
		// - if line has no breakpoints, truncate
		output = w.applyTypographicWrapWord(maxChars, input)
	case enums.WRAP_WORD_CHAR:
		// break onto inserted/new line at end gap
		// - if line has no breakpoints, fallthrough
		output = w.applyTypographicWrapWordChar(maxChars, input)
	case enums.WRAP_CHAR:
		// break onto inserted/new line at maxChars
		output = w.applyTypographicWrapChar(maxChars, input)
	case enums.WRAP_NONE:
		// truncate each line to maxChars
		output = w.applyTypographicWrapNone(ellipsize, maxChars, input)
	}
	return
}

func (w *CWordLine) applyTypographicJustify(justify enums.Justification, maxChars int, fillerStyle paint.Style, input []WordLine) (output []WordLine) {
	switch justify {
	case enums.JUSTIFY_FILL:
		// each non-empty line is space-expanded to fill maxChars
		output = w.applyTypographicJustifyFill(maxChars, fillerStyle, input)
	case enums.JUSTIFY_CENTER:
		// each non-empty line is centered on halfway maxChars
		output = w.applyTypographicJustifyCenter(maxChars, fillerStyle, input)
	case enums.JUSTIFY_RIGHT:
		// each non-empty line is left-padded to fill maxChars
		output = w.applyTypographicJustifyRight(maxChars, fillerStyle, input)
	case enums.JUSTIFY_LEFT:
		// each non-empty line has leading space removed
		output = w.applyTypographicJustifyLeft(input)
	case enums.JUSTIFY_NONE:
		fallthrough
	default:
		output = w.applyTypographicJustifyNone(input)
	}
	return
}