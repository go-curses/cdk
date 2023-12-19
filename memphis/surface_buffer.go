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

	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
)

// provide an underlying buffer for Canvases
type SurfaceBuffer interface {
	String() string
	Style() (style paint.Style)
	SetStyle(style paint.Style)
	Size() (size ptypes.Rectangle)
	Width() (width int)
	Height() (height int)
	Resize(size ptypes.Rectangle)
	GetDim(x, y int) bool
	GetBgColor(x, y int) (bg paint.Color)
	GetCell(x, y int) (textCell TextCell)
	SetCell(x int, y int, r rune, style paint.Style) error
	LoadData(d [][]TextCell)
}

// concrete implementation of the SurfaceBuffer interface
type CSurfaceBuffer struct {
	data  [][]*CTextCell
	style paint.Style

	sync.RWMutex
}

// construct a new canvas buffer
func NewSurfaceBuffer(size ptypes.Rectangle, style paint.Style) *CSurfaceBuffer {
	size.Floor(0, 0)
	b := &CSurfaceBuffer{
		data: make([][]*CTextCell, size.W),
	}
	b.style = style
	b.Resize(size)
	return b
}

// return a string describing the buffer, only useful for debugging purposes
func (b *CSurfaceBuffer) String() string {
	b.RLock()
	defer b.RUnlock()
	return fmt.Sprintf(
		"{Size=%s}",
		b.Size(),
	)
}

// return the rectangle size of the buffer
func (b *CSurfaceBuffer) Style() (style paint.Style) {
	b.RLock()
	defer b.RUnlock()
	return b.style
}

// return the rectangle size of the buffer
func (b *CSurfaceBuffer) Size() (size ptypes.Rectangle) {
	b.RLock()
	defer b.RUnlock()
	w, h := len(b.data), 0
	if w > 0 {
		h = len(b.data[0])
	}
	size = ptypes.MakeRectangle(w, h)
	return
}

// return the rectangle size of the buffer
func (b *CSurfaceBuffer) SetStyle(style paint.Style) {
	b.RLock()
	defer b.RUnlock()
	b.style = style
}

// resize the buffer
func (b *CSurfaceBuffer) Resize(size ptypes.Rectangle) {
	b.Lock()
	defer b.Unlock()

	size.Floor(0, 0)

	if size.Equals(0, 0) || size.W == 0 || size.H == 0 {
		if len(b.data) > 0 {
			b.data = make([][]*CTextCell, 0)
		}
		return
	}

	if size.W == len(b.data) && size.H == len(b.data[0]) {
		// same size, nop
		return
	}

	b.data = make([][]*CTextCell, size.W)
	for x := 0; x < size.W; x++ {
		b.data[x] = make([]*CTextCell, size.H)
		for y := 0; y < size.H; y++ {
			b.data[x][y] = NewTextCellFromRune(' ', b.style)
		}
	}
}

// return the text cell at the given coordinates, nil if not found
func (b *CSurfaceBuffer) GetCell(x int, y int) TextCell {
	b.RLock() // lock so that resize floods don't enable race conditions
	defer b.RUnlock()
	bSize := b.Size()
	if x >= 0 && y >= 0 && x < bSize.W && y < bSize.H {
		return b.data[x][y]
	}
	return nil
}

// return true if the given coordinates are styled 'dim', false otherwise
func (b *CSurfaceBuffer) GetDim(x, y int) bool {
	c := b.GetCell(x, y)
	s := c.Style()
	_, _, a := s.Decompose()
	return a.IsDim()
}

// return the background color at the given coordinates
func (b *CSurfaceBuffer) GetBgColor(x, y int) (bg paint.Color) {
	c := b.GetCell(x, y)
	s := c.Style()
	_, bg, _ = s.Decompose()
	return
}

// set the cell content at the given (literal) coordinates
func (b *CSurfaceBuffer) SetCell(x int, y int, r rune, style paint.Style) error {
	b.Lock()
	defer b.Unlock()
	dxLen := len(b.data)
	if dxLen == 0 {
		return fmt.Errorf("surface has zero size")
	}
	if x >= 0 && x < dxLen {
		dyLen := len(b.data[x])
		if y >= 0 && y < dyLen {
			if b.data[x][y] == nil {
				b.data[x][y] = NewTextCellFromRune(r, style)
			} else {
				b.data[x][y].Set(r)
				b.data[x][y].SetStyle(style)
			}
			// wide runes need extra care for their neighbor... sometimes...
			// not sure how to best figure out if a rune actually consumes more
			// than one monospace character
			if count := b.data[x][y].Count(); count > 1 {
				for i := 1; i < count; i++ {
					if xi := x + i; xi < dxLen {
						b.data[xi][y].SetStyle(style)
					}
				}
			}
			return nil
		}
		return fmt.Errorf("y=%v not in range [0,%d]", y, len(b.data[x])-1)
	}
	return fmt.Errorf("x=%v not in range [0,%d]", x, len(b.data)-1)
}

// given matrix array of text cells, load that data in this canvas space
func (b *CSurfaceBuffer) LoadData(d [][]TextCell) {
	b.Lock()
	defer b.Unlock()
	for x := 0; x < len(d); x++ {
		for y := 0; y < len(d[x]); y++ {
			if y >= len(b.data[x]) {
				b.data[x] = append(b.data[x], NewTextCellFromRune(d[x][y].Value(), d[x][y].Style()))
			} else {
				b.data[x][y].Set(d[x][y].Value())
			}
		}
	}
}
