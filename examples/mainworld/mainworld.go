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
	"time"

	"github.com/go-curses/cdk"
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
)

func main() {
	app := cdk.NewApplication(
		"mainworld",
		"An example CDK application",
		"Main World is an example CDK application",
		"0.0.1",
		"mainworld",
		"Main World",
		"/dev/tty",
	)
	app.Connect(
		cdk.SignalStartup,
		"mainworld-startup",
		cdk.WithArgvApplicationSignalStartup(
			func(app cdk.Application, display cdk.Display, ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) enums.EventFlag {
				log.DebugF("Startup hit")
				display.CaptureCtrlC()
				w := &MainWindow{}
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
	app.Connect(cdk.SignalShutdown, "mainworld-quitter", func(_ []interface{}, _ ...interface{}) enums.EventFlag {
		fmt.Printf("Quitting mainworld normally.\n")
		return enums.EVENT_PASS
	})
	if app.MainInit(nil) {
		app.MainRun(func(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
			for app.MainEventsPending() {
				app.MainIterateEvents()
				select {
				case <-ctx.Done():
					return
				default:
					// nop, throttle CPU
					time.Sleep(cdk.MainIterateDelay)
				}
			}
			app.MainFinish()
		})
	}
}
