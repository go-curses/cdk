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
	"os"
	"sync"
	"time"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
	"github.com/go-curses/cdk/log"
	"github.com/go-curses/cdk/memphis"
	"github.com/gofrs/uuid"
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
	IsMonochrome() bool
	Colors() (numberOfColors int)
	CaptureCtrlC()
	ReleaseCtrlC()
	DefaultTheme() paint.Theme
	ActiveWindow() Window
	SetActiveWindow(w Window)
	AddWindow(w Window)
	RemoveWindow(wid uuid.UUID)
	AddWindowOverlay(pid uuid.UUID, overlay Window, region ptypes.Region)
	RemoveWindowOverlay(pid, oid uuid.UUID)
	GetWindows() (windows []Window)
	GetWindowOverlays(id uuid.UUID) (windows []Window)
	GetWindowTopOverlay(id uuid.UUID) (window Window)
	GetWindowOverlayRegion(windowId, overlayId uuid.UUID) (region ptypes.Region)
	SetWindowOverlayRegion(windowId, overlayId uuid.UUID, region ptypes.Region)
	SetEventFocus(widget Object) error
	GetEventFocus() (widget Object)
	GetPriorEvent() (event Event)
	ProcessEvent(evt Event) enums.EventFlag
	DrawScreen() enums.EventFlag
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

	active uuid.UUID
	// windows map[uuid.UUID]*cWindowCanvas
	// overlay map[uuid.UUID][]*cWindowCanvas
	windows map[uuid.UUID]Window
	overlay map[uuid.UUID][]Window

	app        *CApplication
	ttyPath    string
	ttyHandle  *os.File
	screen     Screen
	captured   bool
	started    bool
	eventFocus Object
	priorEvent Event

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

	d.priorEvent = nil
	d.eventFocus = nil
	d.windows = make(map[uuid.UUID]Window)
	d.overlay = make(map[uuid.UUID][]Window)
	d.active = uuid.Nil
	d.SetTheme(paint.DefaultColorTheme)

	d.runLock = &sync.RWMutex{}
	d.eventMutex = &sync.Mutex{}
	d.drawMutex = &sync.Mutex{}
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
	return d.title
}

func (d *CDisplay) SetTitle(title string) {
	d.title = title
}

func (d *CDisplay) GetTtyPath() string {
	return d.ttyPath
}

func (d *CDisplay) SetTtyPath(ttyPath string) {
	d.ttyPath = ttyPath
}

func (d *CDisplay) GetTtyHandle() *os.File {
	return d.ttyHandle
}

func (d *CDisplay) SetTtyHandle(ttyHandle *os.File) {
	d.ttyHandle = ttyHandle
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
	return d.screen
}

func (d *CDisplay) DisplayCaptured() bool {
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
	if d.captured {
		if d.screen != nil {
			d.screen.Close()
			d.screen = nil
		}
		d.captured = false
	}
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

func (d *CDisplay) ResizeWindows() {
	d.RLock()
	if d.screen == nil {
		d.RUnlock()
		return
	}
	windows := d.windows
	w, h := d.screen.Size()
	d.RUnlock()
	for _, window := range windows {
		size := ptypes.MakeRectangle(w, h)
		if s, err := memphis.GetSurface(window.ObjectID()); err != nil {
			d.LogErr(err)
		} else {
			s.Resize(size, d.GetTheme().Content.Normal)
		}
		_ = window.ProcessEvent(NewEventResize(w, h))
	}
}

func (d *CDisplay) ActiveWindow() Window {
	d.RLock()
	defer d.RUnlock()
	if w, ok := d.windows[d.active]; ok {
		return w
	}
	return nil
}

func (d *CDisplay) SetActiveWindow(w Window) {
	d.RLock()
	if _, ok := d.windows[w.ObjectID()]; !ok {
		d.RUnlock()
		d.AddWindow(w)
	} else {
		d.RUnlock()
	}
	d.Lock()
	d.active = w.ObjectID()
	d.Unlock()
	d.ResizeWindows()
}

func (d *CDisplay) AddWindow(w Window) {
	if _, ok := d.windows[w.ObjectID()]; ok {
		d.LogWarn("window already added to display: %v", w.ObjectName())
		return
	}
	w.SetDisplay(d)
	size := ptypes.MakeRectangle(0, 0)
	d.RLock()
	if d.screen != nil {
		size = ptypes.MakeRectangle(d.screen.Size())
	}
	d.RUnlock()
	if s, err := memphis.GetSurface(w.ObjectID()); err != nil {
		w.LogErr(err)
	} else {
		s.Resize(size, d.GetTheme().Content.Normal)
	}
	d.Lock()
	d.windows[w.ObjectID()] = w
	d.overlay[w.ObjectID()] = nil
	d.Unlock()
}

func (d *CDisplay) RemoveWindow(wid uuid.UUID) {
	d.Lock()
	defer d.Unlock()
	if _, ok := d.windows[wid]; ok {
		delete(d.windows, wid)
	}
	if _, ok := d.overlay[wid]; ok {
		delete(d.overlay, wid)
	}
}

func (d *CDisplay) AddWindowOverlay(pid uuid.UUID, overlay Window, region ptypes.Region) {
	d.Lock()
	if _, ok := d.overlay[pid]; !ok {
		d.overlay[pid] = make([]Window, 0)
	}
	if err := memphis.ConfigureSurface(overlay.ObjectID(), region.Origin(), region.Size(), d.GetTheme().Content.Normal); err != nil {
		overlay.LogErr(err)
	}
	d.overlay[pid] = append(d.overlay[pid], overlay)
	d.Unlock()
}

func (d *CDisplay) RemoveWindowOverlay(pid, oid uuid.UUID) {
	d.Lock()
	if wc, ok := d.overlay[pid]; ok {
		var revised []Window
		for _, oc := range wc {
			if oc.ObjectID() != oid {
				revised = append(revised, oc)
			}
		}
		d.overlay[pid] = revised
	}
	d.Unlock()
}

func (d *CDisplay) GetWindows() (windows []Window) {
	d.RLock()
	defer d.RUnlock()
	for _, w := range d.windows {
		windows = append(windows, w)
	}
	return
}

func (d *CDisplay) GetWindowOverlays(id uuid.UUID) (windows []Window) {
	d.RLock()
	defer d.RUnlock()
	if overlays, ok := d.overlay[id]; ok {
		for _, overlay := range overlays {
			windows = append(windows, overlay)
		}
	}
	return
}

func (d *CDisplay) GetWindowTopOverlay(id uuid.UUID) (window Window) {
	d.RLock()
	defer d.RUnlock()
	if overlays, ok := d.overlay[id]; ok {
		if last := len(overlays) - 1; last > -1 {
			window = overlays[last]
		}
	}
	return
}

func (d *CDisplay) GetWindowOverlayRegion(windowId, overlayId uuid.UUID) (region ptypes.Region) {
	d.RLock()
	defer d.RUnlock()
	if overlays, ok := d.overlay[windowId]; ok {
		for _, overlay := range overlays {
			if overlay.ObjectID() == overlayId {
				if s, err := memphis.GetSurface(overlay.ObjectID()); err != nil {
					overlay.LogErr(err)
				} else {
					origin := s.GetOrigin()
					size := s.GetSize()
					region = ptypes.MakeRegion(origin.X, origin.Y, size.W, size.H)
				}
				break
			}
		}
	} else {
		d.LogError("window not found: %v", windowId)
	}
	return
}

func (d *CDisplay) SetWindowOverlayRegion(windowId, overlayId uuid.UUID, region ptypes.Region) {
	d.Lock()
	defer d.Unlock()
	if overlays, ok := d.overlay[windowId]; ok {
		for _, overlay := range overlays {
			if overlay.ObjectID() == overlayId {
				if err := memphis.ConfigureSurface(overlay.ObjectID(), region.Origin(), region.Size(), d.GetTheme().Content.Normal); err != nil {
					overlay.LogErr(err)
				}
				break
			}
		}
	} else {
		d.LogError("window not found: %v", windowId)
	}
}

func (d *CDisplay) getOverlay(windowId uuid.UUID) (overlay Window) {
	d.RLock()
	defer d.RUnlock()
	if overlays, ok := d.overlay[windowId]; ok {
		if last := len(overlays) - 1; last > -1 {
			overlay = overlays[last]
		}
	}
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
	d.eventMutex.Lock()
	var overlayWindow Window
	if w := d.ActiveWindow(); w != nil {
		if overlay := d.getOverlay(w.ObjectID()); overlay != nil {
			overlayWindow = overlay
		}
	}
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
	} else if overlayWindow != nil {
		switch e := evt.(type) {
		case *EventResize:
			if w := d.ActiveWindow(); w != nil {
				if ac, err := memphis.GetSurface(w.ObjectID()); err != nil {
					alloc := ptypes.MakeRectangle(d.screen.Size())
					ac.Resize(alloc, d.GetTheme().Content.Normal)
					if f := w.ProcessEvent(e); f == enums.EVENT_STOP {
						d.RequestDraw()
						d.RequestSync()
					}
				}
			}
		}
		return overlayWindow.ProcessEvent(evt)
	}
	switch e := evt.(type) {
	case *EventError:
		d.LogErr(e)
		if w := d.ActiveWindow(); w != nil {
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
		if w := d.ActiveWindow(); w != nil {
			if f := w.ProcessEvent(evt); f == enums.EVENT_STOP {
				return enums.EVENT_STOP
			}
		}
		return d.Emit(SignalEventKey, d, e)
	case *EventMouse:
		if w := d.ActiveWindow(); w != nil {
			if f := w.ProcessEvent(evt); f == enums.EVENT_STOP {
				return enums.EVENT_STOP
			}
		}
		return d.Emit(SignalEventMouse, d, e)
	case *EventResize:
		if aw := d.ActiveWindow(); aw != nil {
			if ac, err := memphis.GetSurface(aw.ObjectID()); err != nil {
				aw.LogErr(err)
			} else {
				alloc := ptypes.MakeRectangle(d.screen.Size())
				ac.Resize(alloc, d.GetTheme().Content.Normal)
				if f := aw.ProcessEvent(evt); f == enums.EVENT_STOP {
					d.RequestDraw()
					d.RequestSync()
					return enums.EVENT_STOP
				}
			}
		}
		return d.Emit(SignalEventResize, d, e)
	}
	if w := d.ActiveWindow(); w != nil {
		if f := w.ProcessEvent(evt); f == enums.EVENT_STOP {
			return enums.EVENT_STOP
		}
	}
	return d.Emit(SignalEvent, d, evt)
}

// DrawScreen renders the active window contents to the screen
func (d *CDisplay) DrawScreen() enums.EventFlag {
	d.drawMutex.Lock()
	defer d.drawMutex.Unlock()
	d.RLock()
	if !d.captured || d.screen == nil {
		d.RUnlock()
		d.LogError("display not captured or is otherwise missing")
		return enums.EVENT_PASS
	}
	d.RUnlock()
	if aw := d.ActiveWindow(); aw != nil {
		if ac, err := memphis.GetSurface(aw.ObjectID()); err == nil {
			if f := aw.Draw(); f == enums.EVENT_STOP {
				if overlays, ok := d.overlay[aw.ObjectID()]; ok {
					for _, overlay := range overlays {
						if of := overlay.Draw(); of == enums.EVENT_STOP {
							if s, err := memphis.GetSurface(overlay.ObjectID()); err != nil {
								overlay.LogErr(err)
							} else {
								if err := ac.CompositeSurface(s); err != nil {
									d.LogErr(err)
								}
							}
						}
					}
				}
				if err := ac.Render(d.screen); err != nil {
					d.LogErr(err)
				}
				return enums.EVENT_STOP
			}
		} else {
			d.LogError("missing surface for active window: %v", aw.ObjectID())
		}
	} else {
		d.LogError("active window not found")
	}
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
	for d.IsRunning() && d.screen != nil {
		select {
		case evt := <-d.screen.PollEventChan():
			d.inbound <- evt
		case <-ctx.Done():
			break pollEventWorkerLoop
			// default: // nop
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

func (d *CDisplay) screenRequestWorker(ctx context.Context) {
	// this happens in its own go thread
screenRequestWorkerLoop:
	for d.IsRunning() {
		switch <-d.requests {
		case displayDrawRequest:
			d.RLock()
			if d.started && d.screen != nil {
				d.RUnlock()
				d.DrawScreen()
			} else {
				d.RUnlock()
			}
		case displayShowRequest:
			d.RLock()
			if d.started && d.screen != nil {
				d.RUnlock()
				d.screen.Show()
			} else {
				d.RUnlock()
			}
		case displaySyncRequest:
			d.RLock()
			if d.started && d.screen != nil {
				d.RUnlock()
				d.screen.Sync()
			} else {
				d.RUnlock()
			}
		case displayFuncRequest:
			// one displayFuncRequest per d.queue fn
			qlen := len(d.queue)
			for i := 0; i < qlen; i++ {
				if fn, ok := <-d.queue; ok {
					if err := fn(d); err != nil {
						log.ErrorF("async/await handler error: %v", err)
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
				if err := fn(d); err != nil {
					log.Error(err)
				}
			}
		case evt, ok := <-d.events:
			if ok {
				if d.screen != nil {
					if err := d.screen.PostEvent(evt); err != nil {
						log.Error(err)
					}
				} else {
					d.LogTrace("missing screen, dropping event: %v", evt)
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
	SignalDisplayCaptured Signal = "display-captured"
	SignalInterrupt       Signal = "sigint"
	SignalEvent           Signal = "event"
	SignalEventError      Signal = "event-error"
	SignalEventKey        Signal = "event-key"
	SignalEventMouse      Signal = "event-mouse"
	SignalEventResize     Signal = "event-resize"
	SignalSetEventFocus   Signal = "set-event-focus"
	SignalStartupComplete Signal = "startup-complete"
	SignalDisplayStartup  Signal = "display-startup"
	SignalDisplayShutdown Signal = "display-shutdown"
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
