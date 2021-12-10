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
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
	"github.com/go-curses/cdk/memphis"
	cterm "github.com/go-curses/term"
)

var (
	// DisplayCallCapacity limits the number of concurrent calls on main threads
	DisplayCallCapacity = 16
	// MainIterateDelay is the event iteration loop delay
	MainIterateDelay = time.Millisecond * 50
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
	Call(fn DisplayCommandFn) (err error)
	Command(name string, argv ...string) (err error)
	IsMonochrome() bool
	Colors() (numberOfColors int)
	CaptureCtrlC()
	ReleaseCtrlC()
	DefaultTheme() paint.Theme
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
	Startup() (ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup)
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
	requests chan displayRequest
	compress bool

	runLock    *sync.RWMutex
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

	d.captured = false
	d.started = false
	d.running = false
	d.done = make(chan bool)
	d.queue = make(chan DisplayCallbackFn, DisplayCallCapacity)
	d.mains = make(chan DisplayCallbackFn, DisplayCallCapacity)
	d.events = make(chan Event, DisplayCallCapacity)
	d.buffer = make([]interface{}, 0)
	d.inbound = make(chan Event, DisplayCallCapacity)
	d.requests = make(chan displayRequest, DisplayCallCapacity)
	d.compress = true

	d.cursor = ptypes.NewPoint2I(0, 0)
	d.cursorMoving = false

	d.priorEvent = nil
	d.eventFocus = nil
	d.windows = make([]Window, 0)
	d.SetTheme(paint.DefaultColorTheme)

	d.runLock = &sync.RWMutex{}
	d.eventMutex = &sync.Mutex{}
	d.drawMutex = &sync.Mutex{}

	if err := memphis.RegisterSurface(d.ObjectID(), ptypes.MakePoint2I(0, 0), ptypes.MakeRectangle(0, 0), paint.DefaultColorStyle); err != nil {
		d.LogErr(err)
	}
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
		close(d.requests)
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
	d.screen.SetStyle(paint.DefaultColorStyle)
	d.screen.EnableMouse()
	d.screen.EnablePaste()
	d.screen.Clear()
	d.captured = true
	d.Unlock()
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

func (d *CDisplay) ioCopy(tag string, dst io.Writer, src io.Reader) (err error) {
	d.LogDebug("start copy: %s", tag)
	n := 0
	buf := make([]byte, 1)
	for {
		// log.DebugF("waiting for read: %s", tag)
		n, err = src.Read(buf)
		// log.DebugF("read %v for: %s", buf[:n], tag)
		if err != nil && err != io.EOF {
			break
		}
		if n == 0 {
			break
		}
		if _, err = dst.Write(buf[:n]); err != nil {
			break
		}
	}
	d.LogDebug("finish copy: [%s]", tag)
	return
}

func (d *CDisplay) Call(fn DisplayCommandFn) (err error) {
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
			return
		}
		callTty = os.NewFile(uintptr(dupeFd), d.ttyHandle.Name())
		d.LogDebug("callTty = from d.ttyHandle(%v)", callTty.Name())
	} else {
		ttyPath := "/dev/tty"
		if d.ttyPath != "" {
			ttyPath = d.ttyPath
		}
		d.LogDebug("callTty = os.OpenFile(%v)", ttyPath)
		if callTty, err = os.OpenFile(ttyPath, os.O_APPEND|os.O_RDWR, 0); err != nil {
			return
		}
	}

	var ptmx, ptty *os.File
	if ptmx, ptty, err = pty.Open(); err != nil {
		return
	}

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(callTty, ptmx); err != nil {
				d.LogError("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	var oldState *term.State
	if oldState, err = term.MakeRaw(int(callTty.Fd())); err != nil {
		return
	}

	var nok bool

	Go(func() {
		if e := d.ioCopy(
			fmt.Sprintf("NOK %v->%v", callTty.Name(), ptmx.Name()),
			ptmx,
			callTty,
		); e != nil {
			d.LogErr(e)
		}
		nok = true
	})

	Go(func() {
		if e := d.ioCopy(
			fmt.Sprintf("OK %v->%v", ptmx.Name(), callTty.Name()),
			callTty,
			ptmx,
		); e != nil {
			d.LogErr(e)
		}
	})

	err = fn(ptty, ptty)
	time.Sleep(time.Millisecond * 100) // let things catch up?

	// Cleanup signals when done.
	signal.Stop(ch)
	close(ch)

	if e = ptmx.Close(); e != nil {
		d.LogErr(e)
	}

	if e = ptty.Close(); e != nil {
		d.LogErr(e)
	}

	if !nok {
		d.LogDebug("sending Tiocsti to: %v", callTty.Name())
		if e = cterm.Tiocsti(callTty.Fd(), " "); e != nil {
			d.LogErr(e)
		}
	}

	if oldState != nil {
		d.LogDebug("restoring term state: %v", callTty.Name())
		if e = term.Restore(int(callTty.Fd()), oldState); e != nil {
			d.LogErr(e)
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

func (d *CDisplay) DefaultTheme() paint.Theme {
	d.RLock()
	screen := d.screen
	d.RUnlock()
	if screen != nil {
		if screen.Colors() <= 0 {
			return paint.DefaultMonoTheme
		}
	}
	return paint.DefaultColorTheme
}

func (d *CDisplay) resizeWindows() {
	d.RLock()
	if d.screen == nil {
		d.RUnlock()
		return
	}
	windows := d.windows
	w, h := d.screen.Size()
	size := ptypes.MakeRectangle(w, h)
	d.RUnlock()
	d.Lock()
	if surface, err := memphis.GetSurface(d.ObjectID()); err != nil {
		d.LogErr(err)
	} else {
		surface.Resize(size, d.GetTheme().Content.Normal)
	}
	d.Unlock()
	for _, window := range windows {
		d.Lock()
		if s, err := memphis.GetSurface(window.ObjectID()); err != nil {
			d.LogErr(err)
		} else {
			s.Resize(size, d.GetTheme().Content.Normal)
		}
		d.Unlock()
	}
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
	d.Emit(SignalFocusedWindow, d, w)
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
	if d.screen != nil {
		width, height = d.screen.Size()
	}
	d.RUnlock()
	region := ptypes.MakeRegion(0, 0, width, height)
	d.MapWindowWithRegion(w, region)
}

func (d *CDisplay) MapWindowWithRegion(w Window, region ptypes.Region) {
	d.LogDebug("mapping window: %v, with region: %v", w.ObjectName(), region)
	index := d.findMappedWindowIndex(w)
	w.SetDisplay(d)
	if s, err := memphis.GetSurface(w.ObjectID()); err != nil {
		if err := memphis.RegisterSurface(w.ObjectID(), region.Origin(), region.Size(), w.GetTheme().Content.Normal); err != nil {
			d.LogErr(err)
		}
	} else {
		s.SetOrigin(region.Origin())
		s.Resize(region.Size(), d.GetTheme().Content.Normal)
	}
	d.Lock()
	if index > -1 {
		d.windows = append(d.windows[:index], d.windows[index+1:]...)
	}
	d.windows = append([]Window{w}, d.windows...)
	d.Unlock()
	w.Emit(SignalMappedWindow, d)
}

func (d *CDisplay) UnmapWindow(w Window) {
	if idx := d.findMappedWindowIndex(w); idx > -1 {
		d.LogDebug("unmapping window: %v", w.ObjectName())
		d.Lock()
		memphis.RemoveSurface(w.ObjectID())
		d.windows = append(d.windows[:idx], d.windows[idx+1:]...)
		d.Unlock()
		w.Emit(SignalUnmappedWindow, d)
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
	d.RLock()
	position = d.cursor.Clone()
	moving = d.cursorMoving
	d.RUnlock()
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
	if d.eventFocus != nil {
		if sensitive, ok := d.eventFocus.Self().(Sensitive); ok {
			return sensitive.ProcessEvent(evt)
		}
		d.LogError("event focus does not implement Sensitive: %v (%T)", d.eventFocus, d.eventFocus)
		return enums.EVENT_PASS
	}
	switch e := evt.(type) {
	case *EventError:
		d.LogError("EventError: %v", e)
		if w := d.FocusedWindow(); w != nil {
			if f := w.ProcessEvent(evt); f == enums.EVENT_STOP {
				return enums.EVENT_STOP
			}
		}
		return d.Emit(SignalEventError, d, e)
	case *EventKey:
		if d.captureCtrlC {
			switch e.Rune() {
			case rune(KeyCtrlC):
				d.LogTrace("display captured CtrlC")
				if f := d.Emit(SignalInterrupt, d); f == enums.EVENT_STOP {
					return enums.EVENT_STOP
				}
				d.RequestQuit()
				return enums.EVENT_STOP
			}
		}
		if w := d.FocusedWindow(); w != nil {
			if f := w.ProcessEvent(evt); f == enums.EVENT_STOP {
				return enums.EVENT_STOP
			}
		}
		return d.Emit(SignalEventKey, d, e)
	case *EventMouse:
		d.Lock()
		d.cursor.Set(e.Position())
		d.cursorMoving = e.IsMoving() || e.IsDragging()
		d.Unlock()
		if w := d.FocusedWindow(); w != nil {
			if f := w.ProcessEvent(evt); f == enums.EVENT_STOP {
				return enums.EVENT_STOP
			}
		}
		return d.Emit(SignalEventMouse, d, e)
	case *EventResize:
		// all windows get resize event
		d.resizeWindows()
		stopped := false
		for _, window := range d.GetWindows() {
			if f := window.ProcessEvent(evt); f == enums.EVENT_STOP {
				stopped = true
			}
		}
		if stopped {
			d.RequestDraw()
			d.RequestShow()
		}
		return d.Emit(SignalEventResize, d, e)
	}
	if w := d.FocusedWindow(); w != nil {
		if f := w.ProcessEvent(evt); f == enums.EVENT_STOP {
			return enums.EVENT_STOP
		}
	}
	return d.Emit(SignalEvent, d, evt)
}

func (d *CDisplay) renderScreen() enums.EventFlag {
	if !d.DisplayCaptured() {
		return enums.EVENT_PASS
	}
	d.drawMutex.Lock()
	defer d.drawMutex.Unlock()
	d.Lock()
	windows := d.windows
	if surface, err := memphis.GetSurface(d.ObjectID()); err == nil {
		for i := len(windows) - 1; i >= 0; i-- {
			if wsurface, err := memphis.GetSurface(windows[i].ObjectID()); err == nil {
				if f := windows[i].Draw(); f == enums.EVENT_STOP {
					if err := surface.CompositeSurface(wsurface); err != nil {
						d.LogErr(err)
					}
				}
			} else {
				d.LogError("missing surface for window: %v", windows[i].ObjectID())
			}
		}
		if err := surface.Render(d.screen); err != nil {
			d.LogErr(err)
		}
		d.Unlock()
		return enums.EVENT_STOP
	}
	d.Unlock()
	d.LogError("missing surface for display: %v", d.ObjectID())
	return enums.EVENT_PASS
}

// RequestDraw asks the Display to process a SignalDraw event cycle, this does
// not actually render the contents to in Screen, just update
func (d *CDisplay) RequestDraw() {
	_ = d.AsyncCall(func(_ Display) error {
		if d.IsRunning() {
			d.requests <- displayDrawRequest
		} else {
			log.TraceF("application not running")
		}
		return nil
	})
}

// RequestShow asks the Display to render pending Screen changes
func (d *CDisplay) RequestShow() {
	_ = d.AsyncCall(func(_ Display) error {
		if d.IsRunning() {
			d.requests <- displayShowRequest
		} else {
			log.TraceF("application not running")
		}
		return nil
	})
}

// RequestSync asks the Display to render everything in the Screen
func (d *CDisplay) RequestSync() {
	_ = d.AsyncCall(func(_ Display) error {
		if d.IsRunning() {
			d.requests <- displaySyncRequest
		} else {
			log.TraceF("application not running")
		}
		return nil
	})
}

// RequestQuit asks the Display to quit nicely
func (d *CDisplay) RequestQuit() {
	_ = d.AsyncCall(func(_ Display) error {
		if d.IsRunning() {
			d.requests <- displayQuitRequest
		} else {
			log.TraceF("application not running")
		}
		return nil
	})
}

// IsRunning returns TRUE if the main thread is currently running.
func (d *CDisplay) IsRunning() bool {
	d.runLock.RLock()
	defer d.runLock.RUnlock()
	return d.running
}

func (d *CDisplay) setRunning(isRunning bool) {
	d.runLock.Lock()
	defer d.runLock.Unlock()
	d.running = isRunning
}

// StartupComplete emits SignalStartupComplete
func (d *CDisplay) StartupComplete() {
	d.Emit(SignalStartupComplete)
}

// AsyncCall runs the given DisplayCallbackFn on the UI thread, non-blocking
func (d *CDisplay) AsyncCall(fn DisplayCallbackFn) error {
	if !d.IsRunning() {
		return fmt.Errorf("application not running")
	}
	d.queue <- fn
	d.requests <- displayFuncRequest
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
	d.requests <- displayFuncRequest
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
			select {
			case <-ctx.Done():
				break processEventWorkerLoop
			default: // nop
			}
			if evt != nil {
				switch t := evt.(type) {
				default:
					d.Lock()
					d.buffer = append(d.buffer, t)
					d.Unlock()
				}
			} else {
				// nil event, quit?
				break
			}
		}
	}
}

func (d *CDisplay) startedAndCaptured() bool {
	d.RLock()
	defer d.RUnlock()
	return d.started && d.captured && d.screen != nil
}

func (d *CDisplay) screenRequestWorker(ctx context.Context) {
	// this happens in its own go thread
screenRequestWorkerLoop:
	for d.IsRunning() {
		var buffered []displayRequest
		max := len(d.requests)
		for i := 0; i < max; i++ {
			r := <-d.requests
			found := false
			for j := 0; j < len(buffered); j++ {
				if buffered[j] == r {
					found = true
				}
			}
			if !found {
				buffered = append(buffered, r)
			}
		}
		for _, r := range buffered {
			switch r {
			case displayDrawRequest:
				if d.startedAndCaptured() {
					d.renderScreen()
				}
			case displayShowRequest:
				if d.startedAndCaptured() {
					d.screen.Show()
				}
			case displaySyncRequest:
				if d.startedAndCaptured() {
					d.screen.Sync()
				}
			case displayFuncRequest:
				// one displayFuncRequest per d.queue fn
				if d.DisplayCaptured() {
					qlen := len(d.queue)
					for i := 0; i < qlen; i++ {
						if fn, ok := <-d.queue; ok {
							if err := fn(d); err != nil {
								log.ErrorF("async/await handler error: %v", err)
							}
						}
					}
				}
			case displayQuitRequest:
				d.done <- true
				break screenRequestWorkerLoop
			}
			select {
			case <-ctx.Done():
				break screenRequestWorkerLoop
			default: // nop
			}
		}
		time.Sleep(MainIterateDelay)
	}
}

// Run is the standard means of invoking a Display instance. It calls Startup,
// handles the main event look and finally calls MainFinish when all is
// complete.
func (d *CDisplay) Run() (err error) {
	ctx, _, wg := d.Startup()
	wg.Add(1)
	Go(func() {
		for d.HasPendingEvents() {
			d.IterateBufferedEvents()
			select {
			case <-ctx.Done():
				wg.Done()
				return
			default:
				// nop
				time.Sleep(MainIterateDelay)
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
func (d *CDisplay) Startup() (ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	if err := d.CaptureDisplay(); err != nil {
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
	wg.Add(1)
	Go(func() {
		d.screenRequestWorker(ctx)
		wg.Done()
	})
	_ = d.AsyncCall(func(_ Display) error {
		d.Emit(SignalDisplayStartup, ctx, cancel, wg)
		return nil
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
			mlen := len(d.mains)
			for i := 0; i < mlen; i++ {
				if fn, ok := <-d.mains; ok {
					if err := fn(d); err != nil {
						log.Error(err)
					}
				}
			}
			// guarantee ui calls
			qlen := len(d.queue)
			for i := 0; i < qlen; i++ {
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
	hasEvents = len(d.buffer) > 0
	d.RUnlock()
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
	buffer := make([]interface{}, len(d.buffer))
	for idx, e := range d.buffer {
		buffer[idx] = e
	}
	d.buffer = make([]interface{}, 0)
	d.Unlock()
	var pending []interface{}
	if d.GetCompressEvents() {
		pending = make([]interface{}, 0)
		history := make(map[string]int)
		for _, e := range buffer {
			switch t := e.(type) {
			default:
				key := fmt.Sprintf("%T", t)
				if idx, ok := history[key]; ok {
					if idx < len(pending) {
						pending = append(pending[:idx], pending[idx+1:]...)
					}
				}
				pending = append(pending, e)
				history[key] = len(pending) - 1
			}
		}
	} else {
		pending = buffer
	}
	stopped := false
	for _, e := range pending {
		if evt, ok := e.(Event); ok {
			if f := d.ProcessEvent(evt); f == enums.EVENT_STOP {
				stopped = true
			}
		}
	}
	if stopped {
		d.RequestDraw()
		d.RequestShow()
		return true
	}
	return false
}

type displayRequest uint64

const (
	displayNullRequest displayRequest = 1 << iota
	displayDrawRequest
	displayShowRequest
	displaySyncRequest
	displayFuncRequest
	displayQuitRequest
)

const (
	SignalDisplayCaptured     Signal = "display-captured"
	SignalInterrupt           Signal = "sigint"
	SignalEvent               Signal = "event"
	SignalEventError          Signal = "event-error"
	SignalEventKey            Signal = "event-key"
	SignalEventMouse          Signal = "event-mouse"
	SignalEventResize         Signal = "event-resize"
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
