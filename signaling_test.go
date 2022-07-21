// Copyright 2022  The CDK Authors
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
	"testing"

	"github.com/go-curses/cdk/lib/enums"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignalingBasics(t *testing.T) {
	Convey("Basic Signaling Features", t, func() {
		s := new(CSignaling)
		So(s.Init(), ShouldEqual, false)
		So(s.Init(), ShouldEqual, true)
		someData := "some data"
		signalCaught := false
		s.Connect(
			SignalEventError,
			"basic-test",
			func(data []interface{}, argv ...interface{}) enums.EventFlag {
				So(data, ShouldHaveLength, 1)
				So(data[0], ShouldEqual, "some data")
				So(argv, ShouldHaveLength, 1)
				err := fmt.Errorf("an error")
				So(argv[0], ShouldHaveSameTypeAs, err)
				So(argv[0].(error).Error(), ShouldEqual, err.Error())
				signalCaught = true
				return enums.EVENT_PASS
			},
			someData,
		)
		So(signalCaught, ShouldEqual, false)
		So(
			s.Emit(SignalEventError, fmt.Errorf("an error")),
			ShouldEqual,
			enums.EVENT_PASS,
		)
		So(signalCaught, ShouldEqual, true)
		// So(s.Disconnect(SignalEventError, "basic-test"), ShouldBeNil)
	})
}

func TestSignalingPatterns(t *testing.T) {
	var hit0, hit1, hit2 bool
	hit0, hit1, hit2 = false, false, false
	hit0fn := func(data []interface{}, argv ...interface{}) enums.EventFlag {
		hit0 = true
		return enums.EVENT_PASS
	}
	hit1fn := func(data []interface{}, argv ...interface{}) enums.EventFlag {
		hit1 = true
		return enums.EVENT_STOP
	}
	hit2fn := func(data []interface{}, argv ...interface{}) enums.EventFlag {
		hit2 = true
		return enums.EVENT_PASS
	}
	s := new(CSignaling)
	Convey("Signaling Init", t, func() {
		So(s, ShouldNotBeNil)
		So(s.Init(), ShouldEqual, false)
		So(s.Init(), ShouldEqual, true)
	})
	Convey("Signaling Listeners", t, func() {
		s.Connect(SignalEventError, "many-errors-0", hit0fn)
		s.Connect(SignalEventError, "many-errors-1", hit1fn)
		s.Connect(SignalEventError, "many-errors-2", hit2fn)
		hit0, hit1, hit2 = false, false, false
		So(
			s.Emit(SignalEventError, fmt.Errorf("an error")),
			ShouldEqual,
			enums.EVENT_STOP,
		)
		So(hit0, ShouldEqual, false)
		So(hit1, ShouldEqual, true)
		So(hit2, ShouldEqual, true)
		So(s.Disconnect(SignalEventError, "many-errors-1"), ShouldBeNil)
		hit0, hit1, hit2 = false, false, false
		So(
			s.Emit(SignalEventError, fmt.Errorf("an error")),
			ShouldEqual,
			enums.EVENT_PASS,
		)
		So(hit0, ShouldEqual, true)
		So(hit1, ShouldEqual, false)
		So(hit2, ShouldEqual, true)
		err := s.Disconnect(SignalEventError, "many-errors-1")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "event-error signal handler not found: many-errors-1")
	})
	Convey("Signaling Regulating", t, func() {
		So(s.Disconnect(SignalEventError, "many-errors-0"), ShouldBeNil)
		So(s.Disconnect(SignalEventError, "many-errors-1"), ShouldNotBeNil)
		So(s.Disconnect(SignalEventError, "many-errors-2"), ShouldBeNil)
		s.StopSignal(SignalEventError)
		So(s.IsSignalStopped(SignalEventError), ShouldEqual, true)
		So(s.IsSignalPassed(SignalEventError), ShouldEqual, false)
		So(
			s.Emit(SignalEventError, fmt.Errorf("an error dropped")),
			ShouldEqual,
			enums.EVENT_STOP,
		)
		s.ResumeSignal(SignalEventError)
		s.ResumeSignal(SignalEventError)
		s.ResumeSignal("this is not really signal")
		So(s.IsSignalStopped(SignalEventError), ShouldEqual, false)
		So(s.IsSignalPassed(SignalEventError), ShouldEqual, false)
		So(
			s.Emit(SignalEventError, fmt.Errorf("an error stopped")),
			ShouldEqual,
			enums.EVENT_PASS,
		)
		s.Connect(SignalEventError, "many-errors-1", hit1fn)
		s.PassSignal(SignalEventError)
		So(s.IsSignalStopped(SignalEventError), ShouldEqual, false)
		So(s.IsSignalPassed(SignalEventError), ShouldEqual, true)
		So(
			s.Emit(SignalEventError, fmt.Errorf("an error passed")),
			ShouldEqual,
			enums.EVENT_PASS,
		)
		s.ResumeSignal(SignalEventError)
	})
}
