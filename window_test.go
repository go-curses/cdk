// Copyright (c) 2022-2023  The Go-Curses Authors
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
	. "github.com/smartystreets/goconvey/convey"
)

func TestWindow(t *testing.T) {
	Convey("Basic Window Features", t, func() {
		So(TypesManager.HasType(TypeWindow), ShouldEqual, true)
		w := &CWindow{}
		So(w.valid, ShouldEqual, false)
		So(w.GetTitle(), ShouldEqual, "")
		So(w.GetDisplay(), ShouldBeNil)
		So(w.Init(), ShouldEqual, false)
		So(w.Init(), ShouldEqual, true)
		d := &CDisplay{}
		w.SetDisplay(d)
		So(w.GetDisplay(), ShouldEqual, d)
		w.SetTitle("testing")
		So(w.GetTitle(), ShouldEqual, "testing")
		So(w.Draw(), ShouldEqual, enums.EVENT_PASS)
		So(w.ProcessEvent(&EventError{}), ShouldEqual, enums.EVENT_PASS)
	})
}
