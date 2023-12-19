// Copyright (c) 2021-2023  The Go-Curses Authors
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

package main

import (
	"time"

	"github.com/go-curses/cdk"
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/ptypes"
	"github.com/go-curses/cdk/memphis"
)

const TypeMainWindow = cdk.CTypeTag("main-window")

func init() {
	_ = cdk.TypesManager.AddType(TypeMainWindow, nil)
}

type MainWindow struct {
	cdk.CWindow
}

func (w *MainWindow) Init() (already bool) {
	if w.InitTypeItem(TypeMainWindow, w) {
		return true
	}
	w.CWindow.Init()
	w.Connect(cdk.SignalDraw, "hello-draw-handler", w.draw)
	w.Connect(cdk.SignalEvent, "hello-event-handler", w.event)
	return false
}

func (w *MainWindow) draw(data []interface{}, argv ...interface{}) enums.EventFlag {
	if surface, ok := argv[1].(memphis.Surface); ok {
		theme := w.GetTheme()
		size := surface.GetSize()
		surface.Box(ptypes.Point2I{}, size, true, true, false, ' ', theme.Content.Normal, theme.Border.Normal, theme.Border.BorderRunes)
		content := "<b><span foreground=\"darkgreen\" background=\"yellow\"><u>H</u>ello</span> <span foreground=\"brown\" background=\"orange\"><i>W</i>orld</span></b>\n"
		content += "<span foreground=\"cyan\" background=\"gray\">(press CTRL+c or ESC to exit)</span>\n"
		content += "<span foreground=\"yellow\" background=\"darkblue\">" + time.Now().Format("2006-01-02 15:04:05") + "</span>"
		textPoint := ptypes.MakePoint2I(size.W/2/2, size.H/2-1)
		textSize := ptypes.MakeRectangle(size.W/2, size.H/2)
		surface.DrawText(textPoint, textSize, enums.JUSTIFY_CENTER, false, enums.WRAP_WORD, false, theme.Content.Normal, true, true, content)
		return enums.EVENT_STOP
	}
	return enums.EVENT_PASS
}

func (w *MainWindow) event(data []interface{}, argv ...interface{}) enums.EventFlag {
	if evt, ok := argv[1].(cdk.Event); ok {
		switch v := evt.(type) {
		case *cdk.EventError:
			w.LogInfo("ProcessEvent: Error (error:%v)", v.Err())
		case *cdk.EventKey:
			if v.Key() == cdk.KeyESC {
				w.LogInfo("ProcessEvent: RequestQuit (key:%v)", v.Name())
				cdk.GetDefaultDisplay().RequestQuit()
			} else {
				w.LogInfo("ProcessEvent: Key (key:%v)", v.Name())
			}
		case *cdk.EventMouse:
			w.LogInfo("ProcessEvent: Mouse (moving:%v, pressed:%v)", v.IsMoving(), v.IsPressed())
		case *cdk.EventResize:
			width, height := v.Size()
			w.LogInfo("ProcessEvent: Resize (width:%v, height:%v)", width, height)
		default:
			w.LogInfo("ProcessEvent: %T %v", v, v)
		}
		return enums.EVENT_STOP
	}
	return enums.EVENT_PASS
}
