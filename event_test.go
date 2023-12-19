// Copyright (c) 2022-2023  The Go-Curses Authors
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
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEventTime(t *testing.T) {
	Convey("EventTime checks", t, func() {
		t := time.Unix(0, 0)
		et := NewEventTime(t)
		So(et, ShouldHaveSameTypeAs, &EventTime{})
		So(et.When(), ShouldEqual, t)
		now := time.Now()
		et.SetEventNow()
		So(et.When().UnixNano(), ShouldBeGreaterThan, now.UnixNano())
	})
}
