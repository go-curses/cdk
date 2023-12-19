// Copyright (c) 2022-2023  The Go-Curses Authors
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
	"time"
)

// EventQuit is sent when the display needs to render the screen
type EventQuit struct {
	when time.Time
}

func NewEventQuit() *EventQuit {
	return &EventQuit{when: time.Now()}
}

// When returns the time when the EventQuit was created
func (ev *EventQuit) When() time.Time {
	return ev.when
}
