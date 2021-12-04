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
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
	"github.com/go-curses/cdk/memphis"
)

const (
	TypeWindow       CTypeTag = "cdk-window"
	SignalDraw       Signal   = "draw"
	SignalSetTitle   Signal   = "set-title"
	SignalSetDisplay Signal   = "set-display"
)

func init() {
	_ = TypesManager.AddType(TypeWindow, nil)
}

// Basic window interface
type Window interface {
	Object

	Init() bool
	Destroy()
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
	style := paint.DefaultColorStyle
	if display := GetDefaultDisplay(); display != nil {
		style = display.GetTheme().Content.Normal
	}
	if err := memphis.RegisterSurface(w.ObjectID(), ptypes.Point2I{}, ptypes.Rectangle{}, style); err != nil {
		w.LogErr(err)
	}
	return false
}

func (w *CWindow) Destroy() {
	memphis.RemoveSurface(w.ObjectID())
	w.CObject.Destroy()
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
	w.Lock()
	var err error
	var surface *memphis.CSurface
	if surface, err = memphis.GetSurface(w.ObjectID()); err != nil {
		w.Unlock()
		w.LogErr(err)
		return enums.EVENT_PASS
	}
	w.Unlock()
	return w.Emit(SignalDraw, w, surface)
}

func (w *CWindow) ProcessEvent(evt Event) enums.EventFlag {
	return w.Emit(SignalEvent, w, evt)
}
