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
	TypeSignaling       CTypeTag = "cdk-signaling"
	SignalSignalingInit Signal   = "signaling-init"
)

func init() {
	_ = TypesManager.AddType(TypeSignaling, nil)
}

type Signaling interface {
	TypeItem

	Connect(signal Signal, handle string, c SignalListenerFn, data ...interface{})
	Disconnect(signal Signal, handle string) error
	Emit(signal Signal, argv ...interface{}) enums.EventFlag
	StopSignal(signal Signal)
	IsSignalStopped(signal Signal) bool
	PassSignal(signal Signal)
	IsSignalPassed(signal Signal) bool
	ResumeSignal(signal Signal)
	Freeze()
	Thaw()
	IsFrozen() bool
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
	o.Emit(SignalSignalingInit)
	return false
}

func (o *CSignaling) Handled(signal Signal, handle string) (found bool) {
	if listeners, ok := o.listeners[signal]; ok {
		for _, listener := range listeners {
			if listener.n == handle {
				return true
			}
		}
	}
	return false
}

// Connect callback to signal, identified by handle
func (o *CSignaling) Connect(signal Signal, handle string, c SignalListenerFn, data ...interface{}) {
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
				return
			}
		}
	}
	log.TraceDF(1, "connecting %v listener with handler: %v", signal, handle)
	o.listeners[signal] = append(o.listeners[signal], newSignalListener(signal, handle, c, data))
}

// Disconnect callback from signal identified by handle
func (o *CSignaling) Disconnect(signal Signal, handle string) error {
	if listeners, ok := o.listeners[signal]; ok {
		for idx, listener := range listeners {
			if listener.n == handle {
				o.listeners[signal] = append(o.listeners[signal][:idx], o.listeners[signal][idx+1:]...)
				o.LogTrace("disconnected %v listener: %v", signal, handle)
				return nil
			}
		}
		return fmt.Errorf("%v signal handler not found: %v", signal, handle)
	}
	return fmt.Errorf("signal not found: %v", signal)
}

// Emit a signal event to all connected listener callbacks
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

// Disable propagation of the given signal
func (o *CSignaling) StopSignal(signal Signal) {
	if !o.IsSignalStopped(signal) {
		o.LogTrace("stopping %v signal", signal)
		o.stopped = append(o.stopped, signal)
	}
}

func (o *CSignaling) IsSignalStopped(signal Signal) bool {
	return o.getSignalStopIndex(signal) >= 0
}

func (o *CSignaling) getSignalStopIndex(signal Signal) int {
	for idx, stop := range o.stopped {
		if signal == stop {
			return idx
		}
	}
	return -1
}

func (o *CSignaling) PassSignal(signal Signal) {
	if !o.IsSignalPassed(signal) {
		o.LogTrace("passing %v signal", signal)
		o.passed = append(o.passed, signal)
	}
}

func (o *CSignaling) IsSignalPassed(signal Signal) bool {
	return o.getSignalPassIndex(signal) >= 0
}

func (o *CSignaling) getSignalPassIndex(signal Signal) int {
	for idx, stop := range o.passed {
		if signal == stop {
			return idx
		}
	}
	return -1
}

// Enable propagation of the given signal
func (o *CSignaling) ResumeSignal(signal Signal) {
	id := o.getSignalStopIndex(signal)
	if id >= 0 {
		o.LogTrace("resuming %v signal from being stopped", signal)
		if len(o.stopped) > 1 {
			o.stopped = append(
				o.stopped[:id],
				o.stopped[id+1:]...,
			)
		} else {
			o.stopped = []Signal{}
		}
		return
	}
	id = o.getSignalPassIndex(signal)
	if id >= 0 {
		o.LogTrace("resuming %v signal from being passed", signal)
		if len(o.passed) > 1 {
			o.passed = append(
				o.passed[:id],
				o.passed[id+1:]...,
			)
		} else {
			o.passed = []Signal{}
		}
		return
	}
	if _, ok := o.listeners[signal]; ok {
		o.LogWarn("%v signal already resumed", signal)
	} else {
		o.LogError("failed to resume unknown signal: %v", signal)
	}
}

func (o *CSignaling) Freeze() {
	o.frozen += 1
}

func (o *CSignaling) Thaw() {
	o.frozen -= 1
}

func (o *CSignaling) IsFrozen() bool {
	return o.frozen > 0
}
