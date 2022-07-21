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
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEventMouseBasics(t *testing.T) {
	Convey("EventMouse basics", t, func() {
		// new mouse event
		em := NewEventMouse(1, 1, ButtonNone, ModNone)
		So(em, ShouldHaveSameTypeAs, &EventMouse{})
		em = NewEventMouse(1, 1, ButtonNone, ModNone)
		So(em.Report(), ShouldEqual, "None [Move]")
		em = NewEventMouse(1, 1, WheelDown, ModNone)
		So(em.Report(), ShouldEqual, "WheelDown [Impulse]")
		em = NewEventMouse(1, 1, WheelLeft, ModNone)
		So(em.Report(), ShouldEqual, "WheelLeft [Impulse]")
		em = NewEventMouse(1, 1, WheelRight, ModNone)
		So(em.Report(), ShouldEqual, "WheelRight [Impulse]")
		em = NewEventMouse(1, 1, WheelUp, ModNone)
		So(em.Report(), ShouldEqual, "WheelUp [Impulse]")
	})
}

func TestEventMouseButton(t *testing.T) {
	Convey("EventMouse Button Press/Release", t, func() {
		now := time.Now()
		em := NewEventMouse(1, 1, Button1, ModMeta)
		So(em.When().UnixNano(), ShouldBeGreaterThan, now.UnixNano())
		So(em.Modifiers(), ShouldEqual, ModMeta)
		So(em.Button(), ShouldEqual, Button1)
		So(em.ButtonHas(Button1), ShouldEqual, true)
		So(em.State(), ShouldEqual, BUTTON_PRESS)
		So(em.StateHas(BUTTON_PRESS), ShouldEqual, true)
		So(em.ButtonPressed(), ShouldEqual, Button1)
		So(em.IsPressed(), ShouldEqual, true)
		So(em.IsReleased(), ShouldEqual, false)
		So(em.IsWheelImpulse(), ShouldEqual, false)
		So(em.WheelImpulse(), ShouldEqual, ButtonNone)
		So(DescribeButton(Button1), ShouldEqual, "Button1")

		em = NewEventMouse(1, 1, ButtonNone, ModNone)
		So(em.IsPressed(), ShouldEqual, false)
		So(em.IsReleased(), ShouldEqual, true)
		So(em.ButtonPressed(), ShouldEqual, ButtonNone)
	})
}

func TestEventMouseWheel(t *testing.T) {
	Convey("EventMouse Wheel Impulse", t, func() {
		em := NewEventMouse(1, 1, WheelDown, ModNone)
		So(em.WheelImpulse(), ShouldEqual, WheelDown)
		So(em.ButtonPressed(), ShouldEqual, ButtonNone)
		So(em.IsWheelImpulse(), ShouldEqual, true)
	})
}

func TestEventMouseDrag(t *testing.T) {
	Convey("EventMouse Drag events", t, func() {
		// new mouse event
		_ = NewEventMouse(1, 1, ButtonNone, ModNone)
		// simple mouse movement
		em := NewEventMouse(2, 2, ButtonNone, ModNone)
		So(em.IsMoving(), ShouldEqual, true)
		So(em.IsDragging(), ShouldEqual, false)
		// simple confirmations
		So(em.WheelImpulse(), ShouldEqual, ButtonNone)
		So(em.ButtonPressed(), ShouldEqual, ButtonNone)
		So(em.IsDragging(), ShouldEqual, false)
		// click-drag started
		_ = NewEventMouse(2, 2, Button1, ModNone)
		em = NewEventMouse(3, 3, Button1, ModNone)
		So(em.IsDragStarted(), ShouldEqual, true)
		So(em.IsDragging(), ShouldEqual, true)
		// click-drag dragging
		em = NewEventMouse(4, 4, Button1, ModNone)
		So(em.IsDragging(), ShouldEqual, true)
		So(em.IsReleased(), ShouldEqual, false)
		// click-drag release
		em = NewEventMouse(4, 4, ButtonNone, ModNone)
		So(em.IsDragging(), ShouldEqual, false)
		So(em.IsDragStopped(), ShouldEqual, true)
		// click-drag-move checks
		_ = NewEventMouse(4, 4, Button1, ModNone)
		em = NewEventMouse(5, 5, Button1, ModNone)
		So(em.IsDragStarted(), ShouldEqual, true)
		em = NewEventMouse(5, 5, ButtonNone, ModNone)
		So(em.IsDragStopped(), ShouldEqual, true)
		// click-drag-move stopped by other button
		_ = NewEventMouse(5, 5, Button1, ModNone)
		_ = NewEventMouse(6, 6, Button1, ModNone)
		em = NewEventMouse(6, 6, Button2, ModNone)
		So(em.IsDragStopped(), ShouldEqual, true)
		So(em.IsPressed(), ShouldEqual, false)
		em = NewEventMouse(6, 6, ButtonNone, ModNone)
		So(em.IsPressed(), ShouldEqual, false)
		// click-drag-ghost-move stopped by other button
		_ = NewEventMouse(6, 6, Button1, ModNone)
		em = NewEventMouse(7, 7, Button1, ModNone)
		So(em.IsDragging(), ShouldEqual, true)
		em = NewEventMouse(8, 8, Button2, ModNone)
		So(em.IsDragStopped(), ShouldEqual, true)
		So(em.IsPressed(), ShouldEqual, false)
		em = NewEventMouse(8, 8, ButtonNone, ModNone)
		So(em.IsDragStopped(), ShouldEqual, false)
		So(em.IsPressed(), ShouldEqual, false)
		// click-drag-move-ghost-move stopped by other button
		_ = NewEventMouse(8, 8, Button1, ModNone)
		_ = NewEventMouse(9, 9, Button1, ModNone)
		em = NewEventMouse(9, 9, Button1, ModNone)
		So(em.IsDragging(), ShouldEqual, true)
		So(em.x, ShouldEqual, 9)
		So(em.y, ShouldEqual, 9)
		_ = NewEventMouse(9, 9, Button1, ModNone)
		em = NewEventMouse(9, 9, Button2, ModNone)
		So(em.IsDragging(), ShouldEqual, false)
		So(em.IsPressed(), ShouldEqual, false)
		_ = NewEventMouse(9, 9, ButtonNone, ModNone)
		// click-drag-move-move-ghost-move
		em = NewEventMouse(9, 9, Button1, ModNone) // click
		So(em.IsPressed(), ShouldEqual, true)
		em = NewEventMouse(8, 8, Button1, ModNone) // drag-start
		So(em.IsDragStarted(), ShouldEqual, true)
		em = NewEventMouse(7, 7, Button1, ModNone) // drag-move
		So(em.IsDragging(), ShouldEqual, true)
		em = NewEventMouse(6, 6, Button1, ModNone) // drag-move-move
		So(em.IsDragging(), ShouldEqual, true)
		em = NewEventMouse(5, 5, Button2, ModNone) // drag-move-move-derp
		So(em.IsDragging(), ShouldEqual, false)
		_ = NewEventMouse(5, 5, ButtonNone, ModNone) // drag-move-move-derp
		// click-ghost-release
		em = NewEventMouse(5, 5, Button1, ModNone) // drag-move-move-derp
		So(em.IsPressed(), ShouldEqual, true)
		em = NewEventMouse(6, 6, Button2, ModNone) // drag-move-move-derp
		So(em.IsPressed(), ShouldEqual, true)
		em = NewEventMouse(6, 6, ButtonNone, ModNone) // drag-move-move-derp
		So(em.IsPressed(), ShouldEqual, false)
	})
}

func eventLoop(s OffScreen, evch chan Event) {
	for {
		ev := s.PollEvent()
		if ev == nil {
			close(evch)
			return
		}
		select {
		case evch <- ev:
		case <-time.After(time.Second):
		}
	}
}

func TestMouseEvents(t *testing.T) {

	s := NewTestingScreen(t, "")
	defer s.Close()

	s.EnableMouse()
	s.InjectMouse(4, 9, Button1, ModCtrl)
	evch := make(chan Event)
	em := &EventMouse{}
	done := false
	go eventLoop(s, evch)

	for !done {
		select {
		case ev := <-evch:
			if evm, ok := ev.(*EventMouse); ok {
				em = evm
				done = true
			}
			continue
		case <-time.After(time.Second):
			done = true
		}
	}

	if x, y := em.Position(); x != 4 || y != 9 {
		t.Errorf("Mouse position wrong (%v, %v)", x, y)
	}
	if em.Buttons() != Button1 {
		t.Errorf("Should be Button1")
	}
	if em.Modifiers() != ModCtrl {
		t.Errorf("Modifiers should be control")
	}
}
