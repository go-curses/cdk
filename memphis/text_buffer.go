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
	"sync"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/math"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
	"github.com/go-curses/cdk/log"
)

var (
	TapSpace = "    "
)

type TextBuffer interface {
	Clone() TextBuffer
	Set(input string, style paint.Style)
	Select(start, end int)
	Input() (raw string)
	SetInput(input WordLine)
	Style() paint.Style
	SetStyle(style paint.Style)
	Mnemonic() (enabled bool)
	SetMnemonic(enabled bool)
	CharacterCount() (cellCount int)
	WordCount() (wordCount int)
	LineCount() (lineCount int)
	ClearText(wordWrap enums.WrapMode, ellipsize bool, justify enums.Justification, maxChars int) (plain string)
	PlainText(wordWrap enums.WrapMode, ellipsize bool, justify enums.Justification, maxChars int) (plain string)
	PlainTextInfo(wordWrap enums.WrapMode, ellipsize bool, justify enums.Justification, maxChars int) (longestLine, lineCount int)
	Draw(canvas Surface, singleLine bool, wordWrap enums.WrapMode, ellipsize bool, justify enums.Justification, vAlign enums.VerticalAlignment) enums.EventFlag
}

type CTextBuffer struct {
	raw       string
	input     WordLine
	style     paint.Style
	mnemonics bool
	selection *ptypes.Range

	sync.Mutex
}

func NewEmptyTextBuffer(style paint.Style, mnemonic bool) TextBuffer {
	return &CTextBuffer{
		style:     style,
		mnemonics: mnemonic,
	}
}

func NewTextBuffer(input string, style paint.Style, mnemonic bool) TextBuffer {
	tb := &CTextBuffer{
		style:     style,
		mnemonics: mnemonic,
	}
	tb.Set(input, style)
	return tb
}

func (b *CTextBuffer) Clone() (cloned TextBuffer) {
	b.Lock()
	defer b.Unlock()
	cloned = NewTextBuffer(b.raw, b.style, b.mnemonics)
	return
}

func (b *CTextBuffer) Set(input string, style paint.Style) {
	b.Lock()
	b.raw = input
	b.input = NewWordLine(input, style)
	b.selection = nil
	b.Unlock()
}

func (b *CTextBuffer) Select(start, end int) {
	b.Lock()
	defer b.Unlock()
	b.selection = ptypes.NewRange(start, end)
	for i := start; i <= end; i++ {
		if style, ok := b.input.GetCharacterStyle(i); ok {
			style = style.Reverse(true)
			b.input.SetCharacterStyle(i, style)
		}
	}
}

func (b *CTextBuffer) Input() (raw string) {
	b.Lock()
	defer b.Unlock()
	return b.raw
}

func (b *CTextBuffer) SetInput(input WordLine) {
	b.Lock()
	b.input = input
	b.raw = input.Value()
	b.Unlock()
}

func (b *CTextBuffer) Style() paint.Style {
	b.Lock()
	defer b.Unlock()
	return b.style
}

func (b *CTextBuffer) SetStyle(style paint.Style) {
	b.Lock()
	b.style = style
	if b.input != nil {
		b.input = NewWordLine(b.raw, style)
	}
	b.Unlock()
}

func (b *CTextBuffer) Mnemonic() (enabled bool) {
	b.Lock()
	defer b.Unlock()
	return b.mnemonics
}

func (b *CTextBuffer) SetMnemonic(enabled bool) {
	b.Lock()
	b.mnemonics = enabled
	b.Unlock()
}

func (b *CTextBuffer) CharacterCount() (cellCount int) {
	b.Lock()
	defer b.Unlock()
	if b.input != nil {
		cellCount = b.input.CharacterCount()
	}
	return
}

func (b *CTextBuffer) WordCount() (wordCount int) {
	b.Lock()
	defer b.Unlock()
	if b.input != nil {
		wordCount = b.input.WordCount()
	}
	return
}

func (b *CTextBuffer) LineCount() (lineCount int) {
	b.Lock()
	defer b.Unlock()
	if b.input != nil {
		lineCount = b.input.LineCount()
	}
	return
}

func (b *CTextBuffer) ClearText(wordWrap enums.WrapMode, ellipsize bool, justify enums.Justification, maxChars int) (plain string) {
	if b.input == nil {
		return
	}
	lines := b.input.Make(false, wordWrap, ellipsize, justify, maxChars, b.style)
	for _, line := range lines {
		if len(plain) > 0 {
			plain += "\n"
		}
		for _, word := range line.Words() {
			for _, char := range word.Characters() {
				plain += char.StringValue()
			}
		}
	}
	return
}

func (b *CTextBuffer) PlainText(wordWrap enums.WrapMode, ellipsize bool, justify enums.Justification, maxChars int) (plain string) {
	if b.input == nil {
		return
	}
	lines := b.input.Make(b.mnemonics, wordWrap, ellipsize, justify, maxChars, b.style)
	for _, line := range lines {
		if len(plain) > 0 {
			plain += "\n"
		}
		for _, word := range line.Words() {
			for _, char := range word.Characters() {
				plain += char.StringValue()
			}
		}
	}
	return
}

func (b *CTextBuffer) PlainTextInfo(wordWrap enums.WrapMode, ellipsize bool, justify enums.Justification, maxChars int) (longestLine, lineCount int) {
	if b.input == nil {
		return
	}
	lines := b.input.Make(b.mnemonics, wordWrap, ellipsize, justify, maxChars, b.style)
	lineCount = len(lines)
	for _, line := range lines {
		lcc := line.CharacterCount()
		if longestLine < lcc {
			longestLine = lcc
		}
	}
	return
}

func (b *CTextBuffer) Draw(canvas Surface, singleLine bool, wordWrap enums.WrapMode, ellipsize bool, justify enums.Justification, vAlign enums.VerticalAlignment) enums.EventFlag {
	if b.input == nil {
		return enums.EVENT_PASS
	}
	b.Lock()
	defer b.Unlock()
	characterCount := b.input.CharacterCount()
	if b.input == nil || characterCount == 0 {
		log.TraceDF(1, "text buffer input nil or zero length")
		return enums.EVENT_PASS
	}

	if singleLine {
		wordWrap = enums.WRAP_NONE
	}

	maxChars := canvas.Width()
	lines := b.input.Make(b.mnemonics, wordWrap, ellipsize, justify, maxChars, b.style)
	size := canvas.GetSize()
	if size.W <= 0 || size.H <= 0 {
		log.TraceDF(1, "text buffer zero canvas size")
		return enums.EVENT_PASS
	}
	if len(lines) == 0 {
		log.TraceDF(1, "text buffer zero lines")
		return enums.EVENT_PASS
	}
	lenLines := len(lines)
	if singleLine && lenLines > 1 {
		lenLines = 1
	}

	var atCanvasLine, fromInputLine = 0, 0
	switch vAlign {
	case enums.ALIGN_BOTTOM:
		numLines := lenLines
		if numLines > size.H {
			delta := math.FloorI(numLines-size.H, 0)
			fromInputLine = delta
		} else {
			delta := size.H - numLines
			atCanvasLine = delta
		}
	case enums.ALIGN_MIDDLE:
		numLines := lenLines
		halfLines := numLines / 2
		halfCanvas := size.H / 2
		delta := math.FloorI(halfCanvas-halfLines, 0)
		if numLines > size.H {
			fromInputLine = delta
		} else {
			atCanvasLine = delta
		}
	case enums.ALIGN_TOP:
	default:
	}

	count := 0
	y := atCanvasLine
	for lid := fromInputLine; lid < lenLines; lid++ {
		if lid >= lenLines {
			break
		}
		if y >= size.H {
			break
		}
		x := 0
		words := lines[lid].Words()
		for _, word := range words {
			characters := word.Characters()
			for _, c := range characters {
				if x <= size.W {
					_ = canvas.SetRune(x, y, c.Value(), c.Style())
					x++
					count++
				}
			}
		}
		y++

		if b.selection != nil {
			if count < characterCount && y+lid <= lenLines {
				if b.selection.InRange(count) {
					ey := y - 1
					for ex := x; ex < size.W; ex++ {
						if c := canvas.GetContent(ex, ey); c != nil {
							_ = canvas.SetRune(ex, ey, c.Value(), c.Style().Reverse(true))
						}
					}
				}
			}
		}

		count++
		if singleLine {
			break
		}
	}

	return enums.EVENT_STOP
}
