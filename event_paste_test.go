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
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEventPaste(t *testing.T) {
	Convey("EventPaste basics", t, func() {
		then := time.Now()
		ep := NewEventPaste(false)
		So(ep, ShouldHaveSameTypeAs, &EventPaste{})
		now := time.Now()
		So(ep.When().UnixNano(), ShouldBeGreaterThanOrEqualTo, then.UnixNano())
		So(ep.When().UnixNano(), ShouldBeLessThanOrEqualTo, now.UnixNano())
		So(ep.Start(), ShouldEqual, false)
		So(ep.End(), ShouldEqual, true)
		ep = NewEventPaste(true)
		So(ep.Start(), ShouldEqual, true)
		So(ep.End(), ShouldEqual, false)
	})
}
