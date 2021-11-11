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

package paint

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAttr(t *testing.T) {
	Convey("AttrMask basics", t, func() {
		var am AttrMask
		am = AttrNone
		So(am.IsNormal(), ShouldEqual, true)
		// bold
		So(am.IsBold(), ShouldEqual, false)
		am = am.Bold(true)
		So(am.IsBold(), ShouldEqual, true)
		// blink
		So(am.IsBlink(), ShouldEqual, false)
		am = am.Blink(true)
		So(am.IsBlink(), ShouldEqual, true)
		// dim
		So(am.IsDim(), ShouldEqual, false)
		am = am.Dim(true)
		So(am.IsDim(), ShouldEqual, true)
		// reverse
		So(am.IsReverse(), ShouldEqual, false)
		am = am.Reverse(true)
		So(am.IsReverse(), ShouldEqual, true)
		// underline
		So(am.IsUnderline(), ShouldEqual, false)
		am = am.Underline(true)
		So(am.IsUnderline(), ShouldEqual, true)
	})
	Convey("AttrMask modifiers", t, func() {
		var am AttrMask
		am = AttrNone
		So(am.IsNormal(), ShouldEqual, true)
		am = am.
			Blink(true).
			Bold(true).
			Dim(true).
			Reverse(true).
			Underline(true)
		So(am.IsBlink(), ShouldEqual, true)
		So(am.IsBold(), ShouldEqual, true)
		So(am.IsDim(), ShouldEqual, true)
		So(am.IsReverse(), ShouldEqual, true)
		So(am.IsUnderline(), ShouldEqual, true)
		am = am.
			Blink(false).
			Bold(false).
			Dim(false).
			Reverse(false).
			Underline(false)
		So(am.IsBlink(), ShouldEqual, false)
		So(am.IsBold(), ShouldEqual, false)
		So(am.IsDim(), ShouldEqual, false)
		So(am.IsReverse(), ShouldEqual, false)
		So(am.IsUnderline(), ShouldEqual, false)
	})
	Convey("AttrMask converters", t, func() {
		var am AttrMask
		am = AttrBold
		So(am.IsBold(), ShouldEqual, true)
		am = am.Normal()
		So(am.IsNormal(), ShouldEqual, true)
	})
}
