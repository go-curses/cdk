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

package memphis

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-curses/cdk/lib/paint"
)

func TestTextCell(t *testing.T) {
	Convey("Basic checks", t, func() {
		tc := NewTextCellFromRune('*', paint.GetDefaultMonoStyle())
		So(tc, ShouldNotBeNil)
		So(tc.IsSpace(), ShouldEqual, false)
		So(tc.Width(), ShouldEqual, 1)
		So(tc.Style().String(), ShouldEqual, paint.GetDefaultMonoStyle().String())
		So(tc.Dirty(), ShouldEqual, true)
		So(tc.Value(), ShouldEqual, '*')
		So(tc.String(), ShouldEqual, "{Char=*,Style={white[#ffffff],black[#000000],0}}")
		tc.Set(' ')
		So(tc.IsSpace(), ShouldEqual, true)
		So(tc.Width(), ShouldEqual, 1)
		So(tc.Dirty(), ShouldEqual, true)
		tc.SetByte([]byte{'0'})
		So(tc.IsSpace(), ShouldEqual, false)
		So(tc.Dirty(), ShouldEqual, true)
		tc.SetStyle(paint.GetDefaultColorStyle())
		So(tc.Dirty(), ShouldEqual, true)
		So(tc.Style().String(), ShouldEqual, paint.GetDefaultColorStyle().String())
	})
}
