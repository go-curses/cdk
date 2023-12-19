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

package ptypes

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRectangle(t *testing.T) {
	Convey("Basic Rectangle Features", t, func() {
		r := NewRectangle(2, 2)
		So(r, ShouldNotBeNil)
		So(r, ShouldHaveSameTypeAs, &Rectangle{})
		So(r.String(), ShouldEqual, "{w:2,h:2}")
		r.Set(1, 1)
		So(r.W, ShouldEqual, 1)
		So(r.H, ShouldEqual, 1)
		So(r.Volume(), ShouldEqual, 1)
		r.SetRectangle(Rectangle{2, 2})
		So(r.W, ShouldEqual, 2)
		So(r.H, ShouldEqual, 2)
		So(r.Volume(), ShouldEqual, 4)
		r.Add(1, 1)
		So(r.W, ShouldEqual, 3)
		So(r.H, ShouldEqual, 3)
		So(r.Volume(), ShouldEqual, 9)
		r.AddRectangle(Rectangle{1, 1})
		So(r.W, ShouldEqual, 4)
		So(r.H, ShouldEqual, 4)
		So(r.Volume(), ShouldEqual, 16)
		r.Sub(1, 1)
		So(r.W, ShouldEqual, 3)
		So(r.H, ShouldEqual, 3)
		So(r.Volume(), ShouldEqual, 9)
		r.SubRectangle(Rectangle{1, 1})
		So(r.W, ShouldEqual, 2)
		So(r.H, ShouldEqual, 2)
		So(r.Volume(), ShouldEqual, 4)
		r.Set(10, 10)
		So(r.Volume(), ShouldEqual, 100)
		region := Region{Point2I{5, 5}, Rectangle{5, 5}}
		So(r.ClampToRegion(region), ShouldEqual, false)
		So(r.Volume(), ShouldEqual, 100)
		region.Set(0, 0, 2, 2)
		So(r.ClampToRegion(region), ShouldEqual, true)
		So(r.Volume(), ShouldEqual, 4)
		region.Set(5, 5, 2, 2)
		So(r.ClampToRegion(region), ShouldEqual, true)
		So(r.Volume(), ShouldEqual, 25)
	})
}
