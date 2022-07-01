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
	"unicode/utf8"

	"github.com/gofrs/uuid"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/math"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
)

// a Surface is the primary means of drawing to the terminal display within CDK
type Surface interface {
	GetStyle() (style paint.Style)
	SetStyle(style paint.Style)
	String() string
	Resize(size ptypes.Rectangle)
	GetContent(x, y int) (textCell TextCell)
	SetContent(x, y int, char string, s paint.Style) error
	SetRune(x, y int, r rune, s paint.Style) error
	SetOrigin(origin ptypes.Point2I)
	GetOrigin() ptypes.Point2I
	GetSize() ptypes.Rectangle
	Width() (width int)
	Height() (height int)
	GetRegion() (region ptypes.Region)
	SetRegion(region ptypes.Region)
	Equals(onlyDirty bool, v *CSurface) bool
	CompositeSurface(v *CSurface) error
	Composite(id uuid.UUID) (err error)
	Render(display Renderer) error
	DrawText(pos ptypes.Point2I, size ptypes.Rectangle, justify enums.Justification, singleLineMode bool, wrap enums.WrapMode, ellipsize bool, style paint.Style, markup, mnemonic bool, text string)
	DrawSingleLineText(position ptypes.Point2I, maxChars int, ellipsize bool, justify enums.Justification, style paint.Style, markup, mnemonic bool, text string)
	DrawLine(pos ptypes.Point2I, length int, orient enums.Orientation, style paint.Style)
	DrawHorizontalLine(pos ptypes.Point2I, length int, style paint.Style)
	DrawVerticalLine(pos ptypes.Point2I, length int, style paint.Style)
	Box(pos ptypes.Point2I, size ptypes.Rectangle, border, fill, overlay bool, fillRune rune, contentStyle, borderStyle paint.Style, borderRunes paint.BorderRuneSet)
	BoxWithTheme(pos ptypes.Point2I, size ptypes.Rectangle, border, fill bool, theme paint.Theme)
	DebugBox(color paint.Color, format string, argv ...interface{})
	Fill(theme paint.Theme)
	FillBorder(dim, border bool, theme paint.Theme)
	FillBorderTitle(dim bool, title string, justify enums.Justification, theme paint.Theme)
}

// concrete implementation of the Surface interface
type CSurface struct {
	buffer *CSurfaceBuffer
	origin ptypes.Point2I
	fill   rune

	sync.RWMutex
}

// create a new canvas object with the given origin point, size and theme
func NewSurface(origin ptypes.Point2I, size ptypes.Rectangle, style paint.Style) *CSurface {
	c := &CSurface{
		buffer: NewSurfaceBuffer(size, style),
		origin: origin,
		fill:   ' ',
	}
	return c
}

func (c *CSurface) GetStyle() (style paint.Style) {
	c.RLock()
	defer c.RUnlock()
	return c.buffer.Style()
}

func (c *CSurface) SetStyle(style paint.Style) {
	c.Lock()
	defer c.Unlock()
	c.buffer.Lock()
	defer c.buffer.Unlock()
	c.buffer.style = style
}

// return a string describing the canvas metadata, useful for debugging
func (c *CSurface) String() string {
	c.RLock()
	defer c.RUnlock()
	return fmt.Sprintf(
		"{Origin=%s,Fill=%v,Buffer=%v}",
		c.origin,
		c.fill,
		c.buffer.String(),
	)
}

// change the size of the canvas, not recommended to do this in practice
func (c *CSurface) Resize(size ptypes.Rectangle) {
	c.Lock()
	defer c.Unlock()
	c.buffer.Resize(size)
}

// get the text cell at the given coordinates
func (c *CSurface) GetContent(x, y int) (textCell TextCell) {
	c.RLock()
	defer c.RUnlock()
	return c.buffer.GetCell(x, y)
}

// from the given string, set the character and style of the cell at the given
// coordinates. note that only the first UTF-8 byte is used
func (c *CSurface) SetContent(x, y int, char string, s paint.Style) error {
	c.Lock()
	defer c.Unlock()
	r, _ := utf8.DecodeRune([]byte(char))
	return c.buffer.SetCell(x, y, r, s)
}

// set the rune and the style of the cell at the given coordinates
func (c *CSurface) SetRune(x, y int, r rune, s paint.Style) error {
	c.Lock()
	defer c.Unlock()
	return c.buffer.SetCell(x, y, r, s)
}

// set the origin (top-left corner) position of the canvas, used when
// compositing one canvas with another
func (c *CSurface) SetOrigin(origin ptypes.Point2I) {
	c.Lock()
	defer c.Unlock()
	c.origin = origin
}

// get the origin point of the canvas
func (c *CSurface) GetOrigin() ptypes.Point2I {
	c.RLock()
	defer c.RUnlock()
	return c.origin.Clone()
}

// get the rectangle size of the canvas
func (c *CSurface) GetSize() ptypes.Rectangle {
	c.RLock()
	defer c.RUnlock()
	return c.buffer.Size()
}

// convenience method to get just the width of the canvas
func (c *CSurface) Width() (width int) {
	c.RLock()
	defer c.RUnlock()
	return c.buffer.Size().W
}

// convenience method to get just the height of the canvas
func (c *CSurface) Height() (height int) {
	c.RLock()
	defer c.RUnlock()
	return c.buffer.Size().H
}

// GetRegion returns the origin and size as a Region type.
func (c *CSurface) GetRegion() (region ptypes.Region) {
	origin := c.GetOrigin()
	size := c.GetSize()
	region = ptypes.MakeRegion(origin.X, origin.Y, size.W, size.H)
	return
}

// SetRegion is a convenience method wrapping SetOrigin and Resize, using the
// given Region for the values. The existing Style is used for the Resize call.
func (c *CSurface) SetRegion(region ptypes.Region) {
	c.SetOrigin(region.Origin())
	c.Resize(region.Size())
}

// returns true if the given canvas is painted the same as this one, can compare
// for only cells that were "set" (dirty) or compare every cell of the two
// canvases
func (c *CSurface) Equals(onlyDirty bool, v *CSurface) bool {
	vOrigin := v.GetOrigin()
	bSize := c.buffer.Size()
	vSize := v.GetSize()
	if c.origin.EqualsTo(vOrigin) {
		if bSize.EqualsTo(vSize) {
			for x := 0; x < vSize.W; x++ {
				for y := 0; y < vSize.H; y++ {
					ca := c.GetContent(x, y)
					va := v.GetContent(x, y)
					if !onlyDirty || (onlyDirty && va.Dirty()) {
						if ca.Style() != va.Style() {
							return false
						}
						if ca.Value() != va.Value() {
							return false
						}
					}
				}
			}
		}
	}
	return true
}

// apply the given canvas to this canvas, at the given one's origin. returns
// an error if the underlying buffer write failed or if the given canvas is
// beyond the bounds of this canvas
func (c *CSurface) CompositeSurface(src *CSurface) error {
	if c == nil || c.buffer == nil {
		return fmt.Errorf("canvas is nil")
	}

	dstSize := c.buffer.Size()
	srcSize := src.buffer.Size()

	if dstSize.W <= 0 || srcSize.W <= 0 || dstSize.H <= 0 || srcSize.H <= 0 {
		return fmt.Errorf("either canvas has zero size")
	}

	dstOrigin := c.GetOrigin()
	srcOrigin := src.GetOrigin()
	srcOrigin.SubPoint(dstOrigin)

	c.Lock()
	defer c.Unlock()
	src.Lock()
	defer src.Unlock()

	for x := 0; x < srcSize.W; x++ {
		for y := 0; y < srcSize.H; y++ {
			if cell := src.buffer.GetCell(x, y); cell != nil && !cell.IsNil() {
				local := srcOrigin.Clone()
				local.Add(x, y)
				local.ClampMin(0, 0)
				if local.X < dstSize.W && local.Y < dstSize.H {
					if err := c.buffer.SetCell(
						local.X,
						local.Y,
						cell.Value(),
						cell.Style(),
					); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (c *CSurface) Composite(id uuid.UUID) (err error) {
	if surface, err := GetSurface(id); err == nil {
		return c.CompositeSurface(surface)
	}
	return
}

// render this canvas upon the given display
func (c *CSurface) Render(display Renderer) error {
	origin := c.GetOrigin()
	size := c.GetSize()
	c.Lock()
	defer c.Unlock()
	for x := 0; x < size.W; x++ {
		for y := 0; y < size.H; y++ {
			cell := c.buffer.GetCell(x, y)
			if cell != nil {
				if cell.Dirty() {
					mc, _, style, width := display.GetContent(x, y)
					if !cell.Equals(mc, style, width) {
						display.SetContent(origin.X+x, origin.Y+y, cell.Value(), nil, cell.Style())
					}
				}
			} else {
				bs := c.buffer.Size()
				log.TraceF(
					"invalid cell coordinates: x=%v, y=%v (valid: x=[%v-%v], y=[%v-%v])",
					x, y,
					0, bs.W-1,
					0, bs.H-1,
				)
			}
		}
	}
	return nil
}

// Write text to the canvas buffer
// origin is the top-left coordinate for the text area being rendered
// alignment is based on origin.X boxed by maxChars or canvas size.W
func (c *CSurface) DrawText(pos ptypes.Point2I, size ptypes.Rectangle, justify enums.Justification, singleLineMode bool, wrap enums.WrapMode, ellipsize bool, style paint.Style, markup, mnemonic bool, text string) {
	var tb TextBuffer
	if markup {
		m, err := NewMarkup(text, style)
		if err != nil {
			log.FatalDF(1, "failed to parse markup: %v", err)
		}
		tb = m.TextBuffer(mnemonic)
	} else {
		tb = NewTextBuffer(text, style, mnemonic)
	}
	cSize := c.GetSize()
	if size.W == -1 || size.W >= cSize.W {
		size.W = cSize.W
	}
	v := NewSurface(pos, size, style)
	theme := paint.DefaultColorTheme
	theme.Content.Normal = style
	theme.Content.FillRune = rune(0)
	v.Fill(theme)
	tb.Draw(v, singleLineMode, wrap, ellipsize, justify, enums.ALIGN_TOP)
	if err := c.CompositeSurface(v); err != nil {
		log.ErrorF("composite error: %v", err)
	}
}

// write a single line of text to the canvas at the given position, of at most
// maxChars, with the text justified and styled. supports Tango markup content
func (c *CSurface) DrawSingleLineText(position ptypes.Point2I, maxChars int, ellipsize bool, justify enums.Justification, style paint.Style, markup, mnemonic bool, text string) {
	c.DrawText(position, ptypes.MakeRectangle(maxChars, 1), justify, true, enums.WRAP_NONE, ellipsize, style, markup, mnemonic, text)
}

// draw a line vertically or horizontally with the given style
func (c *CSurface) DrawLine(pos ptypes.Point2I, length int, orient enums.Orientation, style paint.Style) {
	log.TraceF("c.DrawLine(%v,%v,%v,%v)", pos, length, orient, style)
	switch orient {
	case enums.ORIENTATION_HORIZONTAL:
		c.DrawHorizontalLine(pos, length, style)
	case enums.ORIENTATION_VERTICAL:
		c.DrawVerticalLine(pos, length, style)
	}
}

// convenience method to draw a horizontal line
func (c *CSurface) DrawHorizontalLine(pos ptypes.Point2I, length int, style paint.Style) {
	c.Lock()
	defer c.Unlock()
	length = math.ClampI(length, pos.X, c.GetSize().W-pos.X)
	end := pos.X + length
	for i := pos.X; i < end; i++ {
		_ = c.buffer.SetCell(i, pos.Y, paint.RuneHLine, style)
	}
}

// convenience method to draw a vertical line
func (c *CSurface) DrawVerticalLine(pos ptypes.Point2I, length int, style paint.Style) {
	c.Lock()
	defer c.Unlock()
	length = math.ClampI(length, pos.Y, c.GetSize().H-pos.Y)
	end := pos.Y + length
	for i := pos.Y; i < end; i++ {
		_ = c.buffer.SetCell(i, pos.Y, paint.RuneVLine, style)
	}
}

// draw a box, at position, of size, with or without a border, with or without
// being filled in and following the given theme
func (c *CSurface) Box(pos ptypes.Point2I, size ptypes.Rectangle, border, fill, overlay bool, fillRune rune, contentStyle, borderStyle paint.Style, borderRunes paint.BorderRuneSet) {
	c.Lock()
	defer c.Unlock()
	log.TraceDF(1, "c.Box(%v,%v,%v,%v,%v,%v,%v,%v,%v)", pos, size, border, fill, overlay, fillRune, contentStyle, borderStyle, borderRunes)
	xEnd := pos.X + size.W - 1
	yEnd := pos.Y + size.H - 1
	// for each column
	for ix := pos.X; ix < (pos.X + size.W); ix++ {
		// for each row
		for iy := pos.Y; iy < (pos.Y + size.H); iy++ {
			if overlay {
				borderStyle = borderStyle.
					Background(c.buffer.GetBgColor(ix, iy)).
					Dim(c.buffer.GetDim(ix, iy))
				contentStyle = contentStyle.
					Background(c.buffer.GetBgColor(ix, iy)).
					Dim(c.buffer.GetDim(ix, iy))
			}
			switch {
			case ix == pos.X:
				// left column
				switch {
				case iy == pos.Y && border:
					// top left corner
					_ = c.buffer.SetCell(ix, iy, borderRunes.TopLeft, borderStyle)
				case iy == yEnd && border:
					// bottom left corner
					_ = c.buffer.SetCell(ix, iy, borderRunes.BottomLeft, borderStyle)
				default:
					// left border
					if border {
						_ = c.buffer.SetCell(ix, iy, borderRunes.Left, borderStyle)
					} else if fill {
						_ = c.buffer.SetCell(ix, iy, fillRune, contentStyle)
					}
				} // left column switch
			case ix == xEnd:
				// right column
				switch {
				case iy == pos.Y && border:
					// top right corner
					_ = c.buffer.SetCell(ix, iy, borderRunes.TopRight, borderStyle)
				case iy == yEnd && border:
					// bottom right corner
					_ = c.buffer.SetCell(ix, iy, borderRunes.BottomRight, borderStyle)
				default:
					// right border
					if border {
						_ = c.buffer.SetCell(ix, iy, borderRunes.Right, borderStyle)
					} else if fill {
						_ = c.buffer.SetCell(ix, iy, fillRune, contentStyle)
					}
				} // right column switch
			default:
				// middle columns
				switch {
				case iy == pos.Y && border:
					// top middle
					_ = c.buffer.SetCell(ix, iy, borderRunes.Top, borderStyle)
				case iy == yEnd && border:
					// bottom middle
					_ = c.buffer.SetCell(ix, iy, borderRunes.Bottom, borderStyle)
				default:
					// middle middle
					if fill {
						_ = c.buffer.SetCell(ix, iy, fillRune, contentStyle)
					}
				} // middle columns switch
			} // draw switch
		} // for iy
	} // for ix
}

func (c *CSurface) BoxWithTheme(pos ptypes.Point2I, size ptypes.Rectangle, border, fill bool, theme paint.Theme) {
	c.Box(
		pos,
		size,
		border,
		fill,
		false,
		theme.Border.FillRune,
		theme.Border.Normal,
		theme.Border.Normal,
		theme.Border.BorderRunes,
	)
}

// draw a box with Sprintf-formatted text along the top-left of the box, useful
// for debugging more than anything else as the normal draw primitives are far
// more flexible
func (c *CSurface) DebugBox(color paint.Color, format string, argv ...interface{}) {
	text := fmt.Sprintf(format, argv...)
	cSize := c.GetSize()
	if cSize.Equals(0, 0) {
		log.DebugDF(1, "[DebugBox] (zero-size) info: %v (%v)", text, color)
		return
	}
	log.DebugDF(1, "[DebugBox] info: %v (%v)", text, color)
	bs := paint.DefaultMonoTheme // intentionally mono
	bs.Border.Normal = bs.Border.Normal.Foreground(color)
	c.Box(
		ptypes.MakePoint2I(0, 0),
		cSize,
		true,
		false,
		false,
		bs.Border.FillRune,
		bs.Border.Normal,
		bs.Border.Normal,
		bs.Border.BorderRunes,
	)
	c.DrawSingleLineText(ptypes.MakePoint2I(c.origin.X+1, c.origin.Y), cSize.W-2, false, enums.JUSTIFY_LEFT, bs.Border.Normal, false, false, text)
}

// fill the entire canvas according to the given theme
func (c *CSurface) Fill(theme paint.Theme) {
	log.TraceF("c.Fill(%v,%v)", theme)
	c.Box(
		ptypes.MakePoint2I(0, 0),
		c.GetSize(),
		false, true,
		theme.Content.Overlay,
		theme.Content.FillRune,
		theme.Content.Normal,
		theme.Border.Normal,
		theme.Border.BorderRunes,
	)
}

// fill the entire canvas, with or without 'dim' styling, with or without a
// border
func (c *CSurface) FillBorder(dim, border bool, theme paint.Theme) {
	cSize := c.GetSize()
	log.TraceF("c.FillBorder(%v,%v): origin=%v, size=%v", border, theme, c.origin, cSize)
	theme.Content.Normal = theme.Content.Normal.Dim(dim)
	theme.Border.Normal = theme.Border.Normal.Dim(dim)
	c.Box(
		ptypes.MakePoint2I(0, 0),
		cSize,
		border,
		true,
		theme.Content.Overlay,
		theme.Content.FillRune,
		theme.Content.Normal,
		theme.Border.Normal,
		theme.Border.BorderRunes,
	)
}

// fill the entire canvas, with or without 'dim' styling, with plain text
// justified across the top border
func (c *CSurface) FillBorderTitle(dim bool, title string, justify enums.Justification, theme paint.Theme) {
	log.TraceF("c.FillBorderTitle(%v,%v,%v)", title, justify, theme)
	theme.Content.Normal = theme.Content.Normal.Dim(dim)
	theme.Border.Normal = theme.Border.Normal.Dim(dim)
	cSize := c.GetSize()
	c.Box(
		ptypes.MakePoint2I(0, 0),
		cSize,
		true,
		true,
		theme.Content.Overlay,
		theme.Content.FillRune,
		theme.Content.Normal,
		theme.Border.Normal,
		theme.Border.BorderRunes,
	)
	c.DrawSingleLineText(ptypes.MakePoint2I(1, 0), cSize.W-2, false, justify, theme.Content.Normal.Dim(dim), false, false, title)
}