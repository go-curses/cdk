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

package cdk

import (
	"github.com/mattn/go-runewidth"

	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/log"
)

// CellBuffer represents a two dimensional array of character cells.
// This is primarily intended for use by Screen implementors; it
// contains much of the common code they need.  To create one, just
// declare a variable of its type; no explicit initialization is necessary.
//
// CellBuffer should be thread safe, original tcell is not.
type CellBuffer struct {
	w     int
	h     int
	cells []*cell
	valid bool

	// sync.RWMutex
}

func NewCellBuffer() *CellBuffer {
	cb := &CellBuffer{}
	cb.init()
	return cb
}

func (cb *CellBuffer) init() bool {
	if !cb.valid {
		cb.cells = make([]*cell, 0)
		cb.valid = true
		return true
	}
	return false
}

// SetCell sets the contents (primary rune, combining runes,
// and style) for a cell at a given location.
func (cb *CellBuffer) SetCell(x int, y int, mainc rune, combc []rune, style paint.Style) {
	// cb.Lock()
	// defer cb.Unlock()
	if x >= 0 && y >= 0 && x < cb.w && y < cb.h {
		idx := (y * cb.w) + x
		if len(cb.cells) <= idx {
			log.ErrorF("set content index out of range: x=%v,y=%v w=%v,h=%v", x, y, cb.w, cb.h)
			return
		}
		c := cb.cells[idx]
		c.Lock()
		if combc == nil {
			c.currComb = []rune{}
		} else {
			c.currComb = append([]rune{}, combc...)
		}
		if c.currMain != mainc {
			c.width = runewidth.RuneWidth(mainc)
		}
		c.currMain = mainc
		c.currStyle = style
		c.Unlock()
	}
}

// GetCell returns the contents of a character cell, including the
// primary rune, any combining character runes (which will usually be
// nil), the style, and the display width in cells.  (The width can be
// either 1, normally, or 2 for East Asian full-width characters.)
func (cb *CellBuffer) GetCell(x, y int) (mainc rune, combc []rune, style paint.Style, width int) {
	// cb.RLock()
	// defer cb.RUnlock()
	if x >= 0 && y >= 0 && x < cb.w && y < cb.h {
		idx := (y * cb.w) + x
		if len(cb.cells) <= idx {
			log.ErrorDF(1, "index out of range: x=%v,y=%v w=%v,h=%v", x, y, cb.w, cb.h)
			return 0, []rune{}, paint.Style{}, -1
		}
		c := cb.cells[idx]
		c.Lock()
		mainc, combc, style = c.currMain, c.currComb, c.currStyle
		if width = c.width; width == 0 || mainc < ' ' {
			width = 1
			mainc = ' '
		}
		c.Unlock()
	}
	return
}

// Size returns the (width, height) in cells of the buffer.
func (cb *CellBuffer) Size() (w, h int) {
	// cb.RLock()
	// defer cb.RUnlock()
	w, h = cb.w, cb.h
	return
}

// Invalidate marks all characters within the buffer as dirty.
func (cb *CellBuffer) Invalidate() {
	log.TraceF("processing invalidation")
	// cb.Lock()
	// defer cb.Unlock()
	for i := range cb.cells {
		cb.cells[i].Lock()
		cb.cells[i].lastMain = rune(0)
		cb.cells[i].Unlock()
	}
}

// Dirty checks if a character at the given location needs an
// to be refreshed on the physical display.  This returns true
// if the cell content is different since the last time it was
// marked clean.
func (cb *CellBuffer) Dirty(x, y int) bool {
	if x >= 0 && y >= 0 && x < cb.w && y < cb.h {
		c := cb.cells[(y*cb.w)+x]
		c.Lock()
		defer c.Unlock()
		if c.lastMain == rune(0) {
			return true
		}
		if c.lastMain != c.currMain {
			return true
		}
		if c.lastStyle != c.currStyle {
			return true
		}
		if len(c.lastComb) != len(c.currComb) {
			return true
		}
		for i := range c.lastComb {
			if c.lastComb[i] != c.currComb[i] {
				return true
			}
		}
	}
	return false
}

// SetDirty is normally used to indicate that a cell has
// been displayed (in which case dirty is false), or to manually
// force a cell to be marked dirty.
func (cb *CellBuffer) SetDirty(x, y int, dirty bool) {
	if x >= 0 && y >= 0 && x < cb.w && y < cb.h {
		c := cb.cells[(y*cb.w)+x]
		c.Lock()
		defer c.Unlock()
		if dirty {
			c.lastMain = rune(0)
		} else {
			if c.currMain == rune(0) {
				c.currMain = ' '
			}
			c.lastMain = c.currMain
			c.lastComb = c.currComb
			c.lastStyle = c.currStyle
		}
	}
}

// Resize is used to resize the cells array, with different dimensions,
// while preserving the original contents.  The cells will be invalidated
// so that they can be redrawn.
func (cb *CellBuffer) Resize(w, h int) {
	// TraceF("w=%d, h=%d", w, h)
	if cb.h == h && cb.w == w {
		return
	}
	// cb.Lock()
	// defer cb.Unlock()
	if w == 0 || h == 0 {
		cb.cells = make([]*cell, 0)
		return
	}
	newc := make([]*cell, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			nc := newCell()
			k := (y * cb.w) + x
			if len(cb.cells) > k {
				oc := cb.cells[k]
				if oc != nil {
					oc.Lock()
					nc.currMain = oc.currMain
					nc.currComb = oc.currComb
					nc.currStyle = oc.currStyle
					nc.width = oc.width
					oc.Unlock()
				}
			}
			nc.lastMain = rune(0)
			newc[(y*w)+x] = nc
		}
	}
	cb.cells = newc
	cb.h = h
	cb.w = w
}

// Fill fills the entire cell buffer array with the specified character
// and style.  Normally choose ' ' to clear the display.  This API doesn't
// support combining characters, or characters with a width larger than one.
func (cb *CellBuffer) Fill(r rune, style paint.Style) {
	log.TraceF("rune=%v, style=%v", r, style)
	// cb.Lock()
	// defer cb.Unlock()
	for _, c := range cb.cells {
		c.Lock()
		c.currMain = r
		c.currComb = nil
		c.currStyle = style
		c.width = 1
		c.Unlock()
	}
}