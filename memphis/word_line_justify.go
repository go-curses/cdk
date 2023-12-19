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
	"github.com/go-curses/cdk/lib/math"
	"github.com/go-curses/cdk/lib/paint"
)

func (w *CWordLine) longestLineLen(input []WordLine) (maxChars int) {
	maxChars = 0
	for _, line := range input {
		lcc := line.CharacterCount()
		if maxChars < lcc {
			maxChars = lcc
		}
	}
	return
}

// return output lines where each line of input is full-width justified. For
// each input line, spread the words across the maxChars by increasing the sizes
// of the gaps (one or more spaces). if maxChars is -1, then the length of the
// longest line is determined and that value used in place of maxChars
func (w *CWordLine) applyTypographicJustifyFill(maxChars int, fillerStyle paint.Style, input []WordLine) (output []WordLine) {
	// trim left/right space for each line, maximize gaps
	lid := 0
	if maxChars <= -1 {
		maxChars = w.longestLineLen(input)
	}
	for _, line := range input {
		if lid >= len(output) {
			output = append(output, NewEmptyWordLine())
		}
		width := line.CharacterCount()
		gaps := make([]int, 0)
		for _, word := range line.Words() {
			if word.IsSpace() {
				gaps = append(gaps, 1)
			}
		}
		widthMinusGaps := width - len(gaps)
		gaps = math.DistInts(maxChars-widthMinusGaps, gaps)
		gid := 0
		for _, word := range line.Words() {
			if word.IsSpace() {
				wc := NewEmptyWordCell()
				for i := 0; i < gaps[gid]; i++ {
					wc.AppendRune(' ', fillerStyle)
				}
				gid++
				output[lid].AppendWordCell(wc)
			} else {
				output[lid].AppendWordCell(word)
			}
		}
		lid++
	}
	return
}

// return output lines where each line of input is centered on the full-width of
// maxChars per-line. if maxChars is -1, then the length of the
// // longest line is determined and that value used in place of maxChars
func (w *CWordLine) applyTypographicJustifyCenter(maxChars int, fillerStyle paint.Style, input []WordLine) (output []WordLine) {
	// trim left space for each line
	wid, lid := 0, 0
	if maxChars <= -1 {
		maxChars = w.longestLineLen(input)
	}
	for _, line := range input {
		if lid >= len(output) {
			output = append(output, NewEmptyWordLine())
		}
		width := line.CharacterCount()
		halfWidth := width / 2
		halfWay := maxChars / 2
		delta := halfWay - halfWidth
		if delta > 0 {
			for i := 0; i < delta; i++ {
				output[lid].AppendWordCell(NewNilWordCell(fillerStyle))
			}
		}
		for _, word := range line.Words() {
			output[lid].AppendWordCell(word)
			wid++
		}
		lid++
	}
	return
}

// return output lines where for each input line the content is left-padded with
// spaces such that the last character of content is aligned to maxChars. if
// maxChars is -1, then the length of the longest line is determined and that
// value used in place of maxChars
func (w *CWordLine) applyTypographicJustifyRight(maxChars int, fillerStyle paint.Style, input []WordLine) (output []WordLine) {
	// trim left space for each line, assume no line needs wrapping or truncation
	wid, lid := 0, 0
	if maxChars <= -1 {
		maxChars = w.longestLineLen(input)
	}
	for _, line := range input {
		if lid >= len(output) {
			output = append(output, NewEmptyWordLine())
		}
		charCount := line.CharacterCount()
		delta := maxChars - charCount
		if delta > 0 {
			for i := 0; i < delta; i++ {
				output[lid].AppendWordCell(NewNilWordCell(fillerStyle))
			}
		}
		for _, word := range line.Words() {
			output[lid].AppendWordCell(word)
			wid++
		}
		lid++
	}
	return
}

// return output lines where for each input line, any leading space is removed
func (w *CWordLine) applyTypographicJustifyLeft(input []WordLine) (output []WordLine) {
	// trim left space for each line
	wid, lid := 0, 0
	for _, line := range input {
		if lid >= len(output) {
			output = append(output, NewEmptyWordLine())
		}
		start := true
		for _, word := range line.Words() {
			if start {
				if word.IsSpace() {
					continue
				}
				start = false
			}
			if word.IsSpace() {
				if c := word.GetCharacter(0); c != nil {
					wc := NewEmptyWordCell()
					wc.AppendRune(c.Value(), c.Style())
					output[lid].AppendWordCell(wc)
				}
			} else {
				output[lid].AppendWordCell(word)
			}
			wid++
		}
		lid++
	}
	return
}

func (w *CWordLine) applyTypographicJustifyNone(input []WordLine) (output []WordLine) {
	// trim left space for each line
	lid := 0
	for _, line := range input {
		if lid >= len(output) {
			output = append(output, NewEmptyWordLine())
		}
		for _, word := range line.Words() {
			output[lid].AppendWordCell(word)
		}
		lid++
	}
	return
}
