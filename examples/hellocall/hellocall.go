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
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-curses/cdk"
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/ptypes"
	cstrings "github.com/go-curses/cdk/lib/strings"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
	"github.com/go-curses/cdk/memphis"
)

// Build Configuration Flags
// setting these will enable command line flags and their corresponding features
// use `go build -v -ldflags="-X 'main.IncludeLogFullPaths=false'"`
var (
	IncludeProfiling          = "false"
	IncludeLogFile            = "false"
	IncludeLogFormat          = "false"
	IncludeLogFullPaths       = "false"
	IncludeLogLevel           = "false"
	IncludeLogLevels          = "false"
	IncludeLogTimestamps      = "false"
	IncludeLogTimestampFormat = "false"
	IncludeLogOutput          = "false"
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
	cdk.Build.Profiling = cstrings.IsTrue(IncludeProfiling)
	cdk.Build.LogFile = cstrings.IsTrue(IncludeLogFile)
	cdk.Build.LogFormat = cstrings.IsTrue(IncludeLogFormat)
	cdk.Build.LogFullPaths = cstrings.IsTrue(IncludeLogFullPaths)
	cdk.Build.LogLevel = cstrings.IsTrue(IncludeLogLevel)
	cdk.Build.LogLevels = cstrings.IsTrue(IncludeLogLevels)
	cdk.Build.LogTimestamps = cstrings.IsTrue(IncludeLogTimestamps)
	cdk.Build.LogTimestampFormat = cstrings.IsTrue(IncludeLogTimestampFormat)
	cdk.Build.LogOutput = cstrings.IsTrue(IncludeLogOutput)
	CdkApp = cdk.NewApplication(
		"hellocall",
		"An example CDK application/plugin using Display.Call",
		"Hello Call is an example CDK application/plugin using Display.Call",
		"0.0.1",
		"hellocall",
		"Hello Call",
		"/dev/tty",
	)
	CdkApp.Connect(
		cdk.SignalStartup,
		"hellocall-startup-handler",
		cdk.WithArgvApplicationSignalStartup(
			func(app cdk.Application, display cdk.Display, ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) enums.EventFlag {
				log.DebugF("Startup hit")
				display.CaptureCtrlC()
				w := cdk.NewWindow("Demo Plugin", display)
				w.Connect(cdk.SignalDraw, "hellocall-draw-handler", func(data []interface{}, argv ...interface{}) enums.EventFlag {
					if surface, ok := argv[1].(memphis.Surface); ok {
						theme := w.GetTheme()
						size := surface.GetSize()
						surface.Box(ptypes.Point2I{}, size, true, true, false, ' ', theme.Content.Normal, theme.Border.Normal, theme.Border.BorderRunes)
						content := "<b><span foreground=\"darkgreen\" background=\"yellow\"><u>H</u>ello</span> <span foreground=\"brown\" background=\"orange\"><i>W</i>orld</span></b>\n"
						content += "<span foreground=\"cyan\" background=\"gray\">(press CTRL+c or ESC to exit)</span>\n"
						content += "<span foreground=\"cyan\" background=\"gray\">(press b to bash shell)</span>\n"
						content += "<span foreground=\"cyan\" background=\"gray\">(press f to run a func)</span>\n"
						content += "<span foreground=\"yellow\" background=\"darkblue\">" + time.Now().Format("2006-01-02 15:04:05") + "</span>"
						textPoint := ptypes.MakePoint2I(size.W/2/2, size.H/2-1)
						textSize := ptypes.MakeRectangle(size.W/2, size.H/2)
						surface.DrawText(textPoint, textSize, enums.JUSTIFY_CENTER, false, enums.WRAP_WORD, false, theme.Content.Normal, true, true, content)
						return enums.EVENT_STOP
					}
					return enums.EVENT_STOP
				})
				w.Connect(cdk.SignalEvent, "hellocall-event-handler", func(data []interface{}, argv ...interface{}) enums.EventFlag {
					if evt, ok := argv[1].(cdk.Event); ok {
						switch v := evt.(type) {
						case *cdk.EventError:
							w.LogInfo("ProcessEvent: Error (error:%v)", v.Err())
						case *cdk.EventKey:
							if v.Key() == cdk.KeyESC {
								w.LogInfo("ProcessEvent: RequestQuit (key:%v)", v.Name())
								cdk.GetDefaultDisplay().RequestQuit()
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
