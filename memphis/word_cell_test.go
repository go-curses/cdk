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

package memphis

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-curses/cdk/lib/paint"
)

func TestWordCell(t *testing.T) {
	Convey("Word Cells with...", t, func() {
		Convey("Empty Strings", func() {
			wc := NewWordCell("", paint.GetDefaultMonoStyle())
			So(wc, ShouldNotBeNil)
			So(wc.String(), ShouldEqual, "")
			So(wc.Characters(), ShouldHaveLength, 0)
			So(wc.Len(), ShouldEqual, 0)
		})
		Convey("One Word", func() {
			wc := NewWordCell("word", paint.GetDefaultMonoStyle())
			So(wc, ShouldNotBeNil)
			So(wc.Value(), ShouldEqual, "word")
			So(wc.Characters(), ShouldHaveLength, 4)
			So(wc.Len(), ShouldEqual, 4)
		})
		Convey("More Than One Word", func() {
			wc := NewWordCell("another word", paint.GetDefaultMonoStyle())
			So(wc, ShouldNotBeNil)
			So(wc.CompactLen(), ShouldEqual, 12)
		})
		Convey("Basic checks", func() {
			wc := NewEmptyWordCell()
			So(wc, ShouldNotBeNil)
			So(wc.Len(), ShouldEqual, 0)
			So(wc.IsSpace(), ShouldEqual, true)
			So(wc.HasSpace(), ShouldEqual, false)
			So(wc.GetCharacter(0), ShouldBeNil)
			wc.AppendRune(' ', paint.GetDefaultMonoStyle())
			So(wc.Len(), ShouldEqual, 1)
			c := wc.GetCharacter(0)
			So(c, ShouldNotBeNil)
			So(c.Value(), ShouldEqual, ' ')
			So(c.Style().String(), ShouldEqual, paint.GetDefaultMonoStyle().String())
			So(c.Width(), ShouldEqual, 1)
			So(wc.HasSpace(), ShouldEqual, true)
			So(wc.CompactLen(), ShouldEqual, 1)
			wc.AppendRune(' ', paint.GetDefaultMonoStyle())
			So(wc.CompactLen(), ShouldEqual, 1)
			So(wc.Len(), ShouldEqual, 2)
		})
	})
}