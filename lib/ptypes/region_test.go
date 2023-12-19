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

func TestRegion(t *testing.T) {
	Convey("Basic Region Features", t, func() {
		r := NewRegion(1, 1, 2, 2)
		So(r, ShouldNotBeNil)
		So(r, ShouldHaveSameTypeAs, &Region{})
		So(r.String(), ShouldEqual, "{x:1,y:1,w:2,h:2}")
		So(r.Origin(), ShouldHaveSameTypeAs, Point2I{})
		So(r.Origin().X, ShouldEqual, 1)
		So(r.Origin().Y, ShouldEqual, 1)
		So(r.Size(), ShouldHaveSameTypeAs, Rectangle{})
		So(r.Size().W, ShouldEqual, 2)
		So(r.Size().H, ShouldEqual, 2)
		So(r.FarPoint().X, ShouldEqual, 3)
		So(r.FarPoint().Y, ShouldEqual, 3)
		So(r.HasPoint(Point2I{0, 0}), ShouldEqual, false)
		So(r.HasPoint(Point2I{1, 1}), ShouldEqual, true)
		r.Set(2, 2, 4, 4)
		So(r.String(), ShouldEqual, "{x:2,y:2,w:4,h:4}")
	})
}
