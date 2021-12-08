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
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
)

var cdkTimeouts = &timers{
	timers: make(map[uuid.UUID]*timer, 0),
}

type timers struct {
	timers map[uuid.UUID]*timer

	sync.RWMutex
}

func (t *timers) Add(n *timer) (id uuid.UUID) {
	t.Lock()
	if n.id == uuid.Nil {
		n.id, _ = uuid.NewV4()
	}
	t.timers[n.id] = n
	t.Unlock()
	_ = n.display.AsyncCall(func(d Display) error {
		t.Lock()
		n.handler()
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
			t.timers[id].cancel()
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
	created time.Time
	display *CDisplay
	context context.Context
	cancel  context.CancelFunc
}

func (t *timer) handler() {
	if t.display.IsRunning() {
		Go(func() {
			delta := time.Now().Sub(t.created)
			delay := t.delay
			if delta > delay {
				// catch longer than delay to fire, 1ms-delay
				delay = time.Millisecond
			} else if delay >= delta {
				// delay has room, subtract delta so fires closer to correct
				delay -= delta
				if delay <= 0 {
					// catch floor, 1ms-delay
					delay = time.Millisecond
				}
			}
			select {
			case <-time.NewTimer(delay).C:
				if f := t.fn(); f == enums.EVENT_STOP {
					log.TraceF("stopping timeout, fn wants EVENT_STOP: %v", t.id)
					cdkTimeouts.Stop(t.id)
				} else {
					log.TraceF("restarting timeout, fn wants EVENT_PASS: %v", t.id)
					t.context, t.cancel = context.WithCancel(context.Background())
					cdkTimeouts.Add(t)
				}
			case <-t.context.Done():
				log.TraceF("aborting timeout, cancel() received: %v", t.id)
			}
		})
	}

}

type TimerCallbackFn = func() enums.EventFlag

func AddTimeout(delay time.Duration, fn TimerCallbackFn) (id uuid.UUID) {
	if display := GetDefaultDisplay(); display != nil {
		t := &timer{
			id:      uuid.Nil,
			delay:   delay,
			fn:      fn,
			display: display,
			created: time.Now(),
		}
		t.context, t.cancel = context.WithCancel(context.Background())
		id = cdkTimeouts.Add(t)
	}
	return
}

func StopTimeout(id uuid.UUID) {
	if ac, err := GetLocalContext(); err != nil {
		log.WarnDF(1, "error getting app context for CancelTimeout()")
	} else {
		if t := cdkTimeouts.Get(id); t != nil {
			if ac.Display.ObjectID() == t.display.ObjectID() {
				cdkTimeouts.Stop(t.id)
			} else {
				log.WarnDF(1, "cannot stop timeout associated with a different display: %v", id)
			}
		}
	}
	return
}

func CancelAllTimeouts() {
	for _, t := range cdkTimeouts.timers {
		StopTimeout(t.id)
	}
}
