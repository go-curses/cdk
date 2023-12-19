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

func TestResizeEvent(t *testing.T) {
	Convey("Errors checks", t, func() {
		then := time.Now()
		er := NewEventResize(10, 100)
		So(er, ShouldHaveSameTypeAs, &EventResize{})
		now := time.Now()
		So(er.When().UnixNano(), ShouldBeGreaterThanOrEqualTo, then.UnixNano())
		So(er.When().UnixNano(), ShouldBeLessThanOrEqualTo, now.UnixNano())
		h, w := er.Size()
		So(h, ShouldEqual, 10)
		So(w, ShouldEqual, 100)
	})
}
