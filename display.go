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
	"github.com/jtolio/gls"
)

// display is really a wrapper around Screen
// and Simulation screens

// basically a wrapper around Screen()
// manages one or more windows backed by viewports
// viewports manage the allocation of space
// drawables within viewports render the space

// Global configuration variables
var (
	// DisplayCallQueueCapacity limits the number of concurrent calls on the main thread
	DisplayCallQueueCapacity = 16
	// DisplayStartupDelay is the delay for triggering screen resize events during initialization
	DisplayStartupDelay = time.Millisecond * 128
)

const (
	TypeDisplayManager    CTypeTag = "cdk-display-manager"
	SignalDisplayInit     Signal   = "display-init"
	SignalDisplayCaptured Signal   = "display-captured"
	SignalInterrupt       Signal   = "sigint"
	SignalEvent           Signal   = "event"
	SignalEventError      Signal   = "event-error"
	SignalEventKey        Signal   = "event-key"
	SignalEventMouse      Signal   = "event-mouse"
	SignalEventResize     Signal   = "event-resize"
	SignalSetEventFocus   Signal   = "set-event-focus"
)

func init() {
	_ = TypesManager.AddType(TypeDisplayManager, nil)
}

type DisplayCallbackFn = func(d Display) error

type Display interface {
	Object

	Init() (already bool)
	Destroy()
	GetTitle() string
	SetTitle(title string)
	GetTtyPath() string
	SetTtyPath(ttyPath string)
	GetTtyHandle() *os.File
	SetTtyHandle(ttyHandle *os.File)
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
	App() *CApp
	SetEventFocus(widget interface{}) error
	GetEventFocus() (widget interface{})
	GetPriorEvent() (event Event)
	ProcessEvent(evt Event) enums.EventFlag
	DrawScreen() enums.EventFlag
	RequestDraw()
	RequestShow()
	RequestSync()
	RequestQuit()
	AsyncCall(fn DisplayCallbackFn) error
	AwaitCall(fn DisplayCallbackFn) error
	PostEvent(evt Event) error
	AddQuitHandler(tag string, fn func())
	RemoveQuitHandler(tag string)
	Run() error
	IsRunning() bool
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

	app        *CApp
	ttyPath    string
	ttyHandle  *os.File
	screen     Screen
	captured   bool
	eventFocus interface{}
	priorEvent Event
	quitters   map[string]func()

	runLock  *sync.RWMutex
	running  bool
	warmup   bool
	closing  sync.Once
	done     chan bool
	queue    chan DisplayCallbackFn
	events   chan Event
	process  chan Event
	requests chan ScreenStateReq

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
	d.running = false
	d.runLock = &sync.RWMutex{}
	d.warmup = true
	d.done = make(chan bool)
	d.queue = make(chan DisplayCallbackFn, DisplayCallQueueCapacity)
	d.events = make(chan Event, DisplayCallQueueCapacity)
	d.process = make(chan Event, DisplayCallQueueCapacity)
	d.requests = make(chan ScreenStateReq, DisplayCallQueueCapacity)

	d.priorEvent = nil
	d.eventFocus = nil
	d.quitters = make(map[string]func())
	d.windows = make(map[uuid.UUID]Window)
	d.overlay = make(map[uuid.UUID][]Window)
	d.active = uuid.Nil
	d.SetTheme(paint.DefaultColorTheme)

	d.eventMutex = &sync.Mutex{}
	d.drawMutex = &sync.Mutex{}

	d.Emit(SignalDisplayInit, d)
	return false
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
		close(d.process)
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
	if d.screen != nil {
		if d.screen.Colors() <= 0 {
			return paint.DefaultMonoTheme
		}
	}
	return paint.DefaultColorTheme
}

func (d *CDisplay) ActiveWindow() Window {
	if w, ok := d.windows[d.active]; ok {
		return w
	}
	return nil
}

func (d *CDisplay) SetActiveWindow(w Window) {
	if _, ok := d.windows[w.ObjectID()]; !ok {
		d.AddWindow(w)
	}
	d.active = w.ObjectID()
}

func (d *CDisplay) AddWindow(w Window) {
	if _, ok := d.windows[w.ObjectID()]; ok {
		d.LogWarn("window already added to display: %v", w.ObjectName())
		return
	}
	w.SetDisplay(d)
	size := ptypes.MakeRectangle(0, 0)
	if d.screen != nil {
		size = ptypes.MakeRectangle(d.screen.Size())
	}
	if s, err := memphis.GetSurface(w.ObjectID()); err != nil {
		w.LogErr(err)
	} else {
		s.Resize(size, d.GetTheme().Content.Normal)
	}
	d.windows[w.ObjectID()] = w
	d.overlay[w.ObjectID()] = nil
}

func (d *CDisplay) RemoveWindow(wid uuid.UUID) {
	if _, ok := d.windows[wid]; ok {
		delete(d.windows, wid)
	}
	if _, ok := d.overlay[wid]; ok {
		delete(d.overlay, wid)
	}
}

func (d *CDisplay) AddWindowOverlay(pid uuid.UUID, overlay Window, region ptypes.Region) {
	if _, ok := d.overlay[pid]; !ok {
		d.overlay[pid] = make([]Window, 0)
	}
	if err := memphis.ConfigureSurface(overlay.ObjectID(), region.Origin(), region.Size(), d.GetTheme().Content.Normal); err != nil {
		overlay.LogErr(err)
	}
	d.overlay[pid] = append(d.overlay[pid], overlay)
}

func (d *CDisplay) RemoveWindowOverlay(pid, oid uuid.UUID) {
	if wc, ok := d.overlay[pid]; ok {
		var revised []Window
		for _, oc := range wc {
			if oc.ObjectID() != oid {
				revised = append(revised, oc)
			}
		}
		d.overlay[pid] = revised
	}
}

func (d *CDisplay) GetWindows() (windows []Window) {
	for _, w := range d.windows {
		windows = append(windows, w)
	}
	return
}

func (d *CDisplay) GetWindowOverlays(id uuid.UUID) (windows []Window) {
	if overlays, ok := d.overlay[id]; ok {
		for _, overlay := range overlays {
			windows = append(windows, overlay)
		}
	}
	return
}

func (d *CDisplay) GetWindowTopOverlay(id uuid.UUID) (window Window) {
	if overlays, ok := d.overlay[id]; ok {
		if last := len(overlays) - 1; last > -1 {
			window = overlays[last]
		}
	}
	return
}

func (d *CDisplay) GetWindowOverlayRegion(windowId, overlayId uuid.UUID) (region ptypes.Region) {
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
	if overlays, ok := d.overlay[windowId]; ok {
		if last := len(overlays) - 1; last > -1 {
			overlay = overlays[last]
		}
	}
	return
}

func (d *CDisplay) App() *CApp {
	return d.app
}

func (d *CDisplay) SetEventFocus(widget interface{}) error {
	if widget != nil {
		if _, ok := widget.(Sensitive); !ok {
			return fmt.Errorf("widget does not implement Sensitive: %v (%T)", widget, widget)
		}
	}
	if f := d.Emit(SignalSetEventFocus); f == enums.EVENT_PASS {
		d.eventFocus = widget
	}
	return nil
}

func (d *CDisplay) GetEventFocus() (widget interface{}) {
	widget = d.eventFocus
	return
}

func (d *CDisplay) GetPriorEvent() (event Event) {
	return d.priorEvent
}

func (d *CDisplay) ProcessEvent(evt Event) enums.EventFlag {
	d.eventMutex.Lock()
	var overlayWindow Window
	if w := d.ActiveWindow(); w != nil {
		if overlay := d.getOverlay(w.ObjectID()); overlay != nil {
			overlayWindow = overlay
		}
	}
	defer func() {
		d.priorEvent = evt
		d.eventMutex.Unlock()
	}()
	if d.eventFocus != nil {
		if sensitive, ok := d.eventFocus.(Sensitive); ok {
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
			switch e.Key() {
			case KeyCtrlC:
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

func (d *CDisplay) DrawScreen() enums.EventFlag {
	d.drawMutex.Lock()
	defer d.drawMutex.Unlock()
	if !d.captured || d.screen == nil {
		d.LogError("display not captured or is otherwise missing")
		return enums.EVENT_PASS
	}
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

func (d *CDisplay) RequestDraw() {
	if d.IsRunning() {
		d.requests <- DrawRequest
	} else {
		log.TraceF("application not running")
	}
}

func (d *CDisplay) RequestShow() {
	if d.IsRunning() {
		d.requests <- ShowRequest
	} else {
		log.TraceF("application not running")
	}
}

func (d *CDisplay) RequestSync() {
	if d.IsRunning() {
		d.requests <- SyncRequest
	} else {
		log.TraceF("application not running")
	}
}

func (d *CDisplay) RequestQuit() {
	_ = d.AwaitCall(func(_ Display) error {
		if d.IsRunning() {
			d.requests <- QuitRequest
		} else {
			log.TraceF("application not running")
		}
		return nil
	})
}

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

// AsyncCall - run given function on the UI (main) thread, non-blocking
func (d *CDisplay) AsyncCall(fn DisplayCallbackFn) error {
	if !d.IsRunning() {
		return fmt.Errorf("application not running")
	}
	d.queue <- fn
	return nil
}

// AwaitCall - run given function on the UI (main) thread, blocking
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

func (d *CDisplay) AddQuitHandler(tag string, fn func()) {
	if _, ok := d.quitters[tag]; ok {
		d.LogWarn("replacing quit handler: %v", tag)
	}
	d.quitters[tag] = fn
}

func (d *CDisplay) RemoveQuitHandler(tag string) {
	if _, ok := d.quitters[tag]; ok {
		delete(d.quitters, tag)
	}
}

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
			d.process <- evt
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
		case evt := <-d.process:
			select {
			case <-ctx.Done():
				break processEventWorkerLoop
			default: // nop
			}
			if evt != nil {
				if f := d.ProcessEvent(evt); f == enums.EVENT_STOP {
					// TODO: ProcessEvent must ONLY flag stop when UI changes
					d.RequestDraw()
					d.RequestShow()
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
	if err := d.app.InitUI(); err != nil {
		log.FatalDF(1, "%v", err)
	}
	// after a delay, post a resize event and request draw + show
	AddTimeout(DisplayStartupDelay, func() enums.EventFlag {
		if d.screen != nil {
			d.warmup = false
			if err := d.screen.PostEvent(NewEventResize(d.screen.Size())); err != nil {
				log.Error(err)
			} else {
				d.RequestDraw()
				d.RequestSync()
			}
		}
		return enums.EVENT_STOP
	})
screenRequestWorkerLoop:
	for d.IsRunning() {
		switch <-d.requests {
		case DrawRequest:
			if d.screen != nil && !d.warmup {
				d.DrawScreen()
			}
		case ShowRequest:
			if d.screen != nil && !d.warmup {
				d.screen.Show()
			}
		case SyncRequest:
			if d.screen != nil && !d.warmup {
				d.screen.Sync()
			}
		case QuitRequest:
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

func (d *CDisplay) Run() error {
	// this happens in the actual main thread
	if err := d.CaptureDisplay(); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	d.setRunning(true)
	wg.Add(1)
	gls.Go(func() {
		d.pollEventWorker(ctx)
		wg.Done()
	})
	wg.Add(1)
	gls.Go(func() {
		d.processEventWorker(ctx)
		wg.Done()
	})
	wg.Add(1)
	gls.Go(func() {
		d.screenRequestWorker(ctx)
		wg.Done()
	})
	defer func() {
		d.Destroy()
		if p := recover(); p != nil {
			panic(p)
		}
		for _, quitter := range d.quitters {
			quitter()
		}
	}()
	d.RequestDraw()
	d.RequestSync()
runLoop:
	for d.IsRunning() {
		select {
		case fn, ok := <-d.queue:
			if ok {
				if err := fn(d); err != nil {
					log.ErrorF("async/await handler error: %v", err)
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
			break runLoop
		}
	}
	cancel()  // notify threads to exit
	wg.Wait() // wait for all threads to exit
	if goProfile != nil {
		log.DebugF("stopping profiling")
		goProfile.Stop()
	}
	return nil
}

const PropertyDisplayName Property = "display-name"
const PropertyDisplayUser Property = "display-user"
const PropertyDisplayHost Property = "display-host"
