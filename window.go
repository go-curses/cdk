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

package cdk

import (
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/memphis"
)

const (
	TypeWindow         CTypeTag = "cdk-window"
	PropertyWindowType Property = "window-type"
	SignalDraw         Signal   = "draw"
	SignalSetTitle     Signal   = "set-title"
	SignalSetDisplay   Signal   = "set-display"
)

func init() {
	_ = TypesManager.AddType(TypeWindow, nil)
}

// Basic window interface
type Window interface {
	Object

	Init() bool
	Destroy()
	GetWindowType() (value enums.WindowType)
	SetWindowType(hint enums.WindowType)
	SetTitle(title string)
	GetTitle() string
	GetDisplay() Display
	SetDisplay(d Display)
	Draw() enums.EventFlag
	ProcessEvent(evt Event) enums.EventFlag
}

// Basic window type
type CWindow struct {
	CObject

	title   string
	display Display
}

func NewWindow(title string, d Display) Window {
	w := &CWindow{
		title:   title,
		display: d,
	}
	w.Init()
	return w
}

func (w *CWindow) Init() bool {
	if w.InitTypeItem(TypeWindow, w) {
		return true
	}
	w.CObject.Init()
	_ = w.InstallProperty(PropertyWindowType, StructProperty, true, enums.WINDOW_TOPLEVEL)
	return false
}

func (w *CWindow) Destroy() {
	if display := w.GetDisplay(); display != nil {
		display.UnmapWindow(w)
	}
	w.CObject.Destroy()
}

// GetWindowType returns the type of the window.
// See: enums.WindowType.
func (w *CWindow) GetWindowType() (value enums.WindowType) {
	var ok bool
	if v, err := w.GetStructProperty(PropertyWindowType); err != nil {
		w.LogErr(err)
	} else if value, ok = v.(enums.WindowType); !ok {
		value = enums.WINDOW_TOPLEVEL // default is top-level?
		w.LogError("value stored in %v is not of enums.WindowType: %v (%T)", PropertyWindowType, v, v)
	}
	return
}

// SetWindowType updates the type of the window.
// See: enums.WindowType
func (w *CWindow) SetWindowType(hint enums.WindowType) {
	if err := w.SetStructProperty(PropertyWindowType, hint); err != nil {
		w.LogErr(err)
	}
}

func (w *CWindow) SetTitle(title string) {
	if f := w.Emit(SignalSetTitle, w, title); f == enums.EVENT_PASS {
		w.Lock()
		w.title = title
		w.Unlock()
	}
}

func (w *CWindow) GetTitle() string {
	w.RLock()
	defer w.RUnlock()
	return w.title
}

func (w *CWindow) GetDisplay() Display {
	w.RLock()
	defer w.RUnlock()
	return w.display
}

func (w *CWindow) SetDisplay(d Display) {
	if f := w.Emit(SignalSetDisplay, w, d); f == enums.EVENT_PASS {
		w.Lock()
		w.display = d
		w.Unlock()
	}
}

func (w *CWindow) Draw() enums.EventFlag {
	if !w.IsFrozen() {
		if surface, err := memphis.GetSurface(w.ObjectID()); err != nil {
			w.LogErr(err)
		} else {
			return w.Emit(SignalDraw, w, surface)
		}
	}
	return enums.EVENT_PASS
}

func (w *CWindow) ProcessEvent(evt Event) enums.EventFlag {
	return w.Emit(SignalEvent, w, evt)
}
