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
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKey(t *testing.T) {
	Convey("EventKey checks", t, func() {
		// <ctrl> + <space>
		then := time.Now()
		ek := NewEventKey(KeyRune, ' ', ModCtrl)
		So(ek, ShouldHaveSameTypeAs, &EventKey{})
		now := time.Now()
		So(ek.When().UnixNano(), ShouldBeGreaterThanOrEqualTo, then.UnixNano())
		So(ek.When().UnixNano(), ShouldBeLessThanOrEqualTo, now.UnixNano())
		So(ek.Key(), ShouldEqual, KeyRune)
		So(ek.Rune(), ShouldEqual, ' ')
		So(ek.Modifiers(), ShouldEqual, ModCtrl)
		So(ek.Modifiers().Has(ModCtrl), ShouldEqual, true)
		// escape
		ek = NewEventKey(KeyRune, 0x1b, ModNone)
		So(ek.Key(), ShouldEqual, KeyEscape)
		So(ek.Rune(), ShouldEqual, KeyEscape)
		// escape
		ek = NewEventKey(KeyRune, 0x7f, ModNone)
		So(ek.Key(), ShouldEqual, KeyDEL)
		So(ek.Rune(), ShouldEqual, KeyDEL)
		So(ek.Modifiers(), ShouldEqual, ModNone)
		// escape
		ek = NewEventKey(KeyRune, rune(KeyBEL), ModNone)
		So(ek.Key(), ShouldEqual, KeySmallG)
		So(ek.Rune(), ShouldEqual, KeyBEL)
		So(ek.Modifiers(), ShouldEqual, ModCtrl)
	})
	Convey("EventKey name checks", t, func() {
		ek := NewEventKey(KeyRune, ' ', ModCtrl)
		So(ek.Name(), ShouldEqual, "Ctrl+Rune[ ]")
		ek = NewEventKey(KeyRune, ' ', ModAlt)
		So(ek.Name(), ShouldEqual, "Alt+Rune[ ]")
		ek = NewEventKey(KeyRune, ' ', ModShift)
		So(ek.Name(), ShouldEqual, "Shift+Rune[ ]")
		ek = NewEventKey(KeyRune, ' ', ModMeta)
		So(ek.Name(), ShouldEqual, "Meta+Rune[ ]")
		ek = NewEventKey(KeyRune, ' ', ModNone)
		So(ek.Name(), ShouldEqual, "Rune[ ]")
		ek = NewEventKey(' ', ' ', ModNone)
		So(ek.Name(), ShouldEqual, "Key[32,32]")
		ek = NewEventKey(KeyCtrlSpace, rune(KeyCtrlSpace), ModNone)
		So(ek.Name(), ShouldEqual, "Ctrl-Space")
		ek = NewEventKey(KeyCtrlSpace, rune(KeyCtrlSpace), ModCtrl)
		So(ek.Name(), ShouldEqual, "Ctrl+Space")
	})
}
