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

func TestButtonMask(t *testing.T) {
	Convey("ButtonMask checks", t, func() {
		bm := ButtonNone
		So(bm.Has(Button1), ShouldEqual, false)
		bm = bm.Set(Button1)
		So(bm.Has(Button1), ShouldEqual, true)
		bm = bm.Clear(Button1)
		So(bm.Has(Button1), ShouldEqual, false)
		bm = bm.Toggle(Button1)
		So(bm.Has(Button1), ShouldEqual, true)
		So(bm.String(), ShouldEqual, "Button1")
		bm = ButtonMask(LastButtonMask + 1)
		So(bm.String(), ShouldEqual, "ButtonMask(4097)")
	})
}
