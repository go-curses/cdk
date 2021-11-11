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

package enums

import (
	"fmt"
	"strings"
)

// enums, flags, tags, etc

type EnumFromString interface {
	FromString(value string) (enum interface{}, err error)
}

/* Window type */
type WindowType uint64

const (
	WINDOW_TOPLEVEL WindowType = iota
	WINDOW_POPUP
)

/* Wrap mode */
type WrapMode uint64

const (
	WRAP_NONE WrapMode = iota
	WRAP_CHAR
	WRAP_WORD
	WRAP_WORD_CHAR
)

func (m WrapMode) FromString(value string) (enum interface{}, err error) {
	switch strings.ToLower(value) {
	case "none":
		return WRAP_NONE, nil
	case "char":
		return WRAP_CHAR, nil
	case "word":
		return WRAP_WORD, nil
	case "wordchar", "word_char", "word-char":
		return WRAP_WORD_CHAR, nil
	}
	return WRAP_NONE, fmt.Errorf("invalid WrapMode value: %v", value)
}

/* Drag result */
type DragResult uint64

const (
	DRAG_RESULT_SUCCESS DragResult = iota
	DRAG_RESULT_NO_TARGET
	DRAG_RESULT_USER_CANCELLED
	DRAG_RESULT_TIMEOUT_EXPIRED
	DRAG_RESULT_GRAB_BROKEN
	DRAG_RESULT_ERROR
)

/* Dest defaults */
type DestDefaults uint64

const (
	DEST_DEFAULT_MOTION    DestDefaults = 1 << 0
	DEST_DEFAULT_HIGHLIGHT DestDefaults = 1 << iota
	DEST_DEFAULT_DROP
	DEST_DEFAULT_ALL DestDefaults = 0
)

/* Target flags */
type TargetFlags uint64

const (
	TARGET_SAME_APP    TargetFlags = 1 << 0
	TARGET_SAME_WIDGET TargetFlags = 1 << iota
	TARGET_OTHER_APP
	TARGET_OTHER_WIDGET
)

/* Object flags */
type ObjectFlags uint64

const (
	IN_DESTRUCTION ObjectFlags = 1 << 0
	FLOATING       ObjectFlags = 1 << iota
	RESERVED_1
	RESERVED_2
)

// CEvent handling event flag type
type EventFlag int

// CEvent handling event flags
const (
	EVENT_PASS EventFlag = iota // Allow other handlers to process
	EVENT_STOP                  // Prevent further event handling
)

/* Signal flags */
type SignalFlags uint64

const (
	SIGNAL_RUN_FIRST SignalFlags = 1 << 0
	SIGNAL_RUN_LAST  SignalFlags = 1 << iota
	SIGNAL_RUN_CLEANUP
	SIGNAL_NO_RECURSE
	SIGNAL_DETAILED
	SIGNAL_ACTION
	SIGNAL_NO_HOOKS
	SIGNAL_MUST_COLLECT
	SIGNAL_DEPRECATED
)

/* Connect flags */
type ConnectFlags uint64

const (
	CONNECT_AFTER   ConnectFlags = 1 << 0
	CONNECT_SWAPPED ConnectFlags = 1 << iota
)

/* Signal match type */
type SignalMatchType uint64

const (
	SIGNAL_MATCH_ID     SignalMatchType = 1 << 0
	SIGNAL_MATCH_DETAIL SignalMatchType = 1 << iota
	SIGNAL_MATCH_CLOSURE
	SIGNAL_MATCH_FUNC
	SIGNAL_MATCH_DATA
	SIGNAL_MATCH_UNBLOCKED
)

/* Signal run type */
type SignalRunType uint64

const (
	RUN_FIRST      SignalRunType = SignalRunType(SIGNAL_RUN_FIRST)
	RUN_LAST       SignalRunType = SignalRunType(SIGNAL_RUN_LAST)
	RUN_BOTH       SignalRunType = SignalRunType(RUN_FIRST | RUN_LAST)
	RUN_NO_RECURSE SignalRunType = SignalRunType(SIGNAL_NO_RECURSE)
	RUN_ACTION     SignalRunType = SignalRunType(SIGNAL_ACTION)
	RUN_NO_HOOKS   SignalRunType = SignalRunType(SIGNAL_NO_HOOKS)
)

type HorizontalAlignment uint

const (
	ALIGN_LEFT   HorizontalAlignment = 0
	ALIGN_RIGHT  HorizontalAlignment = 1
	ALIGN_CENTER HorizontalAlignment = 2
)

type VerticalAlignment uint

const (
	ALIGN_TOP    VerticalAlignment = 0
	ALIGN_BOTTOM VerticalAlignment = 1
	ALIGN_MIDDLE VerticalAlignment = 2
)

/* Resize mode */
type ResizeMode uint64

const (
	RESIZE_PARENT ResizeMode = iota
	RESIZE_QUEUE
	RESIZE_IMMEDIATE
)

/* Justification */
type Justification uint64

const (
	JUSTIFY_LEFT Justification = iota
	JUSTIFY_RIGHT
	JUSTIFY_CENTER
	JUSTIFY_FILL
)

/* Orientation */
type Orientation uint64

func (o Orientation) FromString(value string) (enum interface{}, err error) {
	switch strings.ToLower(value) {
	case "horizontal":
		return ORIENTATION_HORIZONTAL, nil
	case "vertical":
		return ORIENTATION_VERTICAL, nil
	}
	return ORIENTATION_NONE, fmt.Errorf("invalid orientation value: %v", value)
}

const (
	ORIENTATION_NONE Orientation = iota
	ORIENTATION_HORIZONTAL
	ORIENTATION_VERTICAL
)
