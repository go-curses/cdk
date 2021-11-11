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
	"errors"
	"time"

	"github.com/go-curses/terminfo"
)

var (
	// ErrTermNotFound indicates that a suitable terminal entry could
	// not be found.  This can result from either not having TERM set,
	// or from the TERM failing to support certain minimal functionality,
	// in particular absolute cursor addressability (the cup capability)
	// is required.  For example, legacy "adm3" lacks this capability,
	// whereas the slightly newer "adm3a" supports it.  This failure
	// occurs most often with "dumb".
	ErrTermNotFound = terminfo.ErrTermNotFound

	// ErrNoDisplay indicates that no suitable display could be found.
	// This may result from attempting to run on a platform where there
	// is no support for either termios or console I/O (such as nacl),
	// or from running in an environment where there is no access to
	// a suitable console/terminal device.  (For example, running on
	// without a controlling TTY or with no /dev/tty on POSIX platforms.)
	ErrNoDisplay = errors.New("no suitable display available")

	// ErrNoCharset indicates that the locale environment the
	// program is not supported by the program, because no suitable
	// encoding was found for it.  This problem never occurs if
	// the environment is UTF-8 or UTF-16.
	ErrNoCharset = errors.New("character set not supported")

	// ErrEventQFull indicates that the event queue is full, and
	// cannot accept more events.
	ErrEventQFull = errors.New("event queue full")
)

// An EventError is an event representing some sort of error, and carries
// an error payload.
type EventError struct {
	t   time.Time
	err error
}

// When returns the time when the event was created.
func (ev *EventError) When() time.Time {
	return ev.t
}

// Error implements the error.
func (ev *EventError) Error() string {
	return ev.err.Error()
}

func (ev *EventError) Err() error {
	return ev.err
}

func (ev *EventError) Clone() *EventError {
	return &EventError{t: ev.t, err: ev.err}
}

// NewEventError creates an ErrorEvent with the given error payload.
func NewEventError(err error) *EventError {
	return &EventError{t: time.Now(), err: err}
}
