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
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-curses/cdk"
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/ptypes"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
	"github.com/go-curses/cdk/memphis"
)

var CdkApp cdk.Application

func main() {
	// used when not built with -buildmode=plugin
	if err := CdkApp.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	CdkApp = cdk.NewApplication(
		"demoplugin",
		"An example CDK application plugin",
		"Demo Plugin is an example CDK application plugin",
		"0.0.1",
		"demoplugin",
		"Demo Plugin",
		"/dev/tty",
	)
	CdkApp.Connect(
		cdk.SignalStartup,
		"demoplugin-startup-handler",
		cdk.WithArgvApplicationSignalStartup(
			func(app cdk.Application, display cdk.Display, ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) enums.EventFlag {
				log.DebugF("Startup hit")
				display.CaptureCtrlC()
				w := cdk.NewWindow("Demo Plugin", display)
				w.Connect(cdk.SignalDraw, "demo-plugin-draw-handler", func(data []interface{}, argv ...interface{}) enums.EventFlag {
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
					return enums.EVENT_STOP
				})
				w.Connect(cdk.SignalEvent, "demo-plugin-event-handler", func(data []interface{}, argv ...interface{}) enums.EventFlag {
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
							if surface, err := memphis.GetSurface(w.ObjectID()); err == nil {
								surface.Resize(ptypes.MakeRectangle(width, height))
							}
						default:
							w.LogInfo("ProcessEvent: %T %v", v, v)
						}
						return enums.EVENT_STOP
					}
					return enums.EVENT_PASS
				})
				w.Init()
				display.FocusWindow(w)
				// draw the screen every second so the time displayed is now
				cdk.AddTimeout(time.Second, func() enums.EventFlag {
					display.RequestDraw()   // redraw the window, is buffered
					display.RequestShow()   // flag buffer for immediate show
					return enums.EVENT_PASS // keep looping every second
				})
				app.NotifyStartupComplete()
				return enums.EVENT_PASS
			},
		),
	)
	CdkApp.Connect(cdk.SignalShutdown, "helloworld-shutdown", func(_ []interface{}, _ ...interface{}) enums.EventFlag {
		fmt.Printf("Quitting helloworld normally.\n")
		return enums.EVENT_PASS
	})
}
