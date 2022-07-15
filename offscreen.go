// Copyright 2021  The CDK Authors
// Copyright 2016 The TCell Authors
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
	"fmt"
	"os"
	"unicode/utf8"

	"golang.org/x/text/transform"

	ccharset "github.com/go-curses/cdk/charset"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"

	cstrings "github.com/go-curses/cdk/lib/strings"
)

const OffscreenTtyPath = "<offscreen>"

func MakeOffScreen(charset string) (OffScreen, error) {
	s := NewOffScreen(charset)
	if s == nil {
		return nil, fmt.Errorf("failed to get offscreen")
	}
	if e := s.Init(); e != nil {
		return nil, fmt.Errorf("failed to initialize offscreen: %v", e)
	}
	return s, nil
}

// NewOffScreen returns a OffScreen.  Note that
// OffScreen is also a Display.
func NewOffScreen(charset string) OffScreen {
	if cstrings.IsEmpty(charset) {
		charset = ccharset.Get()
	}
	s := &COffScreen{charset: charset}
	return s
}

// OffScreen represents a screen simulation.  This is intended to
// be a superset of normal Screens, but also adds some important interfaces
// for testing.
type OffScreen interface {
	// InjectKeyBytes injects a stream of bytes corresponding to
	// the native encoding (see charset).  It turns true if the entire
	// set of bytes were processed and delivered as KeyEvents, false
	// if any bytes were not fully understood.  Any bytes that are not
	// fully converted are discarded.
	InjectKeyBytes(buf []byte) bool

	// InjectKey injects a key event.  The rune is a UTF-8 rune, post
	// any translation.
	InjectKey(key Key, r rune, mod ModMask)

	// InjectMouse injects a mouse event.
	InjectMouse(x, y int, buttons ButtonMask, mod ModMask)

	// SetSize resizes the underlying physical screen.  It also causes
	// a resize event to be injected during the next Show() or Sync().
	// A new physical contents array will be allocated (with data from
	// the old copied), so any prior value obtained with GetContents
	// won't be used anymore
	SetSize(width, height int)

	// GetContents returns screen contents as an array of
	// cells, along with the physical width & height.   Note that the
	// physical contents will be used until the next time SetSize()
	// is called.
	GetContents() (cells []OffscreenCell, width int, height int)

	// GetCursor returns the cursor details.
	GetCursor() (x int, y int, visible bool)

	Screen
}

// OffscreenCell represents a simulated screen cell.  The purpose of this
// is to track on screen content.
type OffscreenCell struct {
	// Bytes is the actual character bytes.  Normally this is
	// rune data, but it could be be data in another encoding system.
	Bytes []byte

	// Style is the style used to display the data.
	Style paint.Style

	// Runes is the list of runes, unadulterated, in UTF-8.
	Runes []rune
}

type COffScreen struct {
	physW    int
	physH    int
	finished bool
	style    paint.Style
	evCh     chan Event
	quit     chan struct{}

	front     []OffscreenCell
	back      *CellBuffer
	clear     bool
	cursorX   int
	cursorY   int
	cursorVis bool
	mouse     bool
	paste     bool
	charset   string
	encoder   transform.Transformer
	decoder   transform.Transformer
	fillChar  rune
	fillStyle paint.Style
	fallback  map[rune]string

	sync.Mutex
}

func (o *COffScreen) InitWithFilePath(ttyFile string) (err error) {
	return o.Init()
}

func (o *COffScreen) InitWithFileHandle(ttyHandle *os.File) (err error) {
	return o.Init()
}

func (o *COffScreen) TtyKeepFileHandle(keep bool) {
}

func (o *COffScreen) TtyKeepingFileHandle() (keeping bool) {
	return false
}

func (o *COffScreen) TtyCloseWithStiRead(enabled bool) {}

func (o *COffScreen) GetTtyCloseWithStiRead() (enabled bool) {
	return
}

func (o *COffScreen) Init() error {
	o.evCh = make(chan Event, 10)
	o.quit = make(chan struct{})
	o.fillChar = 'X'
	o.fillStyle = paint.StyleDefault
	o.mouse = false
	o.physW = 80
	o.physH = 25
	o.cursorX = -1
	o.cursorY = -1
	o.style = paint.StyleDefault
	o.back = NewCellBuffer()

	if enc := GetEncoding(o.charset); enc != nil {
		o.encoder = enc.NewEncoder()
		o.decoder = enc.NewDecoder()
	} else {
		return ErrNoCharset
	}

	o.front = make([]OffscreenCell, o.physW*o.physH)
	o.back.Resize(80, 25)

	// default fallbacks
	o.fallback = make(map[rune]string)
	for k, v := range paint.RuneFallbacks {
		o.fallback[k] = v
	}
	return nil
}

func (o *COffScreen) Close() {
	o.Lock()
	defer o.Unlock()
	o.finished = true
	o.back.Resize(0, 0)
	if o.quit != nil {
		close(o.quit)
	}
	o.physW = 0
	o.physH = 0
	o.front = nil
}

func (o *COffScreen) SetStyle(style paint.Style) {
	o.Lock()
	defer o.Unlock()
	o.style = style
}

func (o *COffScreen) Clear() {
	o.Fill(' ', o.style)
}

func (o *COffScreen) Fill(r rune, style paint.Style) {
	o.Lock()
	defer o.Unlock()
	o.back.Fill(r, style)
}

func (o *COffScreen) SetCell(x, y int, style paint.Style, ch ...rune) {
	if len(ch) > 0 {
		o.SetContent(x, y, ch[0], ch[1:], style)
	} else {
		o.SetContent(x, y, ' ', nil, style)
	}
}

func (o *COffScreen) SetContent(x, y int, mc rune, comb []rune, st paint.Style) {
	o.Lock()
	defer o.Unlock()
	o.back.SetCell(x, y, mc, comb, st)
}

func (o *COffScreen) GetContent(x, y int) (mc rune, comb []rune, style paint.Style, width int) {
	o.Lock()
	defer o.Unlock()
	mc, comb, style, width = o.back.GetCell(x, y)
	return
}

func (o *COffScreen) drawCell(x, y int) int {
	mc, comb, style, width := o.back.GetCell(x, y)
	if !o.back.Dirty(x, y) {
		return width
	}
	if x >= o.physW || y >= o.physH || x < 0 || y < 0 {
		return width
	}
	sc := &o.front[(y*o.physW)+x]

	if style == paint.StyleDefault {
		style = o.style
	}
	sc.Style = style
	sc.Runes = append([]rune{mc}, comb...)

	// now emit runes - taking care to not overrun width with a
	// wide character, and to ensure that we emit exactly one regular
	// character followed up by any residual combing characters

	sc.Bytes = nil

	if x > o.physW-width {
		sc.Runes = []rune{' '}
		sc.Bytes = []byte{' '}
		return width
	}

	lBuf := make([]byte, 12)
	uBuf := make([]byte, 12)
	nOut := 0

	for _, r := range sc.Runes {

		l := utf8.EncodeRune(uBuf, r)

		nOut, _, _ = o.encoder.Transform(lBuf, uBuf[:l], true)

		if nOut == 0 || lBuf[0] == '\x1a' {

			// skip combining

			if subst, ok := o.fallback[r]; ok {
				sc.Bytes = append(sc.Bytes,
					[]byte(subst)...)

			} else if r >= ' ' && r <= '~' {
				sc.Bytes = append(sc.Bytes, byte(r))

			} else if sc.Bytes == nil {
				sc.Bytes = append(sc.Bytes, '?')
			}
		} else {
			sc.Bytes = append(sc.Bytes, lBuf[:nOut]...)
		}
	}
	o.back.SetDirty(x, y, false)
	return width
}

func (o *COffScreen) ShowCursor(x, y int) {
	o.cursorX, o.cursorY = x, y
	o.showCursor()
	o.Unlock()
}

func (o *COffScreen) HideCursor() {
	o.ShowCursor(-1, -1)
}

func (o *COffScreen) showCursor() {

	x, y := o.cursorX, o.cursorY
	if x < 0 || y < 0 || x >= o.physW || y >= o.physH {
		o.cursorVis = false
	} else {
		o.cursorVis = true
	}
}

func (o *COffScreen) hideCursor() {
	// does not update cursor position
	o.cursorVis = false
}

func (o *COffScreen) Show() {
	o.Lock()
	defer o.Unlock()
	o.resize()
	o.draw()
}

func (o *COffScreen) clearScreen() {
	// We emulate a hardware clear by filling with a specific pattern
	for i := range o.front {
		o.front[i].Style = o.fillStyle
		o.front[i].Runes = []rune{o.fillChar}
		o.front[i].Bytes = []byte{byte(o.fillChar)}
	}
	o.clear = false
}

func (o *COffScreen) draw() {
	o.hideCursor()
	if o.clear {
		o.clearScreen()
	}

	w, h := o.back.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			width := o.drawCell(x, y)
			x += width - 1
		}
	}
	o.showCursor()
}

func (o *COffScreen) EnableMouse(_ ...MouseFlags) {
	o.mouse = true
}

func (o *COffScreen) DisableMouse() {
	o.mouse = false
}

func (o *COffScreen) EnablePaste() {
	o.paste = true
}

func (o *COffScreen) DisablePaste() {
	o.paste = false
}

func (o *COffScreen) Size() (w, h int) {
	w, h = o.back.Size()
	return
}

func (o *COffScreen) resize() {
	w, h := o.physW, o.physH
	ow, oh := o.back.Size()
	if w != ow || h != oh {
		o.back.Resize(w, h)
		ev := NewEventResize(w, h)
		_ = o.PostEvent(ev)
	}
}

func (o *COffScreen) Colors() int {
	return 256
}

func (o *COffScreen) PollEvent() Event {
	select {
	case <-o.quit:
		return nil
	case ev := <-o.evCh:
		return ev
	}
}

func (o *COffScreen) PollEventChan() (next chan Event) {
	next <- o.PollEvent()
	return
}

func (o *COffScreen) PostEventWait(ev Event) {
	o.evCh <- ev
}

func (o *COffScreen) PostEvent(ev Event) error {
	select {
	case o.evCh <- ev:
		return nil
	default:
		return ErrEventQFull
	}
}

func (o *COffScreen) InjectMouse(x, y int, buttons ButtonMask, mod ModMask) {
	ev := NewEventMouse(x, y, buttons, mod)
	_ = o.PostEvent(ev)
}

func (o *COffScreen) InjectKey(key Key, r rune, mod ModMask) {
	ev := NewEventKey(key, r, mod)
	_ = o.PostEvent(ev)
}

func (o *COffScreen) InjectKeyBytes(b []byte) bool {
	failed := false

outer:
	for len(b) > 0 {
		if b[0] >= ' ' && b[0] <= 0x7F {
			// printable ASCII easy to deal with -- no encodings
			ev := NewEventKey(KeyRune, rune(b[0]), ModNone)
			_ = o.PostEvent(ev)
			b = b[1:]
			continue
		}

		if b[0] < 0x80 {
			mod := ModNone
			// No encodings start with low numbered values
			if Key(b[0]) >= KeyCtrlA && Key(b[0]) <= KeyCtrlZ {
				mod = ModCtrl
			}
			ev := NewEventKey(Key(b[0]), 0, mod)
			_ = o.PostEvent(ev)
			b = b[1:]
			continue
		}

		utfBytes := make([]byte, len(b)*4) // worst case
		for l := 1; l < len(b); l++ {
			o.decoder.Reset()
			nOut, nin, _ := o.decoder.Transform(utfBytes, b[:l], true)

			if nOut != 0 {
				r, _ := utf8.DecodeRune(utfBytes[:nOut])
				if r != utf8.RuneError {
					ev := NewEventKey(KeyRune, r, ModNone)
					_ = o.PostEvent(ev)
				}
				b = b[nin:]
				continue outer
			}
		}
		failed = true
		b = b[1:]
		continue
	}

	return !failed
}

func (o *COffScreen) Sync() {
	o.clear = true
	o.resize()
	o.back.Invalidate()
	o.draw()
	o.Unlock()
}

func (o *COffScreen) CharacterSet() string {
	return o.charset
}

func (o *COffScreen) SetSize(w, h int) {
	o.Lock()
	defer o.Unlock()
	nc := make([]OffscreenCell, w*h)
	for row := 0; row < h && row < o.physH; row++ {
		for col := 0; col < w && col < o.physW; col++ {
			nc[(row*w)+col] = o.front[(row*o.physW)+col]
		}
	}
	o.cursorX, o.cursorY = -1, -1
	o.physW, o.physH = w, h
	o.front = nc
	o.back.Resize(w, h)
}

func (o *COffScreen) GetContents() ([]OffscreenCell, int, int) {
	o.Lock()
	defer o.Unlock()
	cells, w, h := o.front, o.physW, o.physH
	return cells, w, h
}

func (o *COffScreen) GetCursor() (int, int, bool) {
	o.Lock()
	defer o.Unlock()
	x, y, vis := o.cursorX, o.cursorY, o.cursorVis
	return x, y, vis
}

func (o *COffScreen) RegisterRuneFallback(r rune, subst string) {
	o.Lock()
	defer o.Unlock()
	o.fallback[r] = subst
}

func (o *COffScreen) UnregisterRuneFallback(r rune) {
	o.Lock()
	defer o.Unlock()
	delete(o.fallback, r)
}

func (o *COffScreen) CanDisplay(r rune, checkFallbacks bool) bool {
	if enc := o.encoder; enc != nil {
		nb := make([]byte, 6)
		ob := make([]byte, 6)
		num := utf8.EncodeRune(ob, r)

		enc.Reset()
		dst, _, err := enc.Transform(nb, ob[:num], true)
		if dst != 0 && err == nil && nb[0] != '\x1A' {
			return true
		}
	}
	if !checkFallbacks {
		return false
	}
	if _, ok := o.fallback[r]; ok {
		return true
	}
	return false
}

func (o *COffScreen) HasMouse() bool {
	return false
}

func (o *COffScreen) Resize(int, int, int, int) {}

func (o *COffScreen) HasKey(Key) bool {
	return true
}

func (o *COffScreen) Beep() error {
	return nil
}

func (o *COffScreen) Export() *CellBuffer {
	o.Lock()
	defer o.Unlock()
	cb := NewCellBuffer()
	w, h := o.back.Size()
	cb.Resize(w, h)
	for idx, cell := range o.back.cells {
		cb.cells[idx] = cell
	}
	return cb
}

func (o *COffScreen) Import(cb *CellBuffer) {
	o.Lock()
	defer o.Unlock()
	w, h := cb.Size()
	o.back.Resize(w, h)
	for idx, cell := range cb.cells {
		o.back.cells[idx] = cell
	}
}

func (o *COffScreen) CopyToClipboard(s string) {
	log.WarnF("unimplemented for offscreens")
}