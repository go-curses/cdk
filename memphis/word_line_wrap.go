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
	"github.com/go-curses/cdk/lib/paint"
)

// wrap the input lines on the nearest word to maxChars
func (w *CWordLine) applyTypographicWrapWord(maxChars int, input []WordLine) (output []WordLine) {
	cid, wid, lid := 0, 0, 0
	for _, line := range input {
		if lid >= len(output) {
			output = append(output, NewEmptyWordLine())
		}
		if maxChars > -1 && line.CharacterCount() > maxChars {
			if !line.HasSpace() {
				// nothing to break on, truncate on maxChars
				for _, word := range line.Words() {
					if wid >= output[lid].Len() {
						output[lid].AppendWordCell(NewEmptyWordCell())
					}
					for _, c := range word.Characters() {
						if cid >= maxChars {
							lid = len(output) // don't append trailing NEWLs
							break
						}
						_ = output[lid].AppendWordRune(wid, c.Value(), c.Style())
						cid++
					}
					wid++
				}
				continue
			}
		}
		for _, word := range line.Words() {
			wordLen := word.Len()
			if maxChars > -1 && cid+wordLen > maxChars {
				output = append(output, NewEmptyWordLine())
				lid = len(output) - 1
				wid = -1
				cid = 0
				output[lid].AppendWordCell(word)
				wid = output[lid].Len() - 1
				cid += word.Len()
			} else if word.IsSpace() && maxChars > -1 && cid+wordLen+1 > maxChars {
				// continue
			} else {
				output[lid].AppendWordCell(word)
				cid += word.Len()
			}
			wid++
		}
		lid++
		cid = 0
		wid = 0
	}
	return
}

// wrap the input lines on the nearest word to maxChars if the line has space,
// else, truncate at maxChars
func (w *CWordLine) applyTypographicWrapWordChar(maxChars int, input []WordLine) (output []WordLine) {
	for lid, line := range input {
		if lid >= len(output) {
			output = append(output, NewEmptyWordLine())
		}
		if maxChars > -1 && line.CharacterCount() > maxChars {
			if !line.HasSpace() {
				wrapped := w.applyTypographicWrapChar(maxChars, []WordLine{line})
				for wLid, wLine := range wrapped {
					id := lid + wLid
					if id >= len(output) {
						output = append(output, NewEmptyWordLine())
					}
					for _, wWord := range wLine.Words() {
						output[id].AppendWordCell(wWord)
					}
				}
				continue
			}
		}
		wrapped := w.applyTypographicWrapWord(maxChars, []WordLine{line})
		for wLid, wLine := range wrapped {
			id := lid + wLid
			if id >= len(output) {
				output = append(output, NewEmptyWordLine())
			}
			for _, wWord := range wLine.Words() {
				output[id].AppendWordCell(wWord)
			}
		}
	}
	return
}

// wrap the input lines on the nearest character to maxChars
func (w *CWordLine) applyTypographicWrapChar(maxChars int, input []WordLine) (output []WordLine) {
	cid, wid, lid := 0, 0, 0
	for _, line := range input {
		if lid >= len(output) {
			output = append(output, NewEmptyWordLine())
		}
		for _, word := range line.Words() {
			if maxChars > -1 && cid+word.Len() > maxChars {
				firstHalf, secondHalf := NewEmptyWordCell(), NewEmptyWordCell()
				for _, c := range word.Characters() {
					if cid < maxChars {
						firstHalf.AppendRune(c.Value(), c.Style())
					} else {
						secondHalf.AppendRune(c.Value(), c.Style())
					}
					cid++
				}
				output[lid].AppendWordCell(firstHalf)
				output = append(output, NewEmptyWordLine())
				lid = len(output) - 1
				output[lid].AppendWordCell(secondHalf)
				wid = 0
				cid = 0
			} else {
				output[lid].AppendWordCell(word)
				cid += word.Len()
			}
			wid++
		}
		lid++
		cid = 0
	}
	return
}

// truncate the input lines on the nearest character to maxChars
func (w *CWordLine) applyTypographicWrapNone(ellipsize bool, maxChars int, input []WordLine) (output []WordLine) {
	cid, lid := 0, 0
	for _, line := range input {
		if lid >= len(output) {
			output = append(output, NewEmptyWordLine())
		}
		for _, word := range line.Words() {
			if maxChars > -1 && cid+word.Len() > maxChars {
				wc := NewEmptyWordCell()
				for _, c := range word.Characters() {
					if cid+c.Count() > maxChars {
						break
					}
					wc.AppendRune(c.Value(), c.Style())
					cid += c.Count()
				}
				if wc.Len() > 0 {
					output[lid].AppendWordCell(wc)
					if ellipsize {
						// ellipsize here
						eStartIndex := output[lid].CharacterCount() - 1
						if eStartIndex > 0 {
							output[lid].SetCharacter(eStartIndex, paint.RuneEllipsis)
						}
					}
					break
				}
			} else {
				output[lid].AppendWordCell(word)
				cid += word.Len()
			}
		}
		lid++
		cid = 0
	}
	return
}