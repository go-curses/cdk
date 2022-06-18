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
	"sync"

	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
)

// provide an underlying buffer for Canvases
type SurfaceBuffer interface {
	String() string
	Style() (style paint.Style)
	Size() (size ptypes.Rectangle)
	Width() (width int)
	Height() (height int)
	Resize(size ptypes.Rectangle, style paint.Style)
	GetDim(x, y int) bool
	GetBgColor(x, y int) (bg paint.Color)
	GetCell(x, y int) (textCell TextCell)
	SetCell(x int, y int, r rune, style paint.Style) error
	LoadData(d [][]TextCell)

	sync.Locker
}

// concrete implementation of the SurfaceBuffer interface
type CSurfaceBuffer struct {
	data  [][]*CTextCell
	size  ptypes.Rectangle
	style paint.Style

	sync.Mutex
}

// construct a new canvas buffer
func NewSurfaceBuffer(size ptypes.Rectangle, style paint.Style) *CSurfaceBuffer {
	size.Floor(0, 0)
	b := &CSurfaceBuffer{
		data: make([][]*CTextCell, size.W),
		size: ptypes.MakeRectangle(0, 0),
	}
	b.Resize(size, style)
	return b
}

// return a string describing the buffer, only useful for debugging purposes
func (b *CSurfaceBuffer) String() string {
	return fmt.Sprintf(
		"{Size=%s}",
		b.size,
	)
}

// return the rectangle size of the buffer
func (b *CSurfaceBuffer) Style() (style paint.Style) {
	return b.style
}

// return the rectangle size of the buffer
func (b *CSurfaceBuffer) Size() (size ptypes.Rectangle) {
	return b.size
}

// return just the width of the buffer
func (b *CSurfaceBuffer) Width() (width int) {
	return b.size.W
}

// return just the height of the buffer
func (b *CSurfaceBuffer) Height() (height int) {
	return b.size.H
}

// resize the buffer
func (b *CSurfaceBuffer) Resize(size ptypes.Rectangle, style paint.Style) {
	b.Lock()
	defer b.Unlock()
	size.Floor(0, 0)
	if b.size.W == size.W && b.size.H == size.H && b.style.String() == style.String() {
		return
	}
	// fill size, expanding as necessary
	for x := 0; x < size.W; x++ {
		if len(b.data) <= x {
			// need more X space
			b.data = append(b.data, make([]*CTextCell, size.H))
		}
		// fill in Y space for this X space
		for y := 0; y < size.H; y++ {
			if len(b.data[x]) <= y {
				// add Y space
				b.data[x] = append(b.data[x], NewRuneCell(' ', style))
			} else if b.data[x][y] == nil {
				// fill nil Y space
				b.data[x][y] = NewRuneCell(' ', style)
			} else {
				// clear Y space
				b.data[x][y].Set(' ')
				b.data[x][y].SetStyle(style)
			}
		}
	}
	// truncate excess cells
	if b.size.W > size.W {
		// the previous size was larger than this one
		// truncate X space
		b.data = b.data[:size.W]
	}
	if b.size.H > size.H {
		// previous size was larger than this one
		for x := 0; x < size.W; x++ {
			if len(b.data) <= x {
				b.data = append(b.data, make([]*CTextCell, size.H))
			}
			if len(b.data[x]) >= size.H {
				// truncate, too long
				b.data[x] = b.data[x][:size.H]
			} else {
				for y := 0; y < size.H; y++ {
					if len(b.data[x]) <= y {
						// add Y space
						b.data[x] = append(b.data[x], NewRuneCell(' ', style))
					} else if b.data[x][y] == nil {
						// fill nil Y space
						b.data[x][y] = NewRuneCell(' ', style)
					} else {
						// clear Y space
						b.data[x][y].Set(' ')
						b.data[x][y].SetStyle(style)
					}
				}
			}
		}
	}
	// store the size
	b.size = size
}

// return the text cell at the given coordinates, nil if not found
func (b *CSurfaceBuffer) GetCell(x int, y int) TextCell {
	b.Lock() // lock so that resize floods don't enable race conditions
	defer b.Unlock()
	if x >= 0 && y >= 0 && x < b.size.W && y < b.size.H {
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

// set the cell content at the given coordinates
func (b *CSurfaceBuffer) SetCell(x int, y int, r rune, style paint.Style) error {
	dxLen := len(b.data)
	if x >= 0 && x < dxLen {
		dyLen := len(b.data[x])
		if y >= 0 && y < dyLen {
			b.data[x][y].Set(r)
			b.data[x][y].SetStyle(style)
			if count := b.data[x][y].Count(); count > 1 {
				for i := 1; i < count; i++ {
					if xi := x + i; xi < dxLen {
						b.data[xi][y].SetStyle(style)
					}
				}
			}
			return nil
		}
		return fmt.Errorf("y=%v not in range [0-%d]", y, len(b.data[x])-1)
	}
	return fmt.Errorf("x=%v not in range [0-%d]", x, len(b.data)-1)
}

// given matrix array of text cells, load that data in this canvas space
func (b *CSurfaceBuffer) LoadData(d [][]TextCell) {
	b.Lock()
	defer b.Unlock()
	for x := 0; x < len(d); x++ {
		for y := 0; y < len(d[x]); y++ {
			if y >= len(b.data[x]) {
				b.data[x] = append(b.data[x], NewRuneCell(d[x][y].Value(), d[x][y].Style()))
			} else {
				b.data[x][y].Set(d[x][y].Value())
			}
		}
	}
}