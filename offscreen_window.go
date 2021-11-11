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
	"github.com/go-curses/cdk/charset"
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/log"
	"github.com/go-curses/cdk/memphis"
)

const (
	TypeOffscreenWindow CTypeTag = "cdk-offscreen-window"
)

func init() {
	_ = TypesManager.AddType(TypeOffscreenWindow, nil)
}

// Basic window interface
type OffscreenWindow interface {
	Object

	GetTitle() string
	SetTitle(title string)

	GetDisplay() Display
	SetDisplay(d Display)

	Draw() enums.EventFlag
	ProcessEvent(evt Event) enums.EventFlag
}

// Basic window type
type COffscreenWindow struct {
	CObject

	title   string
	display OffScreen
}

func NewOffscreenWindow(title string) Window {
	d, err := MakeOffScreen(charset.Get())
	if err != nil {
		log.Fatal(err)
	}
	w := &COffscreenWindow{
		title:   title,
		display: d,
	}
	w.Init()
	return w
}

func (w *COffscreenWindow) Init() bool {
	if w.InitTypeItem(TypeWindow, w) {
		return true
	}
	w.CObject.Init()
	return false
}

func (w *COffscreenWindow) SetTitle(title string) {
	if f := w.Emit(SignalSetTitle, w, title); f == enums.EVENT_PASS {
		w.title = title
	}
}

func (w *COffscreenWindow) GetTitle() string {
	return w.title
}

func (w *COffscreenWindow) GetDisplay() Display {
	// return w.display
	return nil
}

func (w *COffscreenWindow) SetDisplay(d Display) {
	if f := w.Emit(SignalSetDisplay, w, d); f == enums.EVENT_PASS {
		// w.display = d
	}
}

func (w *COffscreenWindow) Draw() enums.EventFlag {
	if surface, err := memphis.GetSurface(w.ObjectID()); err != nil {
		w.LogErr(err)
	} else {
		return w.Emit(SignalDraw, w, surface)
	}
	return enums.EVENT_PASS
}

func (w *COffscreenWindow) ProcessEvent(evt Event) enums.EventFlag {
	return w.Emit(SignalEvent, w, evt)
}
