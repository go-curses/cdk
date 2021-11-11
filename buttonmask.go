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
	"strconv"
)

// ButtonMask is a mask of mouse buttons and wheel events.  Mouse button presses
// are normally delivered as both press and release events.  Mouse wheel events
// are normally just single impulse events.  Windows supports up to eight
// separate buttons plus all four wheel directions, but XTerm can only support
// mouse buttons 1-3 and wheel up/down.  Its not unheard of for terminals
// to support only one or two buttons (think Macs).  Old terminals, and true
// emulations (such as vt100) won't support mice at all, of course.
type ButtonMask int16

// These are the actual button values.  Note that tcell version 1.x reversed buttons
// two and three on *nix based terminals.  We use button 1 as the primary, and
// button 2 as the secondary, and button 3 (which is often missing) as the middle.
const (
	Button1 ButtonMask = 1 << iota // Usually left mouse button.
	Button2                        // Usually the middle mouse button.
	Button3                        // Usually the right mouse button.
	Button4                        // Often a side button (thumb/next).
	Button5                        // Often a side button (thumb/prev).
	Button6
	Button7
	Button8
	WheelUp                       // Wheel motion up/away from user.
	WheelDown                     // Wheel motion down/towards user.
	WheelLeft                     // Wheel motion to left.
	WheelRight                    // Wheel motion to right.
	LastButtonMask                // Highest mask value
	ButtonNone     ButtonMask = 0 // No button or wheel events.

	ButtonPrimary   = Button1
	ButtonSecondary = Button2
	ButtonMiddle    = Button3
)

type IButtonMask interface {
	Has(m ButtonMask) bool
	Set(m ButtonMask) ButtonMask
	Clear(m ButtonMask) ButtonMask
	Toggle(m ButtonMask) ButtonMask
	String() string
}

// check if the mask has the given flag(s)
func (i ButtonMask) Has(m ButtonMask) bool {
	return i&m != 0
}

// return a button mask with the given flags set, does not modify itself
func (i ButtonMask) Set(m ButtonMask) ButtonMask {
	return i | m
}

// return a button mask with the given flags cleared, does not modify itself
func (i ButtonMask) Clear(m ButtonMask) ButtonMask {
	return i &^ m
}

// return a button mask with the give flags reversed, does not modify itself
func (i ButtonMask) Toggle(m ButtonMask) ButtonMask {
	return i ^ m
}

const buttonMaskName = "ButtonNoneButton1Button2Button3Button4Button5Button6Button7Button8WheelUpWheelDownWheelLeftWheelRight"

var buttonMaskMap = map[ButtonMask]string{
	0:    buttonMaskName[0:10],
	1:    buttonMaskName[10:17],
	2:    buttonMaskName[17:24],
	4:    buttonMaskName[24:31],
	8:    buttonMaskName[31:38],
	16:   buttonMaskName[38:45],
	32:   buttonMaskName[45:52],
	64:   buttonMaskName[52:59],
	128:  buttonMaskName[59:66],
	256:  buttonMaskName[66:73],
	512:  buttonMaskName[73:82],
	1024: buttonMaskName[82:91],
	2048: buttonMaskName[91:101],
}

func (i ButtonMask) String() string {
	if str, ok := buttonMaskMap[i]; ok {
		return str
	}
	return "ButtonMask(" + strconv.FormatInt(int64(i), 10) + ")"
}
