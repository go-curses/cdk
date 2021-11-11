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

type MouseState uint64

const (
	MOUSE_NONE MouseState = 0
	MOUSE_MOVE MouseState = 1 << iota
	BUTTON_PRESS
	BUTTON_RELEASE
	WHEEL_PULSE
	DRAG_START
	DRAG_MOVE
	DRAG_STOP
)

type IMouseState interface {
	Has(m MouseState) bool
	Set(m MouseState) MouseState
	Clear(m MouseState) MouseState
	Toggle(m MouseState) MouseState
	String() string
}

func (i MouseState) Has(m MouseState) bool {
	return i&m != 0
}

func (i MouseState) Set(m MouseState) MouseState {
	return i | m
}

func (i MouseState) Clear(m MouseState) MouseState {
	return i &^ m
}

func (i MouseState) Toggle(m MouseState) MouseState {
	return i ^ m
}

const (
	_MouseState_name_0 = "MOUSE_NONE"
	_MouseState_name_1 = "MOUSE_MOVE"
	_MouseState_name_2 = "BUTTON_PRESS"
	_MouseState_name_3 = "BUTTON_RELEASE"
	_MouseState_name_4 = "WHEEL_PULSE"
	_MouseState_name_5 = "DRAG_START"
	_MouseState_name_6 = "DRAG_MOVE"
	_MouseState_name_7 = "DRAG_STOP"
)

func (i MouseState) String() string {
	switch {
	case i == 0:
		return _MouseState_name_0
	case i == 2:
		return _MouseState_name_1
	case i == 4:
		return _MouseState_name_2
	case i == 8:
		return _MouseState_name_3
	case i == 16:
		return _MouseState_name_4
	case i == 32:
		return _MouseState_name_5
	case i == 64:
		return _MouseState_name_6
	case i == 128:
		return _MouseState_name_7
	default:
		return "MouseState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
