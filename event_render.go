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
	"time"
)

// EventRender is sent when the display needs to render the screen
type EventRender struct {
	when time.Time
	draw bool
	show bool
	sync bool
}

func NewEventRender(draw, show, sync bool) *EventRender {
	return &EventRender{when: time.Now(), draw: draw, show: show, sync: sync}
}

// NewEventDraw creates an EventRender
func NewEventDraw() *EventRender {
	return &EventRender{when: time.Now(), draw: true, show: false, sync: false}
}

// NewEventShow creates an EventRender configured to just show the screen
func NewEventShow() *EventRender {
	return &EventRender{when: time.Now(), draw: false, show: true, sync: false}
}

// NewEventSync creates an EventRender configured to just sync the screen
func NewEventSync() *EventRender {
	return &EventRender{when: time.Now(), draw: false, show: false, sync: true}
}

// NewEventDrawAndShow creates an EventRender configured to also request a show
// after the draw cycle completes
func NewEventDrawAndShow() *EventRender {
	return &EventRender{when: time.Now(), draw: true, show: true, sync: false}
}

// NewEventDrawAndSync creates an EventRender configured to also request a sync
// after the draw cycle completes
func NewEventDrawAndSync() *EventRender {
	return &EventRender{when: time.Now(), draw: true, show: false, sync: true}
}

// When returns the time when the EventRender was created
func (ev *EventRender) When() time.Time {
	return ev.when
}

// Draw returns true when the EventRender was created with one of the
// NewEventDraw*() functions
func (ev *EventRender) Draw() bool {
	return ev.draw
}

// Show returns true when the EventRender was created with either of the
// NewEventShow() or NewDrawAndShow() functions
func (ev *EventRender) Show() bool {
	return ev.show
}

// Sync returns true when the EventRender was created with either of the
// NewEventSync() or NewDrawAndSync() functions
func (ev *EventRender) Sync() bool {
	return ev.sync
}