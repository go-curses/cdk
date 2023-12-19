// Copyright 2022  The CDK Authors
// Copyright 2016 The TCell Authors
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
	"strings"
	"time"
)

// EventKey represents a key press.  Usually this is a key press followed
// by a key release, but since terminal programs don't have a way to report
// key release events, we usually get just one event.  If a key is held down
// then the terminal may synthesize repeated key presses at some predefined
// rate.  We have no control over that, nor visibility into it.
//
// In some cases, we can have a modifier key, such as ModAlt, that can be
// generated with a key press.  (This usually is represented by having the
// high bit set, or in some cases, by sending an ESC prior to the rune.)
//
// If the value of Key() is KeyRune, then the actual key value will be
// available with the Rune() method.  This will be the case for most keys.
// In most situations, the modifiers will not be set.  For example, if the
// rune is 'A', this will be reported without the ModShift bit set, since
// really can't tell if the Shift key was pressed (it might have been CAPSLOCK,
// or a terminal that only can send capitals, or keyboard with separate
// capital letters from lower case letters).
//
// Generally, terminal applications have far less visibility into keyboard
// activity than graphical applications.  Hence, they should avoid depending
// overly much on availability of modifiers, or the availability of any
// specific keys.
type EventKey struct {
	t   time.Time
	mod ModMask
	key Key
	ch  rune
}

// When returns the time when this Event was created, which should closely
// match the time when the key was pressed.
func (ev *EventKey) When() time.Time {
	return ev.t
}

// Rune returns the rune corresponding to the key press, if it makes sense.
// The result is only defined if the value of Key() is KeyRune.
func (ev *EventKey) Rune() rune {
	return ev.ch
}

func (ev *EventKey) RuneAsKey() Key {
	return Key(ev.ch)
}

// Key returns a virtual key code.  We use this to identify specific key
// codes, such as KeyEnter, etc.  Most control and function keys are reported
// with unique Key values.  Normal alphanumeric and punctuation keys will
// generally return KeyRune here; the specific key can be further decoded
// using the Rune() function.
func (ev *EventKey) Key() Key {
	return ev.key
}

func (ev *EventKey) KeyAsRune() rune {
	return rune(ev.key)
}

// Modifiers returns the modifiers that were present with the key press.  Note
// that not all platforms and terminals support this equally well, and some
// cases we will not not know for sure.  Hence, applications should avoid
// using this in most circumstances.
func (ev *EventKey) Modifiers() ModMask {
	return ev.mod
}

// Name returns a printable value or the key stroke.  This can be used
// when printing the event, for example.
func (ev *EventKey) Name() string {
	s := ""
	var m []string
	if ev.mod&ModCtrl != 0 {
		m = append(m, "Ctrl")
	}
	if ev.mod&ModShift != 0 {
		m = append(m, "Shift")
	}
	if ev.mod&ModAlt != 0 {
		m = append(m, "Alt")
	}
	if ev.mod&ModMeta != 0 {
		m = append(m, "Meta")
	}

	ok := false
	if s, ok = KeyNames[ev.key]; !ok {
		if ev.key == KeyRune {
			s = "Rune[" + string(ev.ch) + "]"
		} else {
			s = fmt.Sprintf("Key[%d,%d]", ev.key, int(ev.ch))
		}
	}
	if len(m) != 0 {
		if ev.mod&ModCtrl != 0 && strings.HasPrefix(s, "Ctrl-") {
			s = s[5:]
		}
		return fmt.Sprintf("%s+%s", strings.Join(m, "+"), s)
	}
	return s
}

// NewEventKey attempts to create a suitable event.  It parses the various
// ASCII control sequences if KeyRune is passed for Key, but if the caller
// has more precise information it should set that specifically.  Callers
// that aren't sure about modifier state (most) should just pass ModNone.
func NewEventKey(k Key, ch rune, mod ModMask) *EventKey {
	if k == KeyRune && (ch < ' ' || ch == 0x7f) {
		// Turn specials into proper key codes.  This is for
		// control characters and the DEL.
		k = Key(ch)
		if mod == ModNone && ch < ' ' {
			switch Key(ch) {
			case KeyBackspace, KeyTab, KeyEsc, KeyEnter:
				// these keys are directly typeable without CTRL
			default:
				// most likely entered with a CTRL keypress
				mod = ModCtrl
			}
		}
	}
	// translate KeyCtrl[A-Z] into KeySmall[A-Z] and ModCtrl, if necessary
	if kk, mm, ok := DecodeCtrlKey(k); ok {
		k = kk
		if !mod.Has(ModCtrl) {
			mod |= mm
		}
	}
	return &EventKey{t: time.Now(), key: k, ch: ch, mod: mod}
}
