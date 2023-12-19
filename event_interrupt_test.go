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
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInterrupt(t *testing.T) {
	Convey("Interrupt checks", t, func() {
		then := time.Now()
		ei := NewEventInterrupt(nil)
		So(ei, ShouldHaveSameTypeAs, &EventInterrupt{})
		So(ei.Data(), ShouldEqual, nil)
		now := time.Now()
		So(ei.When().UnixNano(), ShouldBeGreaterThanOrEqualTo, then.UnixNano())
		So(ei.When().UnixNano(), ShouldBeLessThanOrEqualTo, now.UnixNano())
	})
}
