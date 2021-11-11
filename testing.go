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
	"testing"

	"github.com/go-curses/cdk/log"
)

type AppFn func(app App)
type DisplayManagerFn func(d Display)

func WithApp(initFn DisplayInitFn, action AppFn) func() {
	return func() {

		app := NewApp(
			"AppName", "AppUsage",
			"AppDesc", "v0.0.0",
			"app-tag", "AppTitle",
			OffscreenTtyPath,
			initFn,
		)
		defer func() {
			if app != nil {
				app.Destroy()
			}
			app = nil
		}()
		app.SetupDisplay()
		if err := app.InitUI(); err != nil {
			log.Error(err)
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

func TestingMakesNoContent(d Display) error {
	return nil
}

func TestingMakesActiveWindow(d Display) error {
	w := NewOffscreenWindow(d.GetTitle())
	d.SetActiveWindow(w)
	return nil
}

func NewTestingScreen(t *testing.T, charset string) OffScreen {
	s, err := MakeOffScreen(charset)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
