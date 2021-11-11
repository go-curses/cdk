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

	. "github.com/smartystreets/goconvey/convey"
)

func TestMouseState(t *testing.T) {
	Convey("MouseState checks", t, func() {
		ms := MOUSE_NONE
		So(ms, ShouldHaveSameTypeAs, MouseState(0))
		So(ms.String(), ShouldEqual, "MOUSE_NONE")
		ms = ms.Set(MOUSE_MOVE)
		So(ms.Has(MOUSE_MOVE), ShouldEqual, true)
		So(ms.String(), ShouldEqual, "MOUSE_MOVE")
		ms = ms.Clear(MOUSE_MOVE)
		So(ms.Has(MOUSE_MOVE), ShouldEqual, false)
		ms = ms.Toggle(MOUSE_MOVE)
		So(ms.Has(MOUSE_MOVE), ShouldEqual, true)
		So(BUTTON_PRESS.String(), ShouldEqual, "BUTTON_PRESS")
		So(BUTTON_RELEASE.String(), ShouldEqual, "BUTTON_RELEASE")
		So(WHEEL_PULSE.String(), ShouldEqual, "WHEEL_PULSE")
		So(DRAG_START.String(), ShouldEqual, "DRAG_START")
		So(DRAG_MOVE.String(), ShouldEqual, "DRAG_MOVE")
		So(DRAG_STOP.String(), ShouldEqual, "DRAG_STOP")
		So((DRAG_STOP + 1).String(), ShouldEqual, "MouseState(129)")
	})
}
