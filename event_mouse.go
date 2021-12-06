// Copyright 2021  The CDK Authors
// Copyright 2015 The TCell Authors
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
	"fmt"
	"time"

	"github.com/go-curses/cdk/lib/ptypes"
)

// EventMouse is a mouse event.  It is sent on either mouse up or mouse down
// events.  It is also sent on mouse motion events - if the terminal supports
// it.  We make every effort to ensure that mouse release events are delivered.
// Hence, click drag can be identified by a motion event with the mouse down,
// without any intervening button release.  On some terminals only the initiating
// press and terminating release event will be delivered.
//
// Mouse wheel events, when reported, may appear on their own as individual
// impulses; that is, there will normally not be a release event delivered
// for mouse wheel movements.
//
// Most terminals cannot report the state of more than one button at a time --
// and some cannot report motion events unless a button is pressed.
//
// Applications can inspect the time between events to resolve double or
// triple clicks.
type EventMouse struct {
	t   time.Time
	btn ButtonMask
	mod ModMask
	x   int
	y   int
	s   MouseState
	b   ButtonMask
}

var (
	MOUSE_STATES map[MouseState]string = map[MouseState]string{
		MOUSE_NONE:     "None",
		MOUSE_MOVE:     "Move",
		BUTTON_PRESS:   "Pressed",
		BUTTON_RELEASE: "Released",
		WHEEL_PULSE:    "Impulse",
		DRAG_START:     "DragStart",
		DRAG_MOVE:      "DragMove",
		DRAG_STOP:      "DragStop",
	}
	previous_event_mouse *EventMouse = &EventMouse{
		t:   time.Now(),
		x:   0,
		y:   0,
		btn: ButtonNone,
		mod: ModNone,
		s:   MOUSE_NONE,
		b:   ButtonNone,
	}
)

// NewEventMouse is used to create a new mouse event.  Applications
// shouldn't need to use this; its mostly for display implementors.
func NewEventMouse(x, y int, btn ButtonMask, mod ModMask) *EventMouse {
	em := &EventMouse{
		t:   time.Now(),
		x:   x,
		y:   y,
		btn: btn,
		mod: mod,
		s:   MOUSE_NONE,
		b:   ButtonNone,
	}
	em.process_mouse_event()
	return em
}

func (ev *EventMouse) Clone() Event {
	return &EventMouse{
		t:   ev.t,
		x:   ev.x,
		y:   ev.y,
		btn: ev.btn,
		mod: ev.mod,
		s:   ev.s,
		b:   ev.b,
	}
}

func (ev *EventMouse) CloneForPosition(x, y int) Event {
	return &EventMouse{
		t:   ev.t,
		x:   x,
		y:   y,
		btn: ev.btn,
		mod: ev.mod,
		s:   ev.s,
		b:   ev.b,
	}
}

// When returns the time when this EventMouse was created.
func (ev *EventMouse) When() time.Time {
	return ev.t
}

// Buttons returns the list of buttons that were pressed or wheel motions.
func (ev *EventMouse) Buttons() ButtonMask {
	return ev.btn
}

// Modifiers returns a list of keyboard modifiers that were pressed
// with the mouse button(s).
func (ev *EventMouse) Modifiers() ModMask {
	return ev.mod
}

// Position returns the mouse position in character cells.  The origin
// 0, 0 is at the upper left corner.
func (ev *EventMouse) Position() (x, y int) {
	return ev.x, ev.y
}

func (ev *EventMouse) Point2I() (point ptypes.Point2I) {
	point = ptypes.MakePoint2I(ev.x, ev.y)
	return
}

func (ev *EventMouse) Button() ButtonMask {
	return ev.b
}

func (ev *EventMouse) ButtonHas(check ButtonMask) bool {
	return ev.btn.Has(check)
}

func (ev *EventMouse) State() MouseState {
	return ev.s
}

func (ev *EventMouse) StateHas(check MouseState) bool {
	return ev.s.Has(check)
}

func (ev *EventMouse) IsPressed() bool {
	return ev.s.Has(BUTTON_PRESS)
}

func (ev *EventMouse) IsReleased() bool {
	return ev.s.Has(BUTTON_RELEASE)
}

func (ev *EventMouse) IsMoving() bool {
	return ev.s.Has(MOUSE_MOVE)
}

func (ev *EventMouse) IsDragging() bool {
	return ev.s.Has(DRAG_MOVE) || ev.s.Has(DRAG_START)
}

func (ev *EventMouse) IsDragStarted() bool {
	return ev.s.Has(DRAG_START)
}

func (ev *EventMouse) IsDragStopped() bool {
	return ev.s.Has(DRAG_STOP)
}

func (ev *EventMouse) IsWheelImpulse() bool {
	return ev.s.Has(WHEEL_PULSE)
}

func (ev *EventMouse) WheelImpulse() ButtonMask {
	b := ButtonNone
	for i := uint(8); i < 12; i++ {
		if int(ev.btn)&(1<<i) != 0 {
			b |= 1 << i
			break
		}
	}
	return b
}

func (ev *EventMouse) ButtonPressed() ButtonMask {
	b := ButtonNone
	for i := uint(0); i < 8; i++ {
		if int(ev.btn)&(1<<i) != 0 {
			b |= 1 << i
		}
	}
	return b
}

func DescribeButton(button ButtonMask) string {
	desc := ""
	for i := uint(0); i < 8; i++ {
		if int(button)&(1<<i) != 0 {
			desc += fmt.Sprintf("Button%d ", i+1)
		}
	}
	if button&WheelUp != 0 {
		desc += "WheelUp "
	}
	if button&WheelDown != 0 {
		desc += "WheelDown "
	}
	if button&WheelLeft != 0 {
		desc += "WheelLeft "
	}
	if button&WheelRight != 0 {
		desc += "WheelRight "
	}
	if len(desc) > 0 {
		// trim trailing space
		desc = desc[:len(desc)-1]
	} else {
		desc = "None"
	}
	return desc
}

func (ev *EventMouse) Report() string {
	return fmt.Sprintf(
		"%v [%v]",
		DescribeButton(ev.b),
		MOUSE_STATES[ev.s],
	)
}

/*

	MOUSE_NONE: non-event
	MOUSE_MOVE: mouse moved, no buttons
	BUTTON_PRESS: button pressed
	BUTTON_RELEASE: button released
	WHEEL_PULSE: wheel impulse (no release)
	DRAG_START: prev is press, and mouse is moving
	DRAG_MOVE: prev is drag start and mouse still moving
	DRAG_STOP: button released and mouse may have moved

*/
func (ev *EventMouse) process_mouse_event() {
	pem := previous_event_mouse
	previous_event_mouse = ev
	ev.s = MOUSE_NONE
	ev.b = ButtonNone

	did_move := false
	if pem.x != ev.x || pem.y != ev.y {
		did_move = true
	}

	impulse := ButtonNone
	pressed := ev.ButtonPressed()
	if pressed != ButtonNone {
		// press event
		ev.b = pressed
		if did_move {
			if pem.s.Has(DRAG_START) {
				if pem.b == pressed {
					ev.s = DRAG_MOVE
				} else {
					ev.s = DRAG_STOP
					ev.b = pem.b
				}
			} else if pem.s.Has(DRAG_MOVE) {
				if pem.b == pressed {
					ev.s = DRAG_MOVE
				} else {
					ev.s = DRAG_STOP
					ev.b = pem.b
				}
			} else if pem.s.Has(BUTTON_PRESS) {
				if pem.b == pressed {
					ev.s = DRAG_START
				} else {
					ev.s = BUTTON_PRESS
					ev.b = pem.b
				}
			} else {
				ev.s = BUTTON_PRESS
			}
		} else { // didn't move
			if pem.s.Has(DRAG_START) {
				if pem.b == pressed {
					ev.s = DRAG_MOVE
				} else {
					ev.s = DRAG_STOP
					ev.b = pem.b
				}
			} else if pem.s.Has(DRAG_MOVE) {
				if pem.b == pressed {
					ev.s = DRAG_MOVE
				} else {
					ev.s = DRAG_STOP
					ev.b = pem.b
				}
			} else {
				ev.s = BUTTON_PRESS
			}
		}
	} else {
		if pem.s.Has(DRAG_START) {
			ev.s = DRAG_STOP
			ev.b = pem.b
		} else if pem.s.Has(DRAG_MOVE) {
			ev.s = DRAG_STOP
			ev.b = pem.b
		} else if pem.s.Has(BUTTON_PRESS) {
			ev.s = BUTTON_RELEASE
			ev.b = pem.b
		} else {
			impulse = ev.WheelImpulse()
			if impulse != ButtonNone {
				// scroll event
				ev.s = WHEEL_PULSE
				ev.b = impulse
			} else {
				ev.s = MOUSE_MOVE
				ev.b = ButtonNone
			}
		}
	}
}
