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
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/jackdoe/go-gpmctl"
	"golang.org/x/text/transform"

	"github.com/go-curses/term"
	"github.com/go-curses/terminfo"
	_ "github.com/go-curses/terminfo/base"

	"github.com/go-curses/cdk/charset"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
)

// Screen represents the physical (or emulated) display.
// This can be a terminal window or a physical console.  Platforms implement
// this differently.
type Screen interface {
	// Init initializes the screen for use.
	Init() error
	InitWithFilePath(fp string) error
	InitWithFileHandle(fh *os.File) error

	TtyKeepFileHandle(keeping bool)
	TtyKeepingFileHandle() (keeping bool)
	TtyCloseWithStiRead(enabled bool)
	GetTtyCloseWithStiRead() (enabled bool)

	// Close finalizes the screen also releasing resources.
	Close()

	// Clear erases the screen.  The contents of any screen buffers
	// will also be cleared.  This has the logical effect of
	// filling the screen with spaces, using the global default style.
	Clear()

	// Fill fills the screen with the given character and style.
	Fill(rune, paint.Style)

	// SetCell is an older API, and will be removed.  Please use
	// SetContent instead; SetCell is implemented in terms of SetContent.
	SetCell(x int, y int, style paint.Style, ch ...rune)

	// GetContent returns the contents at the given location.  If the
	// coordinates are out of range, then the values will be 0, nil,
	// StyleDefault.  Note that the contents returned are logical contents
	// and may not actually be what is displayed, but rather are what will
	// be displayed if Show() or Sync() is called.  The width is the width
	// in screen cells; most often this will be 1, but some East Asian
	// characters require two cells.
	GetContent(x, y int) (mainc rune, combc []rune, style paint.Style, width int)

	// SetContent sets the contents of the given cell location.  If
	// the coordinates are out of range, then the operation is ignored.
	//
	// The first rune is the primary non-zero width rune.  The array
	// that follows is a possible list of combining characters to append,
	// and will usually be nil (no combining characters.)
	//
	// The results are not displayed until Show() or Sync() is called.
	//
	// Note that wide (East Asian full width) runes occupy two cells,
	// and attempts to place character at next cell to the right will have
	// undefined effects.  Wide runes that are printed in the
	// last column will be replaced with a single width space on output.
	SetContent(x int, y int, mainc rune, combc []rune, style paint.Style)

	// SetStyle sets the default style to use when clearing the screen
	// or when StyleDefault is specified.  If it is also StyleDefault,
	// then whatever system/terminal default is relevant will be used.
	SetStyle(style paint.Style)

	// ShowCursor is used to display the cursor at a given location.
	// If the coordinates -1, -1 are given or are otherwise outside the
	// dimensions of the screen, the cursor will be hidden.
	ShowCursor(x int, y int)

	// HideCursor is used to hide the cursor.  It's an alias for
	// ShowCursor(-1, -1).
	HideCursor()

	// Size returns the screen size as width, height.  This changes in
	// response to a call to Clear or Flush.
	Size() (w, h int)

	// PollEvent waits for events to arrive.  Main application loops
	// must spin on this to prevent the application from stalling.
	// Furthermore, this will return nil if the Screen is finalized.
	PollEvent() Event

	// PollEventChan provides a PollEvent() call response through a channel
	// for the purposes of using a select statement to poll for new events
	PollEventChan() (next chan Event)

	// PostEvent tries to post an event into the event stream.  This
	// can fail if the event queue is full.  In that case, the event
	// is dropped, and ErrEventQFull is returned.
	PostEvent(ev Event) error

	// EnableMouse enables the mouse.  (If your terminal supports it.)
	// If no flags are specified, then all events are reported, if the
	// terminal supports them.
	EnableMouse(...MouseFlags)

	// DisableMouse disables the mouse.
	DisableMouse()

	// EnablePaste enables bracketed paste mode, if supported.
	EnablePaste()

	// DisablePaste disables bracketed paste mode.
	DisablePaste()

	// HasMouse returns true if the terminal (apparently) supports a
	// mouse.  Note that the a return value of true doesn't guarantee that
	// a mouse/pointing device is present; a false return definitely
	// indicates no mouse support is available.
	HasMouse() bool

	// Colors returns the number of colors.  All colors are assumed to
	// use the ANSI color map.  If a terminal is monochrome, it will
	// return 0.
	Colors() int

	// Show makes all the content changes made using SetContent() visible
	// on the screen.
	//
	// It does so in the most efficient and least visually disruptive
	// manner possible.
	Show()

	// Sync works like Show(), but it updates every visible cell on the
	// physical screen, assuming that it is not synchronized with any
	// internal model.  This may be both expensive and visually jarring,
	// so it should only be used when believed to actually be necessary.
	//
	// Typically, this is called as a result of a user-requested redraw
	// (e.g. to clear up on screen corruption caused by some other program),
	// or during a resize event.
	Sync()

	// CharacterSet returns information about the character set.
	// This isn't the full locale, but it does give us the input/output
	// character set.  Note that this is just for diagnostic purposes,
	// we normally translate input/output to/from UTF-8, regardless of
	// what the user's environment is.
	CharacterSet() string

	// RegisterRuneFallback adds a fallback for runes that are not
	// part of the character set -- for example one could register
	// o as a fallback for ø.  This should be done cautiously for
	// characters that might be displayed ordinarily in language
	// specific text -- characters that could change the meaning of
	// of written text would be dangerous.  The intention here is to
	// facilitate fallback characters in pseudo-graphical applications.
	//
	// If the terminal has fallbacks already in place via an alternate
	// character set, those are used in preference.  Also, standard
	// fallbacks for graphical characters in the ACSC terminfo string
	// are registered implicitly.
	//
	// The display string should be the same width as original rune.
	// This makes it possible to register two character replacements
	// for full width East Asian characters, for example.
	//
	// It is recommended that replacement strings consist only of
	// 7-bit ASCII, since other characters may not display everywhere.
	RegisterRuneFallback(r rune, subst string)

	// UnregisterRuneFallback unmaps a replacement.  It will unmap
	// the implicit ASCII replacements for alternate characters as well.
	// When an unmapped char needs to be displayed, but no suitable
	// glyph is available, '?' is emitted instead.  It is not possible
	// to "disable" the use of alternate characters that are supported
	// by your terminal except by changing the terminal database.
	UnregisterRuneFallback(r rune)

	// CanDisplay returns true if the given rune can be displayed on
	// this screen.  Note that this is a best guess effort -- whether
	// your fonts support the character or not may be questionable.
	// Mostly this is for folks who work outside of Unicode.
	//
	// If checkFallbacks is true, then if any (possibly imperfect)
	// fallbacks are registered, this will return true.  This will
	// also return true if the terminal can replace the glyph with
	// one that is visually indistinguishable from the one requested.
	CanDisplay(r rune, checkFallbacks bool) bool

	// HasKey returns true if the keyboard is believed to have the
	// key.  In some cases a keyboard may have keys with this name
	// but no support for them, while in others a key may be reported
	// as supported but not actually be usable (such as some emulators
	// that hijack certain keys).  Its best not to depend to strictly
	// on this function, but it can be used for hinting when building
	// menus, displayed hot-keys, etc.  Note that KeyRune (literal
	// runes) is always true.
	HasKey(Key) bool

	// Beep attempts to sound an OS-dependent audible alert and returns an error
	// when unsuccessful.
	Beep() error

	Export() *CellBuffer
	Import(cb *CellBuffer)

	HostClipboardEnabled() (enabled bool)
	CopyToClipboard(s string)
	PasteFromClipboard() (s string, ok bool)
	EnableHostClipboard(enabled bool)
	EnableTermClipboard(enabled bool)
}

var (
	EventQueueSize    = 1024
	EventKeyQueueSize = 1024
	EventKeyTiming    = time.Millisecond * 50
	SignalQueueSize   = 100
)

// NewScreen returns a Screen that uses the stock TTY interface
// and POSIX terminal control, combined with a terminfo description taken from
// the $TERM environment variable.  It returns an error if the terminal
// is not supported for any reason.
//
// For terminals that do not support dynamic resize events, the $LINES
// $COLUMNS environment variables can be set to the actual window size,
// otherwise defaults taken from the terminal database are used.
func NewScreen() (Screen, error) {
	ti, e := terminfo.LookupTerminfo(os.Getenv("TERM"))
	if e != nil {
		ti, e = loadDynamicTerminfo(os.Getenv("TERM"))
		if e != nil {
			return nil, e
		}
		terminfo.AddTerminfo(ti)
	}
	t := &CScreen{
		ti:          ti,
		ttyPath:     "/dev/tty",
		ttyReadLock: &sync.Mutex{},
	}

	t.keyExist = make(map[Key]bool)
	t.keyCodes = make(map[string]*tKeyCode)
	if len(ti.Mouse) > 0 {
		t.mouse = []byte(ti.Mouse)
	}
	t.prepareKeys()
	t.buildAcsMap()
	t.sigWinch = make(chan os.Signal, SignalQueueSize)
	t.fallback = make(map[rune]string)
	for k, v := range paint.RuneFallbacks {
		t.fallback[k] = v
	}

	return t, nil
}

// tKeyCode represents a combination of a key code and modifiers.
type tKeyCode struct {
	key Key
	mod ModMask
}

// CScreen represents a screen backed by a terminfo implementation.
type CScreen struct {
	ttyPath      string
	ttyFile      *os.File
	ttyKeepFH    bool        // do not close the file handle
	ttyReadSti   bool        // inject " " to cancel term.Read
	ttyReading   bool        // is currently waiting for a term.Read
	ttyReadLock  *sync.Mutex // thread-safe term.Read tracking
	ti           *terminfo.Terminfo
	h            int
	w            int
	finished     bool
	cells        *CellBuffer
	term         *term.Term
	buffering    bool // true if we are collecting writes to buf instead of sending directly to out
	buf          bytes.Buffer
	curStyle     paint.Style
	style        paint.Style
	evCh         chan Event
	sigWinch     chan os.Signal
	quit         chan struct{}
	inDoneQ      chan struct{}
	keyExist     map[Key]bool
	keyCodes     map[string]*tKeyCode
	keyChan      chan []byte
	keyTimer     *time.Timer
	keyExpire    time.Time
	cx           int
	cy           int
	mouse        []byte
	clear        bool
	cursorX      int
	cursorY      int
	wasBtn       bool
	acs          map[rune]string
	charset      string
	encoder      transform.Transformer
	decoder      transform.Transformer
	fallback     map[rune]string
	colors       map[paint.Color]paint.Color
	palette      []paint.Color
	trueColor    bool
	escaped      bool
	buttonDn     bool
	finishOnce   sync.Once
	enablePaste  string
	disablePaste string
	gpmRunning   bool

	useHostClipboard bool
	useTermClipboard bool
	sync.Mutex
}

func (d *CScreen) Init() error {
	return d.initReal()
}

func (d *CScreen) InitWithFilePath(fp string) error {
	d.ttyPath = fp
	d.ttyFile = nil
	return d.initReal()
}

func (d *CScreen) InitWithFileHandle(fh *os.File) error {
	d.ttyPath = ""
	d.ttyFile = fh
	return d.initReal()
}

func (d *CScreen) TtyKeepFileHandle(keep bool) {
	d.ttyKeepFH = keep
}

func (d *CScreen) TtyKeepingFileHandle() (keeping bool) {
	keeping = d.ttyKeepFH
	return
}

func (d *CScreen) TtyCloseWithStiRead(enabled bool) {
	d.ttyReadSti = enabled
}

func (d *CScreen) GetTtyCloseWithStiRead() (enabled bool) {
	enabled = d.ttyReadSti
	return
}

func (d *CScreen) initReal() error {
	d.useHostClipboard = false
	d.useTermClipboard = true
	d.evCh = make(chan Event, EventQueueSize)
	d.inDoneQ = make(chan struct{})
	d.keyChan = make(chan []byte, EventKeyQueueSize)
	d.keyTimer = time.NewTimer(EventKeyTiming)
	d.cells = NewCellBuffer()

	d.charset = charset.Get()
	if enc := GetEncoding(d.charset); enc != nil {
		d.encoder = enc.NewEncoder()
		d.decoder = enc.NewDecoder()
	} else {
		return ErrNoCharset
	}
	ti := d.ti

	// environment overrides
	w := ti.Columns
	h := ti.Lines
	if i, _ := strconv.Atoi(os.Getenv("LINES")); i != 0 {
		h = i
	}
	if i, _ := strconv.Atoi(os.Getenv("COLUMNS")); i != 0 {
		w = i
	}
	if e := d.initialize(); e != nil {
		return e
	}

	if d.ti.SetFgBgRGB != "" || d.ti.SetFgRGB != "" || d.ti.SetBgRGB != "" {
		d.trueColor = true
	}
	// A user who wants to have his themes honored can
	// set this environment variable.
	if os.Getenv("GO_CDK_TRUECOLOR") == "disable" {
		d.trueColor = false
	}
	d.colors = make(map[paint.Color]paint.Color)
	d.palette = make([]paint.Color, d.nColors())
	for i := 0; i < d.nColors(); i++ {
		d.palette[i] = paint.Color(i) | paint.ColorValid
		// identity map for our builtin colors
		d.colors[paint.Color(i)|paint.ColorValid] = paint.Color(i) | paint.ColorValid
	}

	d.TPuts(ti.EnterCA)
	d.TPuts(ti.HideCursor)
	d.TPuts(ti.EnableAcs)
	d.TPuts(ti.Clear)

	d.quit = make(chan struct{})

	d.Lock()
	d.cx = -1
	d.cy = -1
	d.style = paint.StyleDefault
	d.cells.Resize(w, h)
	d.cursorX = -1
	d.cursorY = -1
	d.resize()
	d.Unlock()

	go d.mainLoop()
	go d.inputLoop()

	return nil
}

func (d *CScreen) prepareKeyMod(key Key, mod ModMask, val string) {
	if val != "" {
		// Do not override codes that already exist
		if _, exist := d.keyCodes[val]; !exist {
			d.keyExist[key] = true
			d.keyCodes[val] = &tKeyCode{key: key, mod: mod}
		}
	}
}

func (d *CScreen) prepareKeyModReplace(key Key, replace Key, mod ModMask, val string) {
	if val != "" {
		// Do not override codes that already exist
		if old, exist := d.keyCodes[val]; !exist || old.key == replace {
			d.keyExist[key] = true
			d.keyCodes[val] = &tKeyCode{key: key, mod: mod}
		}
	}
}

func (d *CScreen) prepareKeyModXTerm(key Key, val string) {

	if strings.HasPrefix(val, "\x1b[") && strings.HasSuffix(val, "~") {

		// Drop the trailing ~
		val = val[:len(val)-1]

		// These suffixes are calculated assuming Xterm style modifier suffixes.
		// Please see https://invisible-island.net/xterm/ctlseqs/ctlseqs.pdf for
		// more information (specifically "PC-Style Function Keys").
		d.prepareKeyModReplace(key, key+12, ModShift, val+";2~")
		d.prepareKeyModReplace(key, key+48, ModAlt, val+";3~")
		d.prepareKeyModReplace(key, key+60, ModAlt|ModShift, val+";4~")
		d.prepareKeyModReplace(key, key+24, ModCtrl, val+";5~")
		d.prepareKeyModReplace(key, key+36, ModCtrl|ModShift, val+";6~")
		d.prepareKeyMod(key, ModAlt|ModCtrl, val+";7~")
		d.prepareKeyMod(key, ModShift|ModAlt|ModCtrl, val+";8~")
		d.prepareKeyMod(key, ModMeta, val+";9~")
		d.prepareKeyMod(key, ModMeta|ModShift, val+";10~")
		d.prepareKeyMod(key, ModMeta|ModAlt, val+";11~")
		d.prepareKeyMod(key, ModMeta|ModAlt|ModShift, val+";12~")
		d.prepareKeyMod(key, ModMeta|ModCtrl, val+";13~")
		d.prepareKeyMod(key, ModMeta|ModCtrl|ModShift, val+";14~")
		d.prepareKeyMod(key, ModMeta|ModCtrl|ModAlt, val+";15~")
		d.prepareKeyMod(key, ModMeta|ModCtrl|ModAlt|ModShift, val+";16~")
	} else if strings.HasPrefix(val, "\x1bO") && len(val) == 3 {
		val = val[2:]
		d.prepareKeyModReplace(key, key+12, ModShift, "\x1b[1;2"+val)
		d.prepareKeyModReplace(key, key+48, ModAlt, "\x1b[1;3"+val)
		d.prepareKeyModReplace(key, key+24, ModCtrl, "\x1b[1;5"+val)
		d.prepareKeyModReplace(key, key+36, ModCtrl|ModShift, "\x1b[1;6"+val)
		d.prepareKeyModReplace(key, key+60, ModAlt|ModShift, "\x1b[1;4"+val)
		d.prepareKeyMod(key, ModAlt|ModCtrl, "\x1b[1;7"+val)
		d.prepareKeyMod(key, ModShift|ModAlt|ModCtrl, "\x1b[1;8"+val)
		d.prepareKeyMod(key, ModMeta, "\x1b[1;9"+val)
		d.prepareKeyMod(key, ModMeta|ModShift, "\x1b[1;10"+val)
		d.prepareKeyMod(key, ModMeta|ModAlt, "\x1b[1;11"+val)
		d.prepareKeyMod(key, ModMeta|ModAlt|ModShift, "\x1b[1;12"+val)
		d.prepareKeyMod(key, ModMeta|ModCtrl, "\x1b[1;13"+val)
		d.prepareKeyMod(key, ModMeta|ModCtrl|ModShift, "\x1b[1;14"+val)
		d.prepareKeyMod(key, ModMeta|ModCtrl|ModAlt, "\x1b[1;15"+val)
		d.prepareKeyMod(key, ModMeta|ModCtrl|ModAlt|ModShift, "\x1b[1;16"+val)
	}
}

func (d *CScreen) prepareXtermModifiers() {
	if d.ti.Modifiers != terminfo.ModifiersXTerm {
		return
	}
	d.prepareKeyModXTerm(KeyRight, d.ti.KeyRight)
	d.prepareKeyModXTerm(KeyLeft, d.ti.KeyLeft)
	d.prepareKeyModXTerm(KeyUp, d.ti.KeyUp)
	d.prepareKeyModXTerm(KeyDown, d.ti.KeyDown)
	d.prepareKeyModXTerm(KeyInsert, d.ti.KeyInsert)
	d.prepareKeyModXTerm(KeyDelete, d.ti.KeyDelete)
	d.prepareKeyModXTerm(KeyPgUp, d.ti.KeyPgUp)
	d.prepareKeyModXTerm(KeyPgDn, d.ti.KeyPgDn)
	d.prepareKeyModXTerm(KeyHome, d.ti.KeyHome)
	d.prepareKeyModXTerm(KeyEnd, d.ti.KeyEnd)
	d.prepareKeyModXTerm(KeyF1, d.ti.KeyF1)
	d.prepareKeyModXTerm(KeyF2, d.ti.KeyF2)
	d.prepareKeyModXTerm(KeyF3, d.ti.KeyF3)
	d.prepareKeyModXTerm(KeyF4, d.ti.KeyF4)
	d.prepareKeyModXTerm(KeyF5, d.ti.KeyF5)
	d.prepareKeyModXTerm(KeyF6, d.ti.KeyF6)
	d.prepareKeyModXTerm(KeyF7, d.ti.KeyF7)
	d.prepareKeyModXTerm(KeyF8, d.ti.KeyF8)
	d.prepareKeyModXTerm(KeyF9, d.ti.KeyF9)
	d.prepareKeyModXTerm(KeyF10, d.ti.KeyF10)
	d.prepareKeyModXTerm(KeyF11, d.ti.KeyF11)
	d.prepareKeyModXTerm(KeyF12, d.ti.KeyF12)
}

func (d *CScreen) prepareBracketedPaste() {
	// Another workaround for lack of reporting in terminfo.
	// We assume if the terminal has a mouse entry, that it
	// offers bracketed paste.  But we allow specific overrides
	// via our terminal database.
	if d.ti.EnablePaste != "" {
		d.enablePaste = d.ti.EnablePaste
		d.disablePaste = d.ti.DisablePaste
		d.prepareKey(keyPasteStart, d.ti.PasteStart)
		d.prepareKey(keyPasteEnd, d.ti.PasteEnd)
	} else if d.ti.Mouse != "" {
		d.enablePaste = "\x1b[?2004h"
		d.disablePaste = "\x1b[?2004l"
		d.prepareKey(keyPasteStart, "\x1b[200~")
		d.prepareKey(keyPasteEnd, "\x1b[201~")
	}
}

func (d *CScreen) prepareKey(key Key, val string) {
	d.prepareKeyMod(key, ModNone, val)
}

func (d *CScreen) prepareKeys() {
	ti := d.ti
	d.prepareKey(KeyBackspace, ti.KeyBackspace)
	d.prepareKey(KeyF1, ti.KeyF1)
	d.prepareKey(KeyF2, ti.KeyF2)
	d.prepareKey(KeyF3, ti.KeyF3)
	d.prepareKey(KeyF4, ti.KeyF4)
	d.prepareKey(KeyF5, ti.KeyF5)
	d.prepareKey(KeyF6, ti.KeyF6)
	d.prepareKey(KeyF7, ti.KeyF7)
	d.prepareKey(KeyF8, ti.KeyF8)
	d.prepareKey(KeyF9, ti.KeyF9)
	d.prepareKey(KeyF10, ti.KeyF10)
	d.prepareKey(KeyF11, ti.KeyF11)
	d.prepareKey(KeyF12, ti.KeyF12)
	d.prepareKey(KeyF13, ti.KeyF13)
	d.prepareKey(KeyF14, ti.KeyF14)
	d.prepareKey(KeyF15, ti.KeyF15)
	d.prepareKey(KeyF16, ti.KeyF16)
	d.prepareKey(KeyF17, ti.KeyF17)
	d.prepareKey(KeyF18, ti.KeyF18)
	d.prepareKey(KeyF19, ti.KeyF19)
	d.prepareKey(KeyF20, ti.KeyF20)
	d.prepareKey(KeyF21, ti.KeyF21)
	d.prepareKey(KeyF22, ti.KeyF22)
	d.prepareKey(KeyF23, ti.KeyF23)
	d.prepareKey(KeyF24, ti.KeyF24)
	d.prepareKey(KeyF25, ti.KeyF25)
	d.prepareKey(KeyF26, ti.KeyF26)
	d.prepareKey(KeyF27, ti.KeyF27)
	d.prepareKey(KeyF28, ti.KeyF28)
	d.prepareKey(KeyF29, ti.KeyF29)
	d.prepareKey(KeyF30, ti.KeyF30)
	d.prepareKey(KeyF31, ti.KeyF31)
	d.prepareKey(KeyF32, ti.KeyF32)
	d.prepareKey(KeyF33, ti.KeyF33)
	d.prepareKey(KeyF34, ti.KeyF34)
	d.prepareKey(KeyF35, ti.KeyF35)
	d.prepareKey(KeyF36, ti.KeyF36)
	d.prepareKey(KeyF37, ti.KeyF37)
	d.prepareKey(KeyF38, ti.KeyF38)
	d.prepareKey(KeyF39, ti.KeyF39)
	d.prepareKey(KeyF40, ti.KeyF40)
	d.prepareKey(KeyF41, ti.KeyF41)
	d.prepareKey(KeyF42, ti.KeyF42)
	d.prepareKey(KeyF43, ti.KeyF43)
	d.prepareKey(KeyF44, ti.KeyF44)
	d.prepareKey(KeyF45, ti.KeyF45)
	d.prepareKey(KeyF46, ti.KeyF46)
	d.prepareKey(KeyF47, ti.KeyF47)
	d.prepareKey(KeyF48, ti.KeyF48)
	d.prepareKey(KeyF49, ti.KeyF49)
	d.prepareKey(KeyF50, ti.KeyF50)
	d.prepareKey(KeyF51, ti.KeyF51)
	d.prepareKey(KeyF52, ti.KeyF52)
	d.prepareKey(KeyF53, ti.KeyF53)
	d.prepareKey(KeyF54, ti.KeyF54)
	d.prepareKey(KeyF55, ti.KeyF55)
	d.prepareKey(KeyF56, ti.KeyF56)
	d.prepareKey(KeyF57, ti.KeyF57)
	d.prepareKey(KeyF58, ti.KeyF58)
	d.prepareKey(KeyF59, ti.KeyF59)
	d.prepareKey(KeyF60, ti.KeyF60)
	d.prepareKey(KeyF61, ti.KeyF61)
	d.prepareKey(KeyF62, ti.KeyF62)
	d.prepareKey(KeyF63, ti.KeyF63)
	d.prepareKey(KeyF64, ti.KeyF64)
	d.prepareKey(KeyInsert, ti.KeyInsert)
	d.prepareKey(KeyDelete, ti.KeyDelete)
	d.prepareKey(KeyHome, ti.KeyHome)
	d.prepareKey(KeyEnd, ti.KeyEnd)
	d.prepareKey(KeyUp, ti.KeyUp)
	d.prepareKey(KeyDown, ti.KeyDown)
	d.prepareKey(KeyLeft, ti.KeyLeft)
	d.prepareKey(KeyRight, ti.KeyRight)
	d.prepareKey(KeyPgUp, ti.KeyPgUp)
	d.prepareKey(KeyPgDn, ti.KeyPgDn)
	d.prepareKey(KeyHelp, ti.KeyHelp)
	d.prepareKey(KeyPrint, ti.KeyPrint)
	d.prepareKey(KeyCancel, ti.KeyCancel)
	d.prepareKey(KeyExit, ti.KeyExit)
	d.prepareKey(KeyBacktab, ti.KeyBacktab)

	d.prepareKeyMod(KeyRight, ModShift, ti.KeyShfRight)
	d.prepareKeyMod(KeyLeft, ModShift, ti.KeyShfLeft)
	d.prepareKeyMod(KeyUp, ModShift, ti.KeyShfUp)
	d.prepareKeyMod(KeyDown, ModShift, ti.KeyShfDown)
	d.prepareKeyMod(KeyHome, ModShift, ti.KeyShfHome)
	d.prepareKeyMod(KeyEnd, ModShift, ti.KeyShfEnd)
	d.prepareKeyMod(KeyPgUp, ModShift, ti.KeyShfPgUp)
	d.prepareKeyMod(KeyPgDn, ModShift, ti.KeyShfPgDn)

	d.prepareKeyMod(KeyRight, ModCtrl, ti.KeyCtrlRight)
	d.prepareKeyMod(KeyLeft, ModCtrl, ti.KeyCtrlLeft)
	d.prepareKeyMod(KeyUp, ModCtrl, ti.KeyCtrlUp)
	d.prepareKeyMod(KeyDown, ModCtrl, ti.KeyCtrlDown)
	d.prepareKeyMod(KeyHome, ModCtrl, ti.KeyCtrlHome)
	d.prepareKeyMod(KeyEnd, ModCtrl, ti.KeyCtrlEnd)

	// Sadly, xterm handling of keycodes is somewhat erratic.  In
	// particular, different codes are sent depending on application
	// mode is in use or not, and the entries for many of these are
	// simply absent from terminfo on many systems.  So we insert
	// a number of escape sequences if they are not already used, in
	// order to have the widest correct usage.  Note that prepareKey
	// will not inject codes if the escape sequence is already known.
	// We also only do this for terminals that have the application
	// mode present.

	// Cursor mode
	if ti.EnterKeypad != "" {
		d.prepareKey(KeyUp, "\x1b[A")
		d.prepareKey(KeyDown, "\x1b[B")
		d.prepareKey(KeyRight, "\x1b[C")
		d.prepareKey(KeyLeft, "\x1b[D")
		d.prepareKey(KeyEnd, "\x1b[F")
		d.prepareKey(KeyHome, "\x1b[H")
		d.prepareKey(KeyDelete, "\x1b[3~")
		d.prepareKey(KeyHome, "\x1b[1~")
		d.prepareKey(KeyEnd, "\x1b[4~")
		d.prepareKey(KeyPgUp, "\x1b[5~")
		d.prepareKey(KeyPgDn, "\x1b[6~")

		// Application mode
		d.prepareKey(KeyUp, "\x1bOA")
		d.prepareKey(KeyDown, "\x1bOB")
		d.prepareKey(KeyRight, "\x1bOC")
		d.prepareKey(KeyLeft, "\x1bOD")
		d.prepareKey(KeyHome, "\x1bOH")
	}

	d.prepareKey(keyPasteStart, ti.PasteStart)
	d.prepareKey(keyPasteEnd, ti.PasteEnd)
	d.prepareXtermModifiers()
	d.prepareBracketedPaste()

outer:
	// Add key mappings for control keys.
	for i := 0; i < ' '; i++ {
		// Do not insert direct key codes for ambiguous keys.
		// For example, ESC is used for lots of other keys, so
		// when parsing this we don't want to fast path handling
		// of it, but instead wait a bit before parsing it as in
		// isolation.
		for esc := range d.keyCodes {
			if []byte(esc)[0] == byte(i) {
				continue outer
			}
		}

		d.keyExist[Key(i)] = true

		mod := ModCtrl
		switch Key(i) {
		case KeyBS, KeyTAB, KeyESC, KeyCR:
			// directly type-able- no control sequence
			mod = ModNone
		}
		d.keyCodes[string(rune(i))] = &tKeyCode{key: Key(i), mod: mod}
	}
}

func (d *CScreen) Close() {
	d.finishOnce.Do(d.finish)
}

func (d *CScreen) finish() {
	d.Lock()
	defer d.Unlock()

	ti := d.ti
	d.cells.Resize(0, 0)
	d.TPuts(ti.ShowCursor)
	d.TPuts(ti.AttrOff)
	d.TPuts(ti.Clear)
	d.TPuts(ti.ExitCA)
	d.TPuts(ti.ExitKeypad)
	d.TPuts(d.disablePaste)
	d.DisableMouse()
	d.curStyle = paint.StyleInvalid
	d.clear = false
	d.finished = true

	select {
	case <-d.quit:
	default:
		close(d.quit)
	}

	d.finalize()
}

func (d *CScreen) SetStyle(style paint.Style) {
	d.Lock()
	if !d.finished {
		d.style = style
	}
	d.Unlock()
}

func (d *CScreen) Clear() {
	d.Fill(' ', d.style)
}

func (d *CScreen) Fill(r rune, style paint.Style) {
	d.Lock()
	if !d.finished {
		d.cells.Fill(r, style)
	}
	d.Unlock()
}

func (d *CScreen) SetContent(x, y int, mc rune, comb []rune, style paint.Style) {
	d.Lock()
	if !d.finished {
		d.cells.SetCell(x, y, mc, comb, style)
	}
	d.Unlock()
}

func (d *CScreen) GetContent(x, y int) (rune, []rune, paint.Style, int) {
	d.Lock()
	mc, comb, style, width := d.cells.GetCell(x, y)
	d.Unlock()
	return mc, comb, style, width
}

func (d *CScreen) SetCell(x, y int, style paint.Style, ch ...rune) {
	if len(ch) > 0 {
		d.SetContent(x, y, ch[0], ch[1:], style)
	} else {
		d.SetContent(x, y, ' ', nil, style)
	}
}

func (d *CScreen) encodeRune(r rune, buf []byte) []byte {

	nb := make([]byte, 6)
	ob := make([]byte, 6)
	num := utf8.EncodeRune(ob, r)
	ob = ob[:num]
	dst := 0
	var err error
	if enc := d.encoder; enc != nil {
		enc.Reset()
		dst, _, err = enc.Transform(nb, ob, true)
	}
	if err != nil || dst == 0 || nb[0] == '\x1a' {
		// Combining characters are elided
		if len(buf) == 0 {
			if acs, ok := d.acs[r]; ok {
				buf = append(buf, []byte(acs)...)
			} else if fb, ok := d.fallback[r]; ok {
				buf = append(buf, []byte(fb)...)
			} else {
				buf = append(buf, '?')
			}
		}
	} else {
		buf = append(buf, nb[:dst]...)
	}

	return buf
}

func (d *CScreen) sendFgBg(fg paint.Color, bg paint.Color) {
	ti := d.ti
	if ti.Colors == 0 {
		return
	}
	if fg == paint.ColorReset || bg == paint.ColorReset {
		d.TPuts(ti.ResetFgBg)
	}
	if d.trueColor {
		if ti.SetFgBgRGB != "" && fg.IsRGB() && bg.IsRGB() {
			r1, g1, b1 := fg.RGB()
			r2, g2, b2 := bg.RGB()
			d.TPuts(ti.TParm(ti.SetFgBgRGB,
				int(r1), int(g1), int(b1),
				int(r2), int(g2), int(b2)))
			return
		}

		if fg.IsRGB() && ti.SetFgRGB != "" {
			r, g, b := fg.RGB()
			d.TPuts(ti.TParm(ti.SetFgRGB, int(r), int(g), int(b)))
			fg = paint.ColorDefault
		}

		if bg.IsRGB() && ti.SetBgRGB != "" {
			r, g, b := bg.RGB()
			d.TPuts(ti.TParm(ti.SetBgRGB,
				int(r), int(g), int(b)))
			bg = paint.ColorDefault
		}
	}

	if fg.Valid() {
		if v, ok := d.colors[fg]; ok {
			fg = v
		} else {
			v = paint.FindColor(fg, d.palette)
			d.colors[fg] = v
			fg = v
		}
	}

	if bg.Valid() {
		if v, ok := d.colors[bg]; ok {
			bg = v
		} else {
			v = paint.FindColor(bg, d.palette)
			d.colors[bg] = v
			bg = v
		}
	}

	if fg.Valid() && bg.Valid() && ti.SetFgBg != "" {
		d.TPuts(ti.TParm(ti.SetFgBg, int(fg&0xff), int(bg&0xff)))
	} else {
		if fg.Valid() && ti.SetFg != "" {
			d.TPuts(ti.TParm(ti.SetFg, int(fg&0xff)))
		}
		if bg.Valid() && ti.SetBg != "" {
			d.TPuts(ti.TParm(ti.SetBg, int(bg&0xff)))
		}
	}
}

func (d *CScreen) drawCell(x, y int) int {

	ti := d.ti

	mc, comb, style, width := d.cells.GetCell(x, y)
	if !d.cells.Dirty(x, y) {
		return width
	}

	if d.cy != y || d.cx != x {
		d.TPuts(ti.TGoto(x, y))
		d.cx = x
		d.cy = y
	}

	if style == paint.StyleDefault {
		style = d.style
	}
	if style != d.curStyle {
		fg, bg, attrs := style.Decompose()

		d.TPuts(ti.AttrOff)

		d.sendFgBg(fg, bg)
		if attrs&paint.AttrBold != 0 {
			d.TPuts(ti.Bold)
		}
		if attrs&paint.AttrUnderline != 0 {
			d.TPuts(ti.Underline)
		}
		if attrs&paint.AttrReverse != 0 {
			d.TPuts(ti.Reverse)
		}
		if attrs&paint.AttrBlink != 0 {
			d.TPuts(ti.Blink)
		}
		if attrs&paint.AttrDim != 0 {
			d.TPuts(ti.Dim)
		}
		if attrs&paint.AttrItalic != 0 {
			d.TPuts(ti.Italic)
		}
		if attrs&paint.AttrStrike != 0 {
			d.TPuts(ti.StrikeThrough)
		}
		d.curStyle = style
	}
	// now emit runes - taking care to not overrun width with a
	// wide character, and to ensure that we emit exactly one regular
	// character followed up by any residual combing characters

	if width < 1 {
		width = 1
	}

	var str string

	buf := make([]byte, 0, 6)

	buf = d.encodeRune(mc, buf)
	for _, r := range comb {
		buf = d.encodeRune(r, buf)
	}

	str = string(buf)
	if width > 1 && str == "?" {
		// No FullWidth character support
		str = "? "
		d.cx = -1
	}

	if x > d.w-width {
		// too wide to fit; emit a single space instead
		width = 1
		str = " "
	}
	d.writeString(str)
	d.cx += width
	d.cells.SetDirty(x, y, false)
	if width > 1 {
		d.cx = -1
	}

	return width
}

func (d *CScreen) ShowCursor(x, y int) {
	d.Lock()
	d.cursorX = x
	d.cursorY = y
	d.Unlock()
}

func (d *CScreen) HideCursor() {
	d.ShowCursor(-1, -1)
}

func (d *CScreen) showCursor() {

	x, y := d.cursorX, d.cursorY
	w, h := d.cells.Size()
	if x < 0 || y < 0 || x >= w || y >= h {
		d.hideCursor()
		return
	}
	d.TPuts(d.ti.TGoto(x, y))
	d.TPuts(d.ti.ShowCursor)
	d.cx = x
	d.cy = y
}

// writeString sends a string to the terminal. The string is sent as-is and
// this function does not expand inline padding indications (of the form
// $<[delay]> where [delay] is msec). In order to have these expanded, use
// TPuts. If the screen is "buffering", the string is collected in a buffer,
// with the intention that the entire buffer be sent to the terminal in one
// write operation at some point later.
func (d *CScreen) writeString(s string) {
	if d.buffering {
		_, _ = io.WriteString(&d.buf, s)
	} else {
		_, _ = d.term.Write([]byte(s))
	}
}

func (d *CScreen) TPuts(s string) {
	if d.buffering {
		d.ti.TPuts(&d.buf, s)
	} else {
		_, _ = d.term.Write([]byte(s))
	}
}

func (d *CScreen) Show() {
	d.Lock()
	if !d.finished {
		d.resize()
		d.draw()
	}
	d.Unlock()
}

func (d *CScreen) clearDisplay() {
	fg, bg, _ := d.style.Decompose()
	d.sendFgBg(fg, bg)
	d.TPuts(d.ti.Clear)
	d.clear = false
}

func (d *CScreen) hideCursor() {
	// does not update cursor position
	if d.ti.HideCursor != "" {
		d.TPuts(d.ti.HideCursor)
	} else {
		// No way to hide cursor, stick it
		// at bottom right of screen
		d.cx, d.cy = d.cells.Size()
		d.TPuts(d.ti.TGoto(d.cx, d.cy))
	}
}

func (d *CScreen) draw() {
	// clobber cursor position, because we're gonna change it all
	d.cx = -1
	d.cy = -1

	d.buf.Reset()
	d.buffering = true
	defer func() {
		d.buffering = false
	}()

	// hide the cursor while we move stuff around
	d.hideCursor()

	if d.clear {
		d.clearDisplay()
	}

	for y := 0; y < d.h; y++ {
		for x := 0; x < d.w; x++ {
			width := d.drawCell(x, y)
			if width > 1 {
				if x+1 < d.w {
					// this is necessary so that if we ever
					// go back to drawing that cell, we
					// actually will *draw* it.
					d.cells.SetDirty(x+1, y, true)
				}
			}
			x += width - 1
		}
	}

	// restore the cursor
	d.showCursor()

	_, _ = d.buf.WriteTo(d.term)
}

func (d *CScreen) EnableGPM() {
	if !d.gpmRunning {
		go d.gpmLoop()
	}
}

func (d *CScreen) EnableMouse(flags ...MouseFlags) {
	var f MouseFlags
	flagsPresent := false
	for _, flag := range flags {
		f |= flag
		flagsPresent = true
	}
	if !flagsPresent {
		f = MouseMotionEvents
	}

	// Rather than using terminfo to find mouse escape sequences, we rely on the fact that
	// pretty much *every* terminal that supports mouse tracking follows the
	// XTerm standards (the modern ones).

	if len(d.mouse) != 0 {
		var mm int
		if f&MouseMotionEvents != 0 {
			mm = 1003
		} else if f&MouseDragEvents != 0 {
			mm = 1002
		} else if f&MouseButtonEvents != 0 {
			mm = 1000
		} else {
			// No recognized tracking enabled.
			return
		}

		d.TPuts(fmt.Sprintf("\x1b[?%dh\x1b[?1006h", mm))
	}
}

func (d *CScreen) DisableMouse() {
	if len(d.mouse) != 0 {
		// This turns off everything.
		d.TPuts("\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l")
	}
}

func (d *CScreen) EnablePaste() {
	d.TPuts(d.enablePaste)
}

func (d *CScreen) DisablePaste() {
	d.TPuts(d.disablePaste)
}

func (d *CScreen) Size() (w, h int) {
	d.Lock()
	w, h = d.w, d.h
	d.Unlock()
	return
}

func (d *CScreen) resize() {
	if w, h, e := d.getWinSize(); e == nil {
		if w != d.w || h != d.h {
			d.cx = -1
			d.cy = -1

			d.cells.Resize(w, h)
			d.cells.Invalidate()
			d.h = h
			d.w = w
			ev := NewEventResize(w, h)
			_ = d.PostEvent(ev)
		}
	}
}

func (d *CScreen) Colors() int {
	// this doesn't'd change, no need for lock
	if d.trueColor {
		return 1 << 24
	}
	return d.ti.Colors
}

// nColors returns the size of the built-in palette.
// This is distinct from Colors(), as it will generally
// always be a small number. (<= 256)
func (d *CScreen) nColors() int {
	return d.ti.Colors
}

func (d *CScreen) PollEvent() Event {
	select {
	case <-d.quit:
		return nil
	case ev := <-d.evCh:
		return ev
	}
}

func (d *CScreen) PollEventChan() (next chan Event) {
	// next = make(chan Event)
	// for !d.finished {
	// select {
	// case <-d.quit:
	// 	next <- nil
	// case ev := <-d.evCh:
	// 	next <- ev
	// }
	// }
	next = d.evCh
	return
}

// vtACSNames is a map of bytes defined by terminfo that are used in
// the terminals Alternate Character Set to represent other glyphs.
// For example, the upper left corner of the box drawing set can be
// displayed by printing "l" while in the alternate character set.
// Its not quite that simple, since the "l" is the terminfo name,
// and it may be necessary to use a different character based on
// the terminal implementation (or the terminal may lack support for
// this altogether).  See buildAcsMap below for detail.
var vtACSNames = map[byte]rune{
	'+': paint.RuneRArrow,
	',': paint.RuneLArrow,
	'-': paint.RuneUArrow,
	'.': paint.RuneDArrow,
	'0': paint.RuneBlock,
	'`': paint.RuneDiamond,
	'a': paint.RuneCkBoard,
	'b': '␉', // VT100, Not defined by terminfo
	'c': '␌', // VT100, Not defined by terminfo
	'd': '␋', // VT100, Not defined by terminfo
	'e': '␊', // VT100, Not defined by terminfo
	'f': paint.RuneDegree,
	'g': paint.RunePlMinus,
	'h': paint.RuneBoard,
	'i': paint.RuneLantern,
	'j': paint.RuneLRCorner,
	'k': paint.RuneURCorner,
	'l': paint.RuneULCorner,
	'm': paint.RuneLLCorner,
	'n': paint.RunePlus,
	'o': paint.RuneS1,
	'p': paint.RuneS3,
	'q': paint.RuneHLine,
	'r': paint.RuneS7,
	's': paint.RuneS9,
	't': paint.RuneLTee,
	'u': paint.RuneRTee,
	'v': paint.RuneBTee,
	'w': paint.RuneTTee,
	'x': paint.RuneVLine,
	'y': paint.RuneLEqual,
	'z': paint.RuneGEqual,
	'{': paint.RunePi,
	'|': paint.RuneNEqual,
	'}': paint.RuneSterling,
	'~': paint.RuneBullet,
}

// buildAcsMap builds a map of characters that we translate from Unicode to
// alternate character encodings.  To do this, we use the standard VT100 ACS
// maps.  This is only done if the terminal lacks support for Unicode; we
// always prefer to emit Unicode glyphs when we are able.
func (d *CScreen) buildAcsMap() {
	aCsStr := d.ti.AltChars
	d.acs = make(map[rune]string)
	for len(aCsStr) > 2 {
		vSrc := aCsStr[0]
		vDst := string(aCsStr[1])
		if r, ok := vtACSNames[vSrc]; ok {
			d.acs[r] = d.ti.EnterAcs + vDst + d.ti.ExitAcs
		}
		aCsStr = aCsStr[2:]
	}
}

func (d *CScreen) PostEventWait(ev Event) {
	d.evCh <- ev
}

func (d *CScreen) PostEvent(ev Event) error {
	select {
	case d.evCh <- ev:
		return nil
	default:
		return ErrEventQFull
	}
}

func (d *CScreen) clip(x, y int) (int, int) {
	w, h := d.cells.Size()
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > w-1 {
		x = w - 1
	}
	if y > h-1 {
		y = h - 1
	}
	return x, y
}

// buildMouseEvent returns an event based on the supplied coordinates and button
// state. Note that the screen's mouse button state is updated based on the
// input to this function (i.e. it mutates the receiver).
func (d *CScreen) buildMouseEvent(x, y, btn int) *EventMouse {

	// XTerm mouse events only report at most one button at a time,
	// which may include a wheel button.  Wheel motion events are
	// reported as single impulses, while other button events are reported
	// as separate press & release events.

	button := ButtonNone
	mod := ModNone

	// Mouse wheel has bit 6 set, no release events.  It should be noted
	// that wheel events are sometimes misdelivered as mouse button events
	// during a click-drag, so we debounce these, considering them to be
	// button press events unless we see an intervening release event.
	switch btn & 0x43 {
	case 0:
		button = Button1
		d.wasBtn = true
	case 1:
		button = Button3 // Note we prefer to treat right as button 2
		d.wasBtn = true
	case 2:
		button = Button2 // And the middle button as button 3
		d.wasBtn = true
	case 3:
		button = ButtonNone
		d.wasBtn = false
	case 0x40:
		if !d.wasBtn {
			button = WheelUp
		} else {
			button = Button1
		}
	case 0x41:
		if !d.wasBtn {
			button = WheelDown
		} else {
			button = Button2
		}
	}

	if btn&0x4 != 0 {
		mod |= ModShift
	}
	if btn&0x8 != 0 {
		mod |= ModAlt
	}
	if btn&0x10 != 0 {
		mod |= ModCtrl
	}

	// Some terminals will report mouse coordinates outside the
	// screen, especially with click-drag events.  Clip the coordinates
	// to the screen in that case.
	x, y = d.clip(x, y)

	return NewEventMouse(x, y, button, mod)
}

// parseSgrMouse attempts to locate an SGR mouse record at the start of the
// buffer.  It returns true, true if it found one, and the associated bytes
// be removed from the buffer.  It returns true, false if the buffer might
// contain such an event, but more bytes are necessary (partial match), and
// false, false if the content is definitely *not* an SGR mouse record.
func (d *CScreen) parseSgrMouse(buf *bytes.Buffer, evs *[]Event) (bool, bool) {

	b := buf.Bytes()

	var x, y, btn, state int
	dig := false
	neg := false
	motion := false
	i := 0
	val := 0

	for i = range b {
		switch b[i] {
		case '\x1b':
			if state != 0 {
				return false, false
			}
			state = 1

		case '\x9b':
			if state != 0 {
				return false, false
			}
			state = 2

		case '[':
			if state != 1 {
				return false, false
			}
			state = 2

		case '<':
			if state != 2 {
				return false, false
			}
			val = 0
			dig = false
			neg = false
			state = 3

		case '-':
			if state != 3 && state != 4 && state != 5 {
				return false, false
			}
			if dig || neg {
				return false, false
			}
			neg = true // stay in state

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if state != 3 && state != 4 && state != 5 {
				return false, false
			}
			val *= 10
			val += int(b[i] - '0')
			dig = true // stay in state

		case ';':
			if neg {
				val = -val
			}
			switch state {
			case 3:
				btn, val = val, 0
				neg, dig, state = false, false, 4
			case 4:
				x, val = val-1, 0
				neg, dig, state = false, false, 5
			default:
				return false, false
			}

		case 'm', 'M':
			if state != 5 {
				return false, false
			}
			if neg {
				val = -val
			}
			y = val - 1

			motion = (btn & 32) != 0
			btn &^= 32
			if b[i] == 'm' {
				// mouse release, clear all buttons
				btn |= 3
				btn &^= 0x40
				d.buttonDn = false
			} else if motion {
				/*
				 * Some broken terminals appear to send
				 * mouse button one motion events, instead of
				 * encoding 35 (no buttons) into these events.
				 * We resolve these by looking for a non-motion
				 * event first.
				 */
				if !d.buttonDn {
					btn |= 3
					btn &^= 0x40
				}
			} else {
				d.buttonDn = true
			}
			// consume the event bytes
			for i >= 0 {
				_, _ = buf.ReadByte()
				i--
			}
			*evs = append(*evs, d.buildMouseEvent(x, y, btn))
			return true, true
		}
	}

	// incomplete & inconclusive at this point
	return true, false
}

// parseXtermMouse is like parseSgrMouse, but it parses a legacy
// X11 mouse record.
func (d *CScreen) parseXtermMouse(buf *bytes.Buffer, evs *[]Event) (bool, bool) {

	b := buf.Bytes()

	state := 0
	btn := 0
	x := 0
	y := 0

	for i := range b {
		switch state {
		case 0:
			switch b[i] {
			case '\x1b':
				state = 1
			case '\x9b':
				state = 2
			default:
				return false, false
			}
		case 1:
			if b[i] != '[' {
				return false, false
			}
			state = 2
		case 2:
			if b[i] != 'M' {
				return false, false
			}
			state++
		case 3:
			btn = int(b[i])
			state++
		case 4:
			x = int(b[i]) - 32 - 1
			state++
		case 5:
			y = int(b[i]) - 32 - 1
			for i >= 0 {
				_, _ = buf.ReadByte()
				i--
			}
			*evs = append(*evs, d.buildMouseEvent(x, y, btn))
			return true, true
		}
	}
	return true, false
}

func (d *CScreen) parseFunctionKey(buf *bytes.Buffer, evs *[]Event) (bool, bool) {
	b := buf.Bytes()
	partial := false
	for e, k := range d.keyCodes {
		esc := []byte(e)
		if (len(esc) == 1) && (esc[0] == '\x1b') {
			continue
		}
		if bytes.HasPrefix(b, esc) {
			// matched
			var r rune
			if len(esc) == 1 {
				r = rune(b[0])
			}
			mod := k.mod
			if d.escaped {
				mod |= ModAlt
				d.escaped = false
			}
			switch k.key {
			case keyPasteStart:
				*evs = append(*evs, NewEventPaste(true))
			case keyPasteEnd:
				*evs = append(*evs, NewEventPaste(false))
			default:
				*evs = append(*evs, NewEventKey(k.key, r, mod))
			}
			for i := 0; i < len(esc); i++ {
				_, _ = buf.ReadByte()
			}
			return true, true
		}
		if bytes.HasPrefix(esc, b) {
			partial = true
		}
	}
	return partial, false
}

func (d *CScreen) parseRune(buf *bytes.Buffer, evs *[]Event) (bool, bool) {
	b := buf.Bytes()
	if b[0] >= ' ' && b[0] < 0x7F {
		// printable ASCII easy to deal with -- no encodings
		mod := ModNone
		if d.escaped {
			mod = ModAlt
			d.escaped = false
		}
		*evs = append(*evs, NewEventKey(KeyRune, rune(b[0]), mod))
		_, _ = buf.ReadByte()
		return true, true
	}

	if b[0] < 0x80 {
		// Low numbered values are control keys, not runes.
		return false, false
	}

	utf := make([]byte, 12)
	for l := 1; l <= len(b); l++ {
		d.decoder.Reset()
		nOut, nIn, e := d.decoder.Transform(utf, b[:l], true)
		if e == transform.ErrShortSrc {
			continue
		}
		if nOut != 0 {
			r, _ := utf8.DecodeRune(utf[:nOut])
			if r != utf8.RuneError {
				mod := ModNone
				if d.escaped {
					mod = ModAlt
					d.escaped = false
				}
				*evs = append(*evs, NewEventKey(KeyRune, r, mod))
			}
			for nIn > 0 {
				_, _ = buf.ReadByte()
				nIn--
			}
			return true, true
		}
	}
	// Looks like potential escape
	return true, false
}

func (d *CScreen) scanInput(buf *bytes.Buffer, expire bool) {
	evs := d.collectEventsFromInput(buf, expire)

	for _, ev := range evs {
		_ = d.PostEvent(ev)
	}
}

// Return an array of Events extracted from the supplied buffer. This is done
// while holding the screen's lock - the events can then be queued for
// application processing with the lock released.
func (d *CScreen) collectEventsFromInput(buf *bytes.Buffer, expire bool) []Event {

	res := make([]Event, 0, 20)

	d.Lock()
	defer d.Unlock()

	for {
		b := buf.Bytes()
		if len(b) == 0 {
			buf.Reset()
			return res
		}

		partials := 0

		if part, comp := d.parseRune(buf, &res); comp {
			continue
		} else if part {
			partials++
		}

		if part, comp := d.parseFunctionKey(buf, &res); comp {
			continue
		} else if part {
			partials++
		}

		// Only parse mouse records if this term claims to have
		// mouse support

		if d.ti.Mouse != "" {
			if part, comp := d.parseXtermMouse(buf, &res); comp {
				continue
			} else if part {
				partials++
			}

			if part, comp := d.parseSgrMouse(buf, &res); comp {
				continue
			} else if part {
				partials++
			}
		}

		if partials == 0 || expire {
			if b[0] == '\x1b' {
				if len(b) == 1 {
					res = append(res, NewEventKey(KeyEsc, 0, ModNone))
					d.escaped = false
				} else {
					d.escaped = true
				}
				_, _ = buf.ReadByte()
				continue
			}
			// Nothing was going to match, or we timed out
			// waiting for more data -- just deliver the characters
			// to the app & let them sort it out.  Possibly we
			// should only do this for control characters like ESC.
			by, _ := buf.ReadByte()
			mod := ModNone
			if d.escaped {
				d.escaped = false
				mod = ModAlt
			}
			res = append(res, NewEventKey(KeyRune, rune(by), mod))
			continue
		}

		// well we have some partial data, wait until we get
		// some more
		break
	}

	return res
}

func (d *CScreen) gpmLoop() {
	if d.gpmRunning {
		return
	}
	d.gpmRunning = true
	if gpm, err := gpmctl.NewGPM(gpmctl.DefaultConf); err != nil {
		panic(err)
	} else {
		for {
			event, err := gpm.Read()
			if err != nil {
				panic(err)
			}
			btn := ButtonMask(0)
			if event.Type == gpmctl.DOWN || event.Type == gpmctl.UP || event.Type == gpmctl.DRAG {
				if event.Buttons&gpmctl.B_LEFT != 0 {
					btn = btn.Set(Button1)
				} else if event.Buttons&gpmctl.B_RIGHT != 0 {
					btn = btn.Set(Button2)
				} else if event.Buttons&gpmctl.B_MIDDLE != 0 {
					btn = btn.Set(Button3)
				} else if event.Buttons&gpmctl.B_FOURTH != 0 {
					btn = btn.Set(Button4)
				}
			}
			evt := NewEventMouse(int(event.X), int(event.Y), btn, ModNone)
			_ = d.PostEvent(evt)
			if !d.gpmRunning {
				return
			}
		}
	}
}

func (d *CScreen) mainLoop() {
	buf := &bytes.Buffer{}
	for {
		select {
		case <-d.quit:
			close(d.inDoneQ)
			return
		case <-d.sigWinch:
			d.Lock()
			d.cx = -1
			d.cy = -1
			d.resize()
			d.cells.Invalidate()
			d.draw()
			d.Unlock()
			continue
		case <-d.keyTimer.C:
			// If the timer fired, and the current time
			// is after the expiration of the escape sequence,
			// then we assume the escape sequence reached its
			// conclusion, and process the chunk independently.
			// This lets us detect conflicts such as a lone ESC.
			if buf.Len() > 0 {
				if time.Now().After(d.keyExpire) {
					d.scanInput(buf, true)
				}
			}
			if buf.Len() > 0 {
				if !d.keyTimer.Stop() {
					select {
					case <-d.keyTimer.C:
					default:
					}
				}
				d.keyTimer.Reset(EventKeyTiming)
			}
		case chunk := <-d.keyChan:
			buf.Write(chunk)
			d.keyExpire = time.Now().Add(EventKeyTiming)
			d.scanInput(buf, false)
			if !d.keyTimer.Stop() {
				select {
				case <-d.keyTimer.C:
				default:
				}
			}
			if buf.Len() > 0 {
				d.keyTimer.Reset(EventKeyTiming)
			}
		}
	}
}

func (d *CScreen) inputLoop() {
	for {
		chunk := make([]byte, 128)
		d.ttyReadLock.Lock()
		d.ttyReading = true
		d.ttyReadLock.Unlock()
		n, e := d.term.Read(chunk)
		d.ttyReadLock.Lock()
		d.ttyReading = false
		d.ttyReadLock.Unlock()
		switch e {
		case io.EOF:
		case nil:
		default:
			if strings.Index(e.Error(), "bad file descriptor") > -1 {
				if err := d.reengage(); err != nil {
					log.ErrorF("Screen reengage error handling \"bad file descriptor\": %v", err)
				}
			} else {
				_ = d.PostEvent(NewEventError(e))
			}
			return
		}
		d.keyChan <- chunk[:n]
	}
}

func (d *CScreen) Sync() {
	d.Lock()
	d.cx = -1
	d.cy = -1
	if !d.finished {
		d.resize()
		d.clear = true
		d.cells.Invalidate()
		d.draw()
	}
	d.Unlock()
}

func (d *CScreen) CharacterSet() string {
	return d.charset
}

func (d *CScreen) RegisterRuneFallback(orig rune, fallback string) {
	d.Lock()
	d.fallback[orig] = fallback
	d.Unlock()
}

func (d *CScreen) UnregisterRuneFallback(orig rune) {
	d.Lock()
	delete(d.fallback, orig)
	d.Unlock()
}

func (d *CScreen) CanDisplay(r rune, checkFallbacks bool) bool {

	if enc := d.encoder; enc != nil {
		nb := make([]byte, 6)
		ob := make([]byte, 6)
		num := utf8.EncodeRune(ob, r)

		enc.Reset()
		dst, _, err := enc.Transform(nb, ob[:num], true)
		if dst != 0 && err == nil && nb[0] != '\x1A' {
			return true
		}
	}
	// Terminal fallbacks always permitted, since we assume they are
	// basically nearly perfect renditions.
	if _, ok := d.acs[r]; ok {
		return true
	}
	if !checkFallbacks {
		return false
	}
	if _, ok := d.fallback[r]; ok {
		return true
	}
	return false
}

func (d *CScreen) HasMouse() bool {
	return len(d.mouse) != 0
}

func (d *CScreen) HasKey(k Key) bool {
	if k == KeyRune {
		return true
	}
	return d.keyExist[k]
}

func (d *CScreen) Export() *CellBuffer {
	d.Lock()
	defer d.Unlock()
	cb := NewCellBuffer()
	w, h := d.cells.Size()
	cb.Resize(w, h)
	for idx, cell := range d.cells.cells {
		cb.cells[idx] = cell
	}
	return cb
}

func (d *CScreen) Import(cb *CellBuffer) {
	d.Lock()
	defer d.Unlock()
	w, h := cb.Size()
	d.cells.Resize(w, h)
	for idx, cell := range cb.cells {
		d.cells.cells[idx] = cell
	}
}

func (d *CScreen) HostClipboardEnabled() (enabled bool) {
	enabled = d.useHostClipboard
	return
}

func (d *CScreen) CopyToClipboard(s string) {
	if d.useHostClipboard {
		if err := clipboard.WriteAll(s); err != nil {
			log.Error(err)
		} else {
			log.DebugF("copy sent (github.com/atotto/clipboard): %v", s)
			return
		}
	}
	if d.useTermClipboard {
		b64 := base64.StdEncoding.EncodeToString([]byte(s))
		d.TPuts("\x1b]52;c;" + b64 + "\x07")
		log.DebugF("copy sent (OSC-52 terminal sequence): %v", s)
	}
}

func (d *CScreen) PasteFromClipboard() (s string, ok bool) {
	if d.useHostClipboard {
		var err error
		if s, err = clipboard.ReadAll(); err != nil {
			log.Error(err)
		} else {
			log.DebugF("paste received (github.com/atotto/clipboard): %v", s)
			ok = true
		}
	}
	return
}

func (d *CScreen) EnableHostClipboard(enabled bool) {
	d.Lock()
	defer d.Unlock()
	d.useHostClipboard = enabled
}

func (d *CScreen) EnableTermClipboard(enabled bool) {
	d.Lock()
	defer d.Unlock()
	d.useTermClipboard = enabled
}