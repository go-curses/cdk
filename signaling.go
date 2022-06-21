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
	"fmt"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/log"
)

const (
	TypeSignaling CTypeTag = "cdk-signaling"
)

func init() {
	_ = TypesManager.AddType(TypeSignaling, nil)
}

type Signaling interface {
	TypeItem

	Init() (already bool)
	Handled(signal Signal, handle string) (found bool)
	Connect(signal Signal, handle string, c SignalListenerFn, data ...interface{})
	Disconnect(signal Signal, handle string) error
	Emit(signal Signal, argv ...interface{}) enums.EventFlag
	HasListeners(signal Signal) (has bool)
	DisconnectAll()
	StopSignal(signals ...Signal)
	IsSignalStopped(signals ...Signal) (stopped bool)
	PassSignal(signals ...Signal)
	IsSignalPassed(signals ...Signal) (passed bool)
	ResumeSignal(signals ...Signal)
	Freeze()
	Thaw()
	IsFrozen() (frozen bool)
}

type CSignaling struct {
	CTypeItem

	frozen    uint
	stopped   []Signal
	passed    []Signal
	listeners map[Signal][]*CSignalListener
}

func (o *CSignaling) Init() (already bool) {
	if o.InitTypeItem(TypeSignaling, o) {
		return true
	}
	o.CTypeItem.Init()
	o.frozen = 0
	o.stopped = []Signal{}
	o.passed = []Signal{}
	if o.listeners == nil {
		o.listeners = make(map[Signal][]*CSignalListener)
	}
	return false
}

// Handled returns TRUE if there is at least one signal listener with the given
// handle.
//
// Locking: read
func (o *CSignaling) Handled(signal Signal, handle string) (found bool) {
	o.RLock()
	if listeners, ok := o.listeners[signal]; ok {
		for _, listener := range listeners {
			if listener.n == handle {
				o.RUnlock()
				return true
			}
		}
	}
	o.RUnlock()
	return false
}

// Connect callback to signal, identified by handle
//
// Locking: write
func (o *CSignaling) Connect(signal Signal, handle string, c SignalListenerFn, data ...interface{}) {
	o.Lock()
	if o.listeners == nil {
		o.listeners = make(map[Signal][]*CSignalListener)
	}
	if _, ok := o.listeners[signal]; !ok {
		o.listeners[signal] = make([]*CSignalListener, 0)
	}
	if listeners, ok := o.listeners[signal]; ok {
		for idx, listener := range listeners {
			if listener.n == handle {
				log.TraceDF(1, "replacing %v listener for handler: %v", signal, handle)
				o.listeners[signal][idx].c = c
				o.listeners[signal][idx].d = data
				o.Unlock()
				return
			}
		}
	}
	log.TraceDF(1, "connecting %v listener with handler: %v", signal, handle)
	o.listeners[signal] = append(o.listeners[signal], newSignalListener(signal, handle, c, data))
	o.Unlock()
}

// Disconnect callback from signal identified by handle
//
// Locking: write
func (o *CSignaling) Disconnect(signal Signal, handle string) error {
	o.Lock()
	if listeners, ok := o.listeners[signal]; ok {
		for idx, listener := range listeners {
			if listener.n == handle {
				o.listeners[signal] = append(o.listeners[signal][:idx], o.listeners[signal][idx+1:]...)
				o.LogTrace("disconnected %v listener: %v", signal, handle)
				o.Unlock()
				return nil
			}
		}
		o.Unlock()
		return fmt.Errorf("%v signal handler not found: %v", signal, handle)
	}
	o.Unlock()
	return fmt.Errorf("signal not found: %v", signal)
}

// Emit a signal event to all connected listener callbacks
//
// Locking: none
func (o *CSignaling) Emit(signal Signal, argv ...interface{}) enums.EventFlag {
	if o.frozen > 0 {
		return enums.EVENT_PASS
	}
	if o.IsSignalStopped(signal) {
		return enums.EVENT_STOP
	}
	if o.IsSignalPassed(signal) {
		return enums.EVENT_PASS
	}
	if listeners, ok := o.listeners[signal]; ok {
		if max := len(listeners); max > 0 {
			for i := max - 1; i > -1; i-- {
				listener := listeners[i]
				if r := listener.c(listener.d, argv...); r == enums.EVENT_STOP {
					o.LogTrace("%v signal stopped by listener: %v", signal, listener.n)
					return enums.EVENT_STOP
				}
			}
		}
	}
	return enums.EVENT_PASS
}

// HasListeners returns true if there are one or more listeners connected to the
// given Signal.
func (o *CSignaling) HasListeners(signal Signal) (has bool) {
	if listeners, ok := o.listeners[signal]; ok {
		has = len(listeners) > 0
	}
	return
}

func (o *CSignaling) DisconnectAll() {
	o.RLock()
	signalListeners := o.listeners
	o.RUnlock()
	for signal, listeners := range signalListeners {
		for _, listener := range listeners {
			_ = o.Disconnect(signal, listener.n)
		}
	}
}

// StopSignal disables propagation of the given signal with an EVENT_STOP
//
// Locking: write
func (o *CSignaling) StopSignal(signals ...Signal) {
	for _, signal := range signals {
		if !o.IsSignalStopped(signal) {
			o.LogTrace("stopping %v signal", signal)
			o.Lock()
			o.stopped = append(o.stopped, signal)
			o.Unlock()
		}
	}
}

// IsSignalStopped returns TRUE if the given signal is currently stopped.
//
// Locking: none
func (o *CSignaling) IsSignalStopped(signals ...Signal) (stopped bool) {
	for _, signal := range signals {
		if o.getSignalStopIndex(signal) < 0 {
			return
		}
	}
	stopped = true
	return
}

func (o *CSignaling) getSignalStopIndex(signal Signal) int {
	for idx, stop := range o.stopped {
		if signal == stop {
			return idx
		}
	}
	return -1
}

// PassSignal disables propagation of the given signal with an EVENT_PASS
//
// Locking: write
func (o *CSignaling) PassSignal(signals ...Signal) {
	for _, signal := range signals {
		if !o.IsSignalPassed(signal) {
			o.LogTrace("passing %v signal", signal)
			o.Lock()
			o.passed = append(o.passed, signal)
			o.Unlock()
		}
	}
}

// IsSignalPassed returns TRUE if the given signal is currently passed.
//
// Locking: none
func (o *CSignaling) IsSignalPassed(signals ...Signal) (passed bool) {
	for _, signal := range signals {
		if o.getSignalPassIndex(signal) < 0 {
			return
		}
	}
	passed = true
	return
}

func (o *CSignaling) getSignalPassIndex(signal Signal) int {
	for idx, stop := range o.passed {
		if signal == stop {
			return idx
		}
	}
	return -1
}

// ResumeSignal enables propagation of the given signal if the signal is
// currently stopped.
//
// Locking: write
func (o *CSignaling) ResumeSignal(signals ...Signal) {
	o.Lock()
	for _, signal := range signals {
		if sid := o.getSignalStopIndex(signal); sid >= 0 {
			o.LogTrace("resuming %v stopped signal", signal)
			if len(o.stopped) > 1 {
				o.stopped = append(
					o.stopped[:sid],
					o.stopped[sid+1:]...,
				)
			} else {
				o.stopped = []Signal{}
			}
		}
		if pid := o.getSignalPassIndex(signal); pid >= 0 {
			o.LogTrace("resuming %v passed signal", signal)
			if len(o.passed) > 1 {
				o.passed = append(
					o.passed[:pid],
					o.passed[pid+1:]...,
				)
			} else {
				o.passed = []Signal{}
			}
		}
	}
	o.Unlock()
}

// Freeze pauses all signal emissions until a corresponding Thaw is called.
//
// Locking: write
func (o *CSignaling) Freeze() {
	o.Lock()
	o.frozen += 1
	o.Unlock()
}

// Thaw restores all signal emissions after a Freeze call.
//
// Locking: write
func (o *CSignaling) Thaw() {
	o.Lock()
	if o.frozen <= 0 {
		o.frozen = 0
		o.LogError("Thaw() called too many times")
	} else {
		o.frozen -= 1
	}
	o.Unlock()
}

// IsFrozen returns TRUE if Thaw has been called at least once.
//
// Locking: read, signal read
func (o *CSignaling) IsFrozen() (frozen bool) {
	o.RLock()
	frozen = o.frozen > 0
	o.RUnlock()
	return
}