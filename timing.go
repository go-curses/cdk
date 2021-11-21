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
	"sync"
	"time"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/log"
	"github.com/gofrs/uuid"
)

// var cdkTimeouts *timers
var cdkTimeouts = &timers{
	timers: make(map[uuid.UUID]*timer, 0),
}

type timers struct {
	timers map[uuid.UUID]*timer

	sync.RWMutex
}

func (t *timers) Add(n *timer) (id uuid.UUID) {
	t.Lock()
	n.id, _ = uuid.NewV4()
	t.timers[n.id] = n
	t.Unlock()
	_ = n.display.AsyncCall(func(d Display) error {
		t.Lock()
		n.timer = time.AfterFunc(n.delay, n.handler)
		t.Unlock()
		return nil
	})
	id = n.id
	return
}

func (t *timers) Valid(id uuid.UUID) bool {
	t.RLock()
	defer t.RUnlock()
	_, ok := t.timers[id]
	return ok
}

func (t *timers) Get(id uuid.UUID) *timer {
	if t.Valid(id) {
		t.RLock()
		defer t.RUnlock()
		return t.timers[id]
	}
	return nil
}

func (t *timers) Stop(id uuid.UUID) bool {
	if t.Valid(id) {
		t.Lock()
		defer t.Unlock()
		_ = t.timers[id].display.AwaitCall(func(_ Display) error {
			t.timers[id].timer.Stop()
			t.timers[id] = nil
			log.TraceF("stopped timer: %d", id)
			return nil
		})
		return true
	}
	return false
}

type timer struct {
	id      uuid.UUID
	delay   time.Duration
	fn      TimerCallbackFn
	display *CDisplay
	timer   *time.Timer
}

func (t *timer) handler() {
	if t.display.IsRunning() {
		if f := t.fn(); f == enums.EVENT_STOP {
			cdkTimeouts.Stop(t.id)
			return
		}
	}
	t.timer.Stop()
	t.timer = time.AfterFunc(t.delay, t.handler)
}

type TimerCallbackFn = func() enums.EventFlag

func AddTimeout(delay time.Duration, fn TimerCallbackFn) (id uuid.UUID) {
	if display := GetDefaultDisplay(); display != nil {
		t := &timer{
			delay:   delay,
			fn:      fn,
			display: display,
		}
		id = cdkTimeouts.Add(t)
	}
	return
}

func StopTimeout(id uuid.UUID) {
	if ac, err := GetLocalContext(); err != nil {
		log.WarnDF(1, "error getting app context for CancelTimeout()")
	} else {
		for cdkTimeouts.Valid(id) {
			if t := cdkTimeouts.Get(id); t != nil {
				if ac.Display.ObjectName() == t.display.ObjectName() {
					cdkTimeouts.Stop(t.id)
				}
			}
		}
	}
	return
}

func CancelAllTimeouts() {
	if ac, err := GetLocalContext(); err != nil {
		log.WarnDF(1, "error getting app context for CancelAllTimeouts()")
	} else {
		for _, t := range cdkTimeouts.timers {
			if t != nil && ac.Display.ObjectName() == t.display.ObjectName() {
				cdkTimeouts.Stop(t.id)
			}
		}
	}
}
