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

package ptypes

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPoint2I(t *testing.T) {
	Convey("Basic Point2I Features", t, func() {
		p := NewPoint2I(2, 2)
		So(p, ShouldNotBeNil)
		So(p, ShouldHaveSameTypeAs, &Point2I{})
		So(p.String(), ShouldEqual, "{x:2,y:2}")
		p.Set(1, 1)
		So(p.X, ShouldEqual, 1)
		So(p.Y, ShouldEqual, 1)
		p.SetPoint(Point2I{2, 2})
		So(p.X, ShouldEqual, 2)
		So(p.Y, ShouldEqual, 2)
		p.Add(1, 1)
		So(p.X, ShouldEqual, 3)
		So(p.Y, ShouldEqual, 3)
		p.AddPoint(Point2I{1, 1})
		So(p.X, ShouldEqual, 4)
		So(p.Y, ShouldEqual, 4)
		p.Sub(1, 1)
		So(p.X, ShouldEqual, 3)
		So(p.Y, ShouldEqual, 3)
		p.SubPoint(Point2I{1, 1})
		So(p.X, ShouldEqual, 2)
		So(p.Y, ShouldEqual, 2)
		p.Set(10, 10)
		region := Region{Point2I{5, 5}, Rectangle{5, 5}}
		So(p.ClampToRegion(region), ShouldEqual, false)
		So(p.String(), ShouldEqual, "{x:10,y:10}")
		region.Set(0, 0, 2, 2)
		So(p.ClampToRegion(region), ShouldEqual, true)
		So(p.String(), ShouldEqual, "{x:2,y:2}")
		region.Set(5, 5, 2, 2)
		So(p.ClampToRegion(region), ShouldEqual, true)
		So(p.String(), ShouldEqual, "{x:5,y:5}")
	})
}
