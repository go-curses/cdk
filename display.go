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

package cdk

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	cterm "github.com/go-curses/term"

	"github.com/go-curses/cdk/env"
	"github.com/go-curses/cdk/lib/enums"
	cexec "github.com/go-curses/cdk/lib/exec"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
	cstrings "github.com/go-curses/cdk/lib/strings"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
	"github.com/go-curses/cdk/memphis"
)

var (
	// DisplayCallCapacity limits the number of concurrent calls on main threads
	DisplayCallCapacity    = 128
	DisplayEventCapacity   = 1024
	DisplayMainsCapacity   = 128
	DisplayInboundCapacity = 1024
	// MainIterateDelay is the event iteration loop delay
	MainIterateDelay = time.Millisecond * 25
	// MainDrawInterval is the interval between renders (milliseconds)
	MainDrawInterval    int64 = 50
	MainLoopInterval    int64 = 10
	DisplayLoopCapacity       = 1024
)

const (
	TypeDisplayManager CTypeTag = "cdk-display"
)

func init() {
	_ = TypesManager.AddType(TypeDisplayManager, nil)
}

type Display interface {
	Object

	Init() (already bool)
	App() *CApplication
	Destroy()
	GetTitle() string
	SetTitle(title string)
	GetTtyPath() string
	SetTtyPath(ttyPath string)
	GetTtyHandle() *os.File
	SetTtyHandle(ttyHandle *os.File)
	GetCompressEvents() bool
	SetCompressEvents(compress bool)
	Screen() Screen
	DisplayCaptured() bool
	CaptureDisplay() (err error)
	ReleaseDisplay()
	CallEnabled() (enabled bool, err error)
	Call(fn cexec.Callback) (err error)
	Command(name string, argv ...string) (err error)
	IsMonochrome() bool
	Colors() (numberOfColors int)
	CaptureCtrlC()
	ReleaseCtrlC()
	CapturedCtrlC() bool
	GetClipboard() (clipboard Clipboard)
	FocusedWindow() Window
	FocusWindow(w Window)
	FocusNextWindow()
	FocusPreviousWindow()
	MapWindow(w Window)
	MapWindowWithRegion(w Window, region ptypes.Region)
	UnmapWindow(w Window)
	IsMappedWindow(w Window) (mapped bool)
	GetWindows() (windows []Window)
	GetWindowAtPoint(point ptypes.Point2I) (window Window)
	CursorPosition() (position ptypes.Point2I, moving bool)
	SetEventFocus(widget Object) error
	GetEventFocus() (widget Object)
	GetPriorEvent() (event Event)
	ProcessEvent(evt Event) enums.EventFlag
	RequestDraw()
	RequestShow()
	RequestSync()
	RequestQuit()
	IsRunning() bool
	StartupComplete()
	AsyncCall(fn DisplayCallbackFn) error
	AwaitCall(fn DisplayCallbackFn) error
	AsyncCallMain(fn DisplayCallbackFn) error
	AwaitCallMain(fn DisplayCallbackFn) error
	PostEvent(evt Event) error
	Run() (err error)
	Startup() (ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, err error)
	Main(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) (err error)
	MainFinish()
	HasPendingEvents() (pending bool)
	HasBufferedEvents() (hasEvents bool)
	IterateBufferedEvents() (refreshed bool)
}

// Basic display type
type CDisplay struct {
	CObject

	title string

	captureCtrlC bool
	clipboard    *CClipboard

	windows []Window

	app        *CApplication
	ttyPath    string
	ttyHandle  *os.File
	screen     Screen
	captured   bool
	started    bool
	eventFocus Object
	priorEvent Event

	cursor       *ptypes.Point2I
	cursorMoving bool

	running  bool
	closing  sync.Once
	done     chan bool
	queue    chan DisplayCallbackFn
	mains    chan DisplayCallbackFn
	events   chan Event
	buffer   []interface{}
	inbound  chan Event
	compress bool
	lastLoop time.Time
	loopNow  chan bool

	notifyLoopNow bool

	eventMutex *sync.Mutex
	drawMutex  *sync.Mutex
}

func NewDisplay(title string, ttyPath string) (d *CDisplay) {
	d = new(CDisplay)
	d.title = title
	d.ttyPath = ttyPath
	d.ttyHandle = nil
	d.Init()
	return d
}

func NewDisplayWithHandle(title string, ttyHandle *os.File) (d *CDisplay) {
	d = new(CDisplay)
	d.title = title
	d.ttyPath = ""
	d.ttyHandle = ttyHandle
	d.Init()
	return d
}

func (d *CDisplay) Init() (already bool) {
	if d.InitTypeItem(TypeDisplayManager, d) {
		return true
	}
	d.CObject.Init()

	username := env.Get("USER", "nil")
	displayname := cstrings.MakeObjectName("tty", username, "/dev/tty")
	_ = d.InstallProperty(PropertyDisplayName, StringProperty, true, displayname)
	_ = d.InstallProperty(PropertyDisplayUser, StringProperty, true, username)
	_ = d.InstallProperty(PropertyDisplayHost, StringProperty, true, "/dev/tty")

	d.captured = false
	d.started = false
	d.running = false
	d.done = make(chan bool)
	d.queue = make(chan DisplayCallbackFn, DisplayCallCapacity)
	d.mains = make(chan DisplayCallbackFn, DisplayMainsCapacity)
	d.events = make(chan Event, DisplayEventCapacity)
	d.buffer = make([]interface{}, 0)
	d.inbound = make(chan Event, DisplayInboundCapacity)
	d.compress = true
	d.lastLoop = time.Unix(0, 0)
	d.loopNow = make(chan bool, DisplayLoopCapacity)

	d.cursor = ptypes.NewPoint2I(0, 0)
	d.cursorMoving = false

	d.clipboard = nil

	d.priorEvent = nil
	d.eventFocus = nil
	d.windows = make([]Window, 0)

	d.eventMutex = &sync.Mutex{}
	d.drawMutex = &sync.Mutex{}

	theme, _ := paint.GetTheme(paint.DisplayTheme)
	d.SetTheme(theme)

	return false
}

func (d *CDisplay) App() *CApplication {
	return d.app
}

func (d *CDisplay) Destroy() {
	d.setRunning(false)
	d.ReleaseDisplay()
	d.closeChannels()
}

func (d *CDisplay) closeChannels() {
	d.closing.Do(func() {
		close(d.done)
		close(d.queue)
		close(d.mains)
		close(d.inbound)
	})
}

func (d *CDisplay) GetTitle() string {
	d.RLock()
	defer d.RUnlock()
	return d.title
}

func (d *CDisplay) SetTitle(title string) {
	d.Lock()
	d.title = title
	d.Unlock()
}

func (d *CDisplay) GetTtyPath() string {
	d.RLock()
	defer d.RUnlock()
	return d.ttyPath
}

func (d *CDisplay) SetTtyPath(ttyPath string) {
	d.Lock()
	d.ttyPath = ttyPath
	d.Unlock()
}

func (d *CDisplay) GetTtyHandle() *os.File {
	d.RLock()
	defer d.RUnlock()
	return d.ttyHandle
}

func (d *CDisplay) SetTtyHandle(ttyHandle *os.File) {
	d.Lock()
	d.ttyHandle = ttyHandle
	d.Unlock()
}

func (d *CDisplay) GetCompressEvents() bool {
	d.RLock()
	defer d.RUnlock()
	return d.compress
}

func (d *CDisplay) SetCompressEvents(compress bool) {
	d.Lock()
	d.compress = compress
	d.Unlock()
}

func (d *CDisplay) Screen() Screen {
	d.RLock()
	defer d.RUnlock()
	return d.screen
}

func (d *CDisplay) DisplayCaptured() bool {
	d.RLock()
	defer d.RUnlock()
	return d.screen != nil && d.captured
}

func (d *CDisplay) CaptureDisplay() (err error) {
	d.Lock()
	if d.ttyPath == OffscreenTtyPath {
		d.screen = NewOffScreen("UTF-8")
	} else {
		if d.screen, err = NewScreen(); err != nil {
			d.Unlock()
			return fmt.Errorf("error getting new screen: %v", err)
		}
	}
	if d.ttyHandle != nil {
		if err = d.screen.InitWithFileHandle(d.ttyHandle); err != nil {
			d.Unlock()
			return fmt.Errorf("error initializing new tty handle screen: %v", err)
		}
	} else {
		if err = d.screen.InitWithFilePath(d.ttyPath); err != nil {
			d.Unlock()
			return fmt.Errorf("error initializing new tty path screen: %v", err)
		}
	}
	theme, _ := paint.GetTheme(paint.DisplayTheme)
	enabled, _ := d.CallEnabled()
	d.screen.TtyCloseWithStiRead(enabled)
	d.screen.EnableMouse()
	d.screen.EnablePaste()
	d.screen.SetStyle(theme.Content.Normal)
	d.screen.Clear()
	d.captured = true
	d.Unlock()
	d.SetTheme(theme)

	d.Emit(SignalDisplayCaptured, d)
	return
}

func (d *CDisplay) ReleaseDisplay() {
	if d.DisplayCaptured() {
		d.Lock()
		d.screen.Close()
		d.screen = nil
		d.captured = false
		d.Unlock()
	}
}

func (d *CDisplay) CallEnabled() (enabled bool, err error) {
	enabled = true
	remote := false
	if name, e := d.GetStringProperty(PropertyDisplayHost); e == nil {
		remote = name != "/dev/tty"
	}
	if Build.DisableLocalCall && !remote {
		enabled = false
		err = fmt.Errorf("local call feature is disabled")
	}
	if Build.DisableRemoteCall && remote {
		enabled = false
		err = fmt.Errorf("remote call feature is disabled")
	}
	return
}

func (d *CDisplay) Call(fn cexec.Callback) (err error) {
	if enabled, err := d.CallEnabled(); !enabled {
		return err
	}
	if !d.startedAndCaptured() {
		return fmt.Errorf("display is not captured or not completely started up yet")
	}
	d.LogDebug("starting new Call")
	if d.ttyHandle != nil {
		d.screen.TtyKeepFileHandle(true)
	}
	d.ReleaseDisplay()
	d.LogDebug("display released, calling fn")

	var e error
	var callTty *os.File

	if d.ttyHandle != nil {
		var dupeFd int
		if dupeFd, err = syscall.Dup(int(d.ttyHandle.Fd())); err != nil {
			return fmt.Errorf("syscall.Dup error: %v", err)
		}
		callTty = os.NewFile(uintptr(dupeFd), d.ttyHandle.Name())
		d.LogDebug("callTty = os.NewFile(%v, %v)", dupeFd, callTty.Name())
	} else {
		ttyPath := "/dev/tty"
		if d.ttyPath != "" {
			ttyPath = d.ttyPath
		}
		if callTty, err = os.OpenFile(ttyPath, os.O_RDWR, 0); err != nil {
			return fmt.Errorf("os.OpenFile error: %v", err)
		}
		d.LogDebug("callTty = os.OpenFile(%v)", ttyPath)
	}

	if err = cexec.CallWithTty(callTty, fn); err != nil {
		return fmt.Errorf("cexec.CallWithTty error: %v", err)
	}

	d.LogDebug("sending Tiocsti: %v", callTty.Name())
	if err := cterm.Tiocsti(callTty.Fd(), " "); err != nil {
		log.Error(err)
		d.LogDebug("[trying again] writing Tiocsti: %v", callTty.Name())
		if _, err := callTty.Write([]byte(" ")); err != nil {
			log.Error(err)
		}
	}

	d.LogDebug("closing callTty: %v", callTty.Name())
	if e = callTty.Close(); e != nil {
		d.LogErr(e)
	}

	d.LogDebug("fn released, capturing display")
	if e = d.CaptureDisplay(); e != nil {
		d.LogErr(e)
	} else if d.startedAndCaptured() {
		d.LogDebug("restoring display")
		d.RequestDraw()
		d.RequestSync()
	} else {
		d.LogError("attempted capture display, yet not started and captured")
	}
	return
}

func (d *CDisplay) Command(name string, argv ...string) (err error) {
	return d.Call(func(in, out *os.File) (err error) {
		d.LogDebug("invoking exec.Command: %v %v", name, argv)
		cmd := exec.Command(name, argv...)
		cmd.Stdin = in
		cmd.Stdout = out
		cmd.Stderr = out
		return cmd.Run()
	})
}

func (d *CDisplay) IsMonochrome() bool {
	return d.Colors() == 0
}

func (d *CDisplay) Colors() (numberOfColors int) {
	numberOfColors = 0
	if d.screen != nil {
		numberOfColors = d.screen.Colors()
	}
	return
}

func (d *CDisplay) CaptureCtrlC() {
	d.Lock()
	defer d.Unlock()
	d.captureCtrlC = true
}

func (d *CDisplay) ReleaseCtrlC() {
	d.Lock()
	defer d.Unlock()
	d.captureCtrlC = false
}

func (d *CDisplay) CapturedCtrlC() bool {
	d.RLock()
	defer d.RUnlock()
	return d.captureCtrlC
}

func (d *CDisplay) GetClipboard() (clipboard Clipboard) {
	d.RLock()
	defer d.RUnlock()
	if d.clipboard == nil {
		if d.screen != nil {
			d.RUnlock()
			d.Lock()
			d.clipboard = newClipboard(d.screen)
			d.Unlock()
			d.RLock()
		}
	}
	return d.clipboard
}

func (d *CDisplay) SetTheme(theme paint.Theme) {
	d.CObject.SetTheme(theme)
	d.Lock()
	defer d.Unlock()
	if d.screen != nil {
		d.screen.SetStyle(d.GetTheme().Content.Normal)
	}
}

func (d *CDisplay) resizeWindowSurfacesOnStartupCompleted() (w, h int) {
	d.RLock()
	if d.screen == nil {
		d.RUnlock()
		return
	}
	w, h = d.screen.Size()
	d.RUnlock()
	d.Lock()
	defer d.Unlock()

	theme := d.GetTheme()
	style := theme.Content.Normal

	size := ptypes.MakeRectangle(w, h)
	if err := memphis.MakeConfigureSurface(d.ObjectID(), ptypes.MakePoint2I(0, 0), size, style); err != nil {
		d.LogErr(err)
	}

	return
}

func (d *CDisplay) findMappedWindowIndex(w Window) (index int) {
	d.RLock()
	index = -1
	for idx, window := range d.windows {
		if w.ObjectID() == window.ObjectID() {
			index = idx
			break
		}
	}
	d.RUnlock()
	return
}

func (d *CDisplay) FocusedWindow() Window {
	d.RLock()
	defer d.RUnlock()
	if numWindows := len(d.windows); numWindows > 0 {
		for i := 0; i < numWindows; i++ {
			if d.windows[i].GetWindowType() == enums.WINDOW_TOPLEVEL {
				return d.windows[i]
			}
		}
	}
	return nil
}

func (d *CDisplay) FocusWindow(w Window) {
	mappedWindowIndex := d.findMappedWindowIndex(w)
	if mappedWindowIndex > -1 {
		d.Lock()
		existing := d.windows[mappedWindowIndex]
		d.windows = append(d.windows[:mappedWindowIndex], d.windows[mappedWindowIndex+1:]...)
		d.windows = append([]Window{existing}, d.windows...)
		d.Unlock()
	} else {
		d.MapWindow(w)
	}
	d.Emit(SignalFocusedWindow, w)
}

func (d *CDisplay) FocusNextWindow() {
	windows := d.GetWindows()
	numWindows := len(windows)
	if numWindows > 1 {
		if f := d.Emit(SignalFocusNextWindow, d, windows[1]); f == enums.EVENT_PASS {
			d.FocusWindow(windows[1])
			return
		}
	}
}

func (d *CDisplay) FocusPreviousWindow() {
	windows := d.GetWindows()
	numWindows := len(windows)
	lastWindow := numWindows - 1
	if numWindows > 1 {
		if f := d.Emit(SignalFocusPreviousWindow, d, windows[lastWindow]); f == enums.EVENT_PASS {
			d.FocusWindow(windows[lastWindow])
			return
		}
	}
}

func (d *CDisplay) MapWindow(w Window) {
	w.SetDisplay(d)
	width, height := 0, 0
	d.RLock()
	if d.startedAndCaptured() {
		width, height = d.screen.Size()
	}
	d.RUnlock()
	region := ptypes.MakeRegion(0, 0, width, height)
	d.MapWindowWithRegion(w, region)
}

func (d *CDisplay) MapWindowWithRegion(w Window, region ptypes.Region) {
	log.DebugDF(1, "mapping window: %v, with region: %v", w.ObjectName(), region)
	index := d.findMappedWindowIndex(w)
	w.SetDisplay(d)
	style := w.GetTheme().Content.Normal
	if err := memphis.MakeConfigureSurface(w.ObjectID(), region.Origin(), region.Size(), style); err != nil {
		d.LogErr(err)
	}
	d.Lock()
	if index > -1 {
		d.windows = append(d.windows[:index], d.windows[index+1:]...)
	}
	d.windows = append([]Window{w}, d.windows...)
	d.Unlock()
	d.RequestDraw()
	d.RequestShow()
	w.Emit(SignalMappedWindow, d)
}

func (d *CDisplay) UnmapWindow(w Window) {
	if idx := d.findMappedWindowIndex(w); idx > -1 {
		d.LogDebug("unmapping window: %v", w.ObjectName())
		d.Lock()
		memphis.RemoveSurface(w.ObjectID())
		d.windows = append(d.windows[:idx], d.windows[idx+1:]...)
		var restoreFocusedWindow Window
		if len(d.windows) > 0 {
			restoreFocusedWindow = d.windows[0]
		}
		d.Unlock()
		d.RequestDraw()
		d.RequestShow()
		w.Emit(SignalUnmappedWindow, d)
		if restoreFocusedWindow != nil {
			d.FocusWindow(restoreFocusedWindow)
		}
	}
}

func (d *CDisplay) IsMappedWindow(w Window) (mapped bool) {
	mapped = d.findMappedWindowIndex(w) > -1
	return
}

func (d *CDisplay) GetWindows() (windows []Window) {
	d.RLock()
	defer d.RUnlock()
	for _, w := range d.windows {
		windows = append(windows, w)
	}
	return
}

func (d *CDisplay) GetWindowAtPoint(point ptypes.Point2I) (window Window) {
	d.RLock()
	for i := 0; i < len(d.windows); i++ {
		if surface, err := memphis.GetSurface(d.windows[i].ObjectID()); err != nil {
			d.LogErr(err)
		} else {
			region := surface.GetRegion()
			if region.HasPoint(point) {
				window = d.windows[i]
				break
			}
		}
	}
	d.RUnlock()
	return
}

func (d *CDisplay) CursorPosition() (position ptypes.Point2I, moving bool) {
	// d.RLock()
	position = d.cursor.Clone()
	moving = d.cursorMoving
	// d.RUnlock()
	return
}

func (d *CDisplay) SetEventFocus(widget Object) error {
	d.Lock()
	if widget != nil {
		if _, ok := widget.Self().(Sensitive); !ok {
			d.Unlock()
			return fmt.Errorf("widget does not implement Sensitive: %v (%T)", widget, widget)
		}
	}
	d.Unlock()
	if f := d.Emit(SignalSetEventFocus); f == enums.EVENT_PASS {
		d.Lock()
		d.eventFocus = widget
		d.Unlock()
	}
	return nil
}

func (d *CDisplay) GetEventFocus() (widget Object) {
	d.RLock()
	defer d.RUnlock()
	widget = d.eventFocus
	return
}

func (d *CDisplay) GetPriorEvent() (event Event) {
	d.RLock()
	defer d.RUnlock()
	return d.priorEvent
}

// ProcessEvent handles events sent from the Screen instance and manages passing
// those events to the active window
func (d *CDisplay) ProcessEvent(evt Event) enums.EventFlag {
	if !d.startedAndCaptured() {
		return enums.EVENT_PASS
	}

	d.eventMutex.Lock()
	defer func() {
		d.Lock()
		d.priorEvent = evt
		d.Unlock()
		d.eventMutex.Unlock()
	}()

	if _, ok := evt.(*EventQuit); ok {
		d.done <- true
		return enums.EVENT_STOP
	}

	if req, ok := evt.(*EventRender); ok {
		d.RLock()
		hasScreen := d.screen != nil
		d.RUnlock()
		if hasScreen {
			if req.Draw() {
				d.renderScreen()
			}
			if req.Sync() {
				d.RLock()
				if d.screen != nil {
					d.screen.Sync()
				}
				d.RUnlock()
			} else if req.Show() {
				d.RLock()
				if d.screen != nil {
					d.screen.Show()
				}
				d.RUnlock()
			}
		}
		return enums.EVENT_STOP
	}

	if d.eventFocus != nil {
		if sensitive, ok := d.eventFocus.Self().(Sensitive); ok {
			return sensitive.ProcessEvent(evt)
		}
		d.LogError("event focus does not implement Sensitive: %v (%T)", d.eventFocus, d.eventFocus)
		return enums.EVENT_PASS
	}

	switch e := evt.(type) {
	case *EventPaste:
		if w := d.FocusedWindow(); w != nil {
			if f := w.ProcessEvent(e); f == enums.EVENT_STOP {
				d.RequestDraw()
				d.RequestShow()
				return enums.EVENT_STOP
			}
		}
		if f := d.Emit(SignalEventPaste, d, e); f == enums.EVENT_STOP {
			d.RequestDraw()
			d.RequestShow()
			return enums.EVENT_STOP
		}
		return enums.EVENT_PASS

	case *EventError:
		d.LogError("EventError: %v", e)
		if w := d.FocusedWindow(); w != nil {
			if f := w.ProcessEvent(e); f == enums.EVENT_STOP {
				d.RequestDraw()
				d.RequestShow()
				return enums.EVENT_STOP
			}
		}
		if f := d.Emit(SignalEventError, d, e); f == enums.EVENT_STOP {
			d.RequestDraw()
			d.RequestShow()
			return enums.EVENT_STOP
		}
		return enums.EVENT_PASS

	case *EventKey:
		if d.captureCtrlC {
			switch e.Rune() {
			case rune(KeyCtrlC):
				d.LogTrace("display captured <Ctrl+C>")
				if f := d.Emit(SignalInterrupt, d); f == enums.EVENT_STOP {
					return enums.EVENT_STOP
				}
				d.RequestQuit()
				return enums.EVENT_STOP
			}
		}
		if w := d.FocusedWindow(); w != nil {
			if f := w.ProcessEvent(e); f == enums.EVENT_STOP {
				d.RequestDraw()
				d.RequestShow()
				return enums.EVENT_STOP
			}
		}
		if f := d.Emit(SignalEventKey, d, e); f == enums.EVENT_STOP {
			d.RequestDraw()
			d.RequestShow()
			return enums.EVENT_STOP
		}
		return enums.EVENT_PASS

	case *EventMouse:
		d.Lock()
		d.cursor.Set(e.Position())
		d.cursorMoving = e.IsMoving() || e.IsDragging()
		d.Unlock()
		if w := d.FocusedWindow(); w != nil {
			if f := w.ProcessEvent(e); f == enums.EVENT_STOP {
				d.RequestDraw()
				d.RequestShow()
				return enums.EVENT_STOP
			}
		}
		if f := d.Emit(SignalEventMouse, d, e); f == enums.EVENT_STOP {
			d.RequestDraw()
			d.RequestShow()
			return enums.EVENT_STOP
		}
		return enums.EVENT_PASS

	case *EventResize:
		origin := ptypes.MakePoint2I(0, 0)
		alloc := ptypes.MakeRectangle(e.Size())
		style := d.GetTheme().Content.Normal
		if err := memphis.MakeConfigureSurface(d.ObjectID(), origin, alloc, style); err != nil {
			d.LogErr(err)
		}
		// all windows get resize event
		for _, window := range d.GetWindows() {
			window.ProcessEvent(e)
		}
		f := d.Emit(SignalEventResize, d, e)
		d.RequestDraw()
		d.RequestSync()
		return f

	default:
		d.LogWarn("processing unknown event: (%T)%v", e, e)
	}

	if w := d.FocusedWindow(); w != nil {
		if f := w.ProcessEvent(evt); f == enums.EVENT_STOP {
			d.RequestDraw()
			d.RequestShow()
			return enums.EVENT_STOP
		}
	}
	if f := d.Emit(SignalEvent, d, evt); f == enums.EVENT_STOP {
		d.RequestDraw()
		d.RequestShow()
		return enums.EVENT_STOP
	}
	return enums.EVENT_PASS
}

func (d *CDisplay) renderScreen() enums.EventFlag {
	if !d.DisplayCaptured() || !d.IsRunning() {
		return enums.EVENT_PASS
	}
	d.drawMutex.Lock()
	defer d.drawMutex.Unlock()
	d.Lock()
	windows := d.windows
	d.Unlock()
	if surface, err := memphis.GetSurface(d.ObjectID()); err == nil {
		theme := d.GetTheme()
		surface.Fill(theme)
		for i := len(windows) - 1; i >= 0; i-- {
			windows[i].Draw()
			if err := surface.Composite(windows[i].ObjectID()); err != nil {
				d.LogErr(err)
			}
		}
		d.Lock()
		if d.screen != nil {
			if err := surface.Render(d.screen); err != nil {
				d.LogErr(err)
			}
		}
		d.Unlock()
		return enums.EVENT_STOP
	}
	d.LogError("missing surface for display: %v", d.ObjectID())
	return enums.EVENT_PASS
}

// RequestDraw asks the Display to process a SignalDraw event cycle, this does
// not actually render the contents to in Screen, just update
func (d *CDisplay) RequestDraw() {
	_ = d.PostEvent(NewEventDraw())
}

// RequestShow asks the Display to render pending Screen changes
func (d *CDisplay) RequestShow() {
	_ = d.PostEvent(NewEventShow())
}

// RequestSync asks the Display to render everything in the Screen
func (d *CDisplay) RequestSync() {
	_ = d.PostEvent(NewEventShow())
}

// RequestQuit asks the Display to quit nicely
func (d *CDisplay) RequestQuit() {
	_ = d.PostEvent(NewEventQuit())
}

// IsRunning returns TRUE if the main thread is currently running.
func (d *CDisplay) IsRunning() bool {
	d.RLock()
	defer d.RUnlock()
	return d.running
}

func (d *CDisplay) setRunning(isRunning bool) {
	d.Lock()
	defer d.Unlock()
	d.running = isRunning
}

// StartupComplete emits SignalStartupComplete
func (d *CDisplay) StartupComplete() {
	w, h := d.resizeWindowSurfacesOnStartupCompleted()
	d.Emit(SignalStartupComplete)
	_ = d.PostEvent(NewEventResize(w, h))
}

// AsyncCall runs the given DisplayCallbackFn on the UI thread, non-blocking
func (d *CDisplay) AsyncCall(fn DisplayCallbackFn) error {
	if !d.IsRunning() {
		return fmt.Errorf("application not running")
	}
	d.queue <- fn
	return nil
}

// AwaitCall runs the given DisplayCallbackFn on the UI thread, blocking
func (d *CDisplay) AwaitCall(fn DisplayCallbackFn) error {
	if !d.IsRunning() {
		return fmt.Errorf("application not running")
	}
	var err error
	done := make(chan bool)
	d.queue <- func(d Display) error {
		err = fn(d)
		done <- true
		return nil
	}
	<-done
	return err
}

// AsyncCallMain will run the given DisplayCallbackFn on the main runner thread,
// non-blocking
func (d *CDisplay) AsyncCallMain(fn DisplayCallbackFn) error {
	if !d.IsRunning() {
		return fmt.Errorf("application not running")
	}
	d.mains <- fn
	return nil
}

// AwaitCallMain will run the given DisplayCallbackFn on the main runner thread,
// blocking
func (d *CDisplay) AwaitCallMain(fn DisplayCallbackFn) error {
	if !d.IsRunning() {
		return fmt.Errorf("application not running")
	}
	var err error
	done := make(chan bool)
	d.mains <- func(d Display) error {
		err = fn(d)
		done <- true
		return nil
	}
	<-done
	return err
}

// PostEvent sends the given Event to the Display Screen for processing. This
// is mainly useful for synthesizing Screen events, though not a recommended
// practice.
func (d *CDisplay) PostEvent(evt Event) error {
	if !d.IsRunning() {
		return fmt.Errorf("application not running")
	}
	d.events <- evt
	return nil
}

func (d *CDisplay) pollEventWorker(ctx context.Context) {
	// this happens in its own go thread
pollEventWorkerLoop:
	for d.IsRunning() {
		if d.DisplayCaptured() {
			select {
			case evt := <-d.screen.PollEventChan():
				d.inbound <- evt
			case <-ctx.Done():
				break pollEventWorkerLoop
			}
		} else {
			time.Sleep(MainIterateDelay)
		}
	}
}

func (d *CDisplay) processEventWorker(ctx context.Context) {
	// this happens in its own go thread
processEventWorkerLoop:
	for d.IsRunning() {
		select {
		case <-ctx.Done():
			break processEventWorkerLoop

		case evt := <-d.inbound:
			if evt != nil {
				// store the instance by type rather than the Event interface
				switch t := evt.(type) {
				default:
					d.Lock()
					d.buffer = append(d.buffer, t)
					if d.notifyLoopNow {
						d.loopNow <- true
					}
					d.Unlock()
				}
			} else {
				// nil event, quit?
				break
			}

		case fn, ok := <-d.queue:
			if ok {
				if err := fn(d); err != nil {
					log.ErrorF("async/await handler error: %v", err)
				}
			}
		}
	}
}

func (d *CDisplay) startedAndCaptured() bool {
	d.RLock()
	defer d.RUnlock()
	return d.started && d.captured && d.screen != nil
}

// Run is the standard means of invoking a Display instance. It calls Startup,
// handles the main event look and finally calls MainFinish when all is
// complete.
func (d *CDisplay) Run() (err error) {
	var ctx context.Context
	var wg *sync.WaitGroup
	if ctx, _, wg, err = d.Startup(); err != nil {
		return
	}
	d.notifyLoopNow = true
	wg.Add(1)
	Go(func() {
		for d.HasPendingEvents() {
			select {
			case <-ctx.Done():
				wg.Done()
				return
			case <-d.loopNow:
				d.IterateBufferedEvents()
			}
		}
		wg.Done()
	})
	wg.Wait()
	d.MainFinish()
	return
}

// Startup captures the Display, sets the internal running state and allocates
// the necessary runtime context.WithCancel and sync.WaitGroup for the main
// runner thread of the Display. Once setup, starts the Main runner with the
// necessary rigging for thread synchronization and shutdown mechanics.
func (d *CDisplay) Startup() (ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, err error) {
	if err = d.CaptureDisplay(); err != nil {
		d.LogErr(err)
		return
	}
	d.Connect(SignalStartupComplete, DisplayStartupCompleteHandle, func(data []interface{}, argv ...interface{}) enums.EventFlag {
		_ = d.Disconnect(SignalStartupComplete, DisplayStartupCompleteHandle)
		d.Lock()
		d.started = true
		d.Unlock()
		return enums.EVENT_PASS
	})
	d.setRunning(true)
	ctx, cancel = context.WithCancel(context.Background())
	wg = &sync.WaitGroup{}
	wg.Add(1)
	Go(func() {
		if err := d.Main(ctx, cancel, wg); err != nil {
			d.LogErr(err)
		}
		wg.Done()
	})
	_ = d.AsyncCall(func(_ Display) error {
		d.Emit(SignalDisplayStartup, ctx, cancel, wg)
		return nil
	})
	return
}

// Main is the primary Display thread. It starts the event receiver, event
// processor and screen worker threads and proceeds to handle AsyncCallMain,
// AwaitCallMain, screen event transmitter and shutdown mechanics. When
// RequestQuit is called, the main loop exits, cancels all threads, destroys the
// display object, recovers from any go panics and finally emits a
// SignalDisplayShutdown.
func (d *CDisplay) Main(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) (err error) {
	wg.Add(1)
	Go(func() {
		d.pollEventWorker(ctx)
		wg.Done()
	})
	wg.Add(1)
	Go(func() {
		d.processEventWorker(ctx)
		wg.Done()
	})
mainForLoop:
	for d.IsRunning() {
		select {
		case fn, ok := <-d.mains:
			if ok {
				if d.DisplayCaptured() {
					if err := fn(d); err != nil {
						log.Error(err)
					}
				}
			}
		case evt, ok := <-d.events:
			if ok {
				if d.DisplayCaptured() {
					if err := d.screen.PostEvent(evt); err != nil {
						log.Error(err)
					}
				}
			}
		case <-d.done:
			d.setRunning(false)
			CancelAllTimeouts()
			cancel() // notify threads to exit
			// guarantee main calls
			for i := 0; i < len(d.mains); i++ {
				if fn, ok := <-d.mains; ok {
					if err := fn(d); err != nil {
						log.Error(err)
					}
				}
			}
			// guarantee async calls
			for i := 0; i < len(d.queue); i++ {
				if fn, ok := <-d.queue; ok {
					if err := fn(d); err != nil {
						log.ErrorF("async/await handler error: %v", err)
					}
				}
			}
			break mainForLoop
		}
	}
	d.Destroy()
	if p := recover(); p != nil {
		panic(p)
	}
	d.Emit(SignalDisplayShutdown)
	return nil
}

// MainFinish cleans up any pending internal processes remaining after Main()
// has completed processing.
func (d *CDisplay) MainFinish() {
	if goProfile != nil {
		log.DebugF("stopping profiling")
		goProfile.Stop()
	}
	return
}

// HasPendingEvents returns TRUE if there are any pending events, or if the Main
// thread is still running (and waiting for events).
func (d *CDisplay) HasPendingEvents() (pending bool) {
	if d.HasBufferedEvents() {
		pending = true
	} else if d.IsRunning() {
		pending = true
	}
	return
}

// HasBufferedEvents returns TRUE if there are any pending events buffered.
func (d *CDisplay) HasBufferedEvents() (hasEvents bool) {
	d.RLock()
	defer d.RUnlock()
	hasEvents = len(d.buffer) > 0
	return
}

// IterateBufferedEvents compresses the pending event buffer by reducing
// multiple events of the same type to just the last ones received. Each
// remaining pending event is then processed. If any of the events return
// EVENT_STOP from their signal listeners, draw and show requests are made to
// refresh the display contents.
func (d *CDisplay) IterateBufferedEvents() (refreshed bool) {
	if !d.DisplayCaptured() {
		return false
	}

	d.Lock()
	buffer := d.buffer
	d.buffer = nil
	d.Unlock()

	var render *EventRender
	pending := make([]interface{}, 0)

	for _, e := range buffer {
		switch t := e.(type) {
		case *EventPaste, *EventKey:
			// never compress paste or keys
			pending = append(pending, t)

		case *EventRender:
			// always compress render into a single request event
			if render == nil {
				render = t
			} else if t.Draw() && !render.Draw() {
				render = NewEventRender(true, render.Show(), render.Sync())
			} else if t.Show() && !render.Show() && !render.Sync() {
				render = NewEventRender(render.Draw(), true, render.Sync())
			} else if t.Sync() && !render.Sync() {
				render = NewEventRender(render.Draw(), false, true)
			}

		default:
			if d.GetCompressEvents() {
				last := len(pending) - 1
				if last <= 0 {
					pending = append(pending, t)
				} else {
					// only compressing repeats
					thisType := fmt.Sprintf("%T", t)
					lastType := fmt.Sprintf("%T", pending[last])
					if thisType == lastType {
						pending[last] = t
					} else {
						pending = append(pending, t)
					}
				}
			} else {
				pending = append(pending, t)
			}
		}
	}

	buffer = nil

	stopped := false
	for _, e := range pending {
		if evt, ok := e.(Event); ok {
			if f := d.ProcessEvent(evt); f == enums.EVENT_STOP {
				stopped = true
			}
		}
	}

	if render != nil {
		d.ProcessEvent(render)
		return true
	} else if stopped {
		d.RequestDraw()
		d.RequestShow()
		return true
	}

	return false
}

const (
	SignalDisplayCaptured     Signal = "display-captured"
	SignalInterrupt           Signal = "sigint"
	SignalEvent               Signal = "event"
	SignalEventError          Signal = "event-error"
	SignalEventKey            Signal = "event-key"
	SignalEventMouse          Signal = "event-mouse"
	SignalEventResize         Signal = "event-resize"
	SignalEventPaste          Signal = "event-paste"
	SignalSetEventFocus       Signal = "set-event-focus"
	SignalStartupComplete     Signal = "startup-complete"
	SignalDisplayStartup      Signal = "display-startup"
	SignalDisplayShutdown     Signal = "display-shutdown"
	SignalMappedWindow        Signal = "mapped-window"
	SignalUnmappedWindow      Signal = "unmapped-window"
	SignalFocusedWindow       Signal = "focused-window"
	SignalFocusNextWindow     Signal = "focus-next-window"
	SignalFocusPreviousWindow Signal = "focus-previous-window"
)

const (
	PropertyDisplayName Property = "display-name"
	PropertyDisplayUser Property = "display-user"
	PropertyDisplayHost Property = "display-host"
)

const (
	DisplayStartupCompleteHandle = "display-screen-startup-complete-handler"
)

type DisplayCallbackFn = func(d Display) error

type DisplayCommandFn = func(in, out *os.File) error

func DisplaySignalDisplayStartupArgv(argv ...interface{}) (ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, ok bool) {
	if len(argv) == 3 {
		if ctx, ok = argv[0].(context.Context); ok {
			if cancel, ok = argv[1].(context.CancelFunc); ok {
				if wg, ok = argv[2].(*sync.WaitGroup); ok {
					return
				}
				cancel = nil
			}
			ctx = nil
		}
	}
	return
}
