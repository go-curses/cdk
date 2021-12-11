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

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-curses/cdk"
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/ptypes"
	"github.com/go-curses/cdk/memphis"
)

const TypeAppWindow = cdk.CTypeTag("app-window")

func init() {
	_ = cdk.TypesManager.AddType(TypeAppWindow, nil)
}

type AppWindow struct {
	cdk.CWindow
}

func (w *AppWindow) Init() (already bool) {
	if w.InitTypeItem(TypeAppWindow, w) {
		return true
	}
	w.CWindow.Init()
	w.Connect(cdk.SignalDraw, "hello-draw-handler", w.draw)
	w.Connect(cdk.SignalEvent, "hello-event-handler", w.event)
	return false
}

func (w *AppWindow) draw(data []interface{}, argv ...interface{}) enums.EventFlag {
	var err error
	var ctx *cdk.CLocalContextData
	if surface, ok := argv[1].(memphis.Surface); ok {
		theme := w.GetTheme()
		size := surface.GetSize()
		surface.Box(ptypes.Point2I{}, size, true, true, false, ' ', theme.Content.Normal, theme.Border.Normal, theme.Border.BorderRunes)
		title := w.GetTitle()
		surface.DrawText(
			ptypes.MakePoint2I((size.W/2)-(len(title)/2), 0),
			ptypes.MakeRectangle(len(title), 1),
			enums.JUSTIFY_CENTER, true, enums.WRAP_WORD, false,
			theme.Content.Normal, true, false,
			fmt.Sprintf("<style foreground=\"#ffffff\"><b>%s</b></style>", title),
		)
		content := "<b><span foreground=\"darkgreen\" background=\"yellow\"><u>H</u>ello</span> <span foreground=\"brown\" background=\"orange\"><i>W</i>orld</span></b>\n"
		content += "<span foreground=\"cyan\" background=\"gray\">(press CTRL+c or ESC to exit)</span>\n"
		content += "<span foreground=\"cyan\" background=\"gray\">(press b to bash shell)</span>\n"
		content += "<span foreground=\"cyan\" background=\"gray\">(press f to run a func)</span>\n"
		content += "<span foreground=\"yellow\" background=\"darkblue\">" + time.Now().Format("2006-01-02 15:04:05") + "</span>\n"
		if ctx, err = cdk.GetLocalContext(); err != nil {
			w.LogError("error getting local context: %v", err)
			content += "(missing app context)"
		} else {
			if asc, ok := ctx.Data.(*cdk.CApplicationServer); ok {
				content += fmt.Sprintf("listening on %s:%d\n", asc.GetListenAddress(), asc.GetListenPort())
				clients := asc.GetClients()
				numClients := len(clients)
				if numClients == 0 {
					content += "<u>no clients connected</u>\n"
				} else {
					content += "<u>connected clients:</u>\n"
					for _, id := range clients {
						if client, err := asc.GetClient(id); err != nil {
							w.LogError("error getting app server client data: %v", err)
							continue
						} else {
							content += client.String() + "\n"
						}
					}
				}
			}
		}
		textPoint := ptypes.MakePoint2I(size.W/2/2, size.H/2-1)
		textSize := ptypes.MakeRectangle(size.W/2, size.H/2)
		surface.DrawText(textPoint, textSize, enums.JUSTIFY_CENTER, false, enums.WRAP_WORD, false, theme.Content.Normal, true, true, content)
		return enums.EVENT_STOP
	}
	return enums.EVENT_PASS
}

func (w *AppWindow) event(data []interface{}, argv ...interface{}) enums.EventFlag {
	display := cdk.GetDefaultDisplay()
	if evt, ok := argv[1].(cdk.Event); ok {
		switch v := evt.(type) {
		case *cdk.EventError:
			w.LogInfo("ProcessEvent: Error (error:%v)", v.Err())
		case *cdk.EventKey:
			if v.Key() == cdk.KeyESC {
				w.LogInfo("ProcessEvent: RequestQuit (key:%v)", v.Name())
				display.RequestQuit()
			} else if v.Rune() == rune('f') {
				w.LogInfo("ProcessEvent: Call func (key:%v)", v.Name())
				fn := func(_, tty *os.File) (err error) {
					w.LogDebug("fn is running!")
					_, _ = fmt.Fprintf(tty, "# waiting for 5 seconds #\r\n")
					for _, i := range []int{5, 4, 3, 2, 1} {
						_, _ = fmt.Fprintf(tty, "# %v...\r\n", i)
						time.Sleep(time.Second)
					}
					_, _ = fmt.Fprintf(tty, "# returning to hellocall now! #\r\n")
					time.Sleep(time.Millisecond * 500)
					return nil
				}
				if err := display.Call(fn); err != nil {
					w.LogErr(err)
				}
			} else if v.Rune() == rune('b') {
				w.LogInfo("ProcessEvent: Call bash (key:%v)", v.Name())
				if err := display.Command("/bin/bash", "-l"); err != nil {
					w.LogErr(err)
				}
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
