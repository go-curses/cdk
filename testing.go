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
	"testing"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/log"
)

type AppFn func(app Application)
type DisplayManagerFn func(d Display)

func WithApp(initFn SignalListenerFn, action AppFn) func() {
	return func() {

		app := NewApplication(
			"AppName", "AppUsage",
			"AppDesc", "v0.0.0",
			"app-tag", "AppTitle",
			OffscreenTtyPath,
		)
		app.Connect(SignalStartup, "testing-withapp-init-fn-handler", initFn)
		defer func() {
			if app != nil {
				app.Destroy()
			}
			app = nil
		}()
		app.SetupDisplay()
		if f := app.Emit(SignalStartup); f == enums.EVENT_STOP {
			log.ErrorF("withapp startup listeners requested EVENT_STOP")
		} else {
			action(app)
		}
	}
}

func WithDisplayManager(action DisplayManagerFn) func() {
	return func() {
		d := NewDisplay("testing", OffscreenTtyPath)
		_ = d.CaptureDisplay()
		defer d.ReleaseDisplay()
		action(d)
	}
}

func TestingMakesNoContent(_ []interface{}, _ ...interface{}) enums.EventFlag {
	return enums.EVENT_PASS
}

func TestingMakesActiveWindow(d Display) error {
	w := NewOffscreenWindow(d.GetTitle())
	d.FocusWindow(w)
	return nil
}

func NewTestingScreen(t *testing.T, charset string) OffScreen {
	s, err := MakeOffScreen(charset)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
