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

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-curses/cdk/lib/paint"
)

func TestCellBasics(t *testing.T) {
	Convey("Cell basics", t, func() {
		c := newCell()
		So(c, ShouldNotBeNil)
		So(c.init(), ShouldEqual, false)
	})
}

func TestCellBufferBasics(t *testing.T) {
	Convey("CellBuffer basics", t, func() {
		cb := NewCellBuffer()
		So(cb, ShouldNotBeNil)
		So(cb.init(), ShouldEqual, false)
		cb.Resize(1, 1)
		cb.SetCell(0, 0, '0', []rune{'0'}, paint.StyleDefault)
		So(cb.Dirty(0, 0), ShouldEqual, true)
		cb.SetDirty(0, 0, false)
		So(cb.Dirty(0, 0), ShouldEqual, false)
		cb.Invalidate()
		So(cb.Dirty(0, 0), ShouldEqual, true)
		cb.SetDirty(0, 0, false)
		So(cb.Dirty(0, 0), ShouldEqual, false)
		cb.SetDirty(0, 0, true)
		So(cb.Dirty(0, 0), ShouldEqual, true)
		cb.SetDirty(0, 0, false)
		cb.Resize(1, 1)
		So(cb.Dirty(0, 0), ShouldEqual, false)
		cb.SetCell(0, 0, '0', []rune{'0'}, paint.StyleDefault)
		cb.SetDirty(0, 0, false)
		prev_main := cb.cells[0].currMain
		cb.cells[0].currMain = '1'
		So(cb.Dirty(0, 0), ShouldEqual, true)
		cb.cells[0].currMain = prev_main
		So(cb.Dirty(0, 0), ShouldEqual, false)
		prev_style := cb.cells[0].currStyle
		cb.cells[0].currStyle = paint.StyleDefault.Reverse(true)
		So(cb.Dirty(0, 0), ShouldEqual, true)
		cb.cells[0].currStyle = prev_style
		So(cb.Dirty(0, 0), ShouldEqual, false)
		prev_comb := cb.cells[0].currComb
		cb.cells[0].currComb = []rune{'0', '1'}
		So(cb.Dirty(0, 0), ShouldEqual, true)
		cb.cells[0].currComb = prev_comb
		So(cb.Dirty(0, 0), ShouldEqual, false)
		prev_comb = cb.cells[0].currComb
		cb.cells[0].currComb = []rune{'1'}
		So(cb.Dirty(0, 0), ShouldEqual, true)
		cb.cells[0].currComb = prev_comb
		So(cb.Dirty(0, 0), ShouldEqual, false)
		cb.SetCell(0, 0, '0', nil, paint.StyleDefault)
		So(len(cb.cells[0].currComb), ShouldEqual, 0)
	})
}
