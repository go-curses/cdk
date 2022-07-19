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

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
)

func TestObject(t *testing.T) {
	Convey("Basic Object Features", t, func() {
		o := &CObject{}
		So(o, ShouldImplement, (*Object)(nil))
		So(o.Init(), ShouldEqual, false)
		So(o.Init(), ShouldEqual, true)
		// normal testing
		So(o.GetTheme().String(), ShouldEqual, paint.GetDefaultColorTheme().String())
		o.SetTheme(paint.GetDefaultMonoTheme())
		So(o.GetTheme().String(), ShouldEqual, paint.GetDefaultMonoTheme().String())
		So(o.IsProperty("testing"), ShouldEqual, false)
		So(o.InstallProperty("debug", BoolProperty, true, false), ShouldNotBeNil)
		So(o.InstallProperty("testing", BoolProperty, true, false), ShouldBeNil)
		So(o.IsProperty("testing"), ShouldEqual, true)
		So(o.SetBoolProperty("testing", false), ShouldBeNil)
		So(o.SetStringProperty("testing", ""), ShouldNotBeNil)
		So(o.SetIntProperty("testing", 0), ShouldNotBeNil)
		So(o.SetFloatProperty("testing", 0.0), ShouldNotBeNil)
		// destruction testing
		hit0 := false
		o.Connect(SignalDestroy, "basic-destroy", func(data []interface{}, argv ...interface{}) enums.EventFlag {
			hit0 = true
			return enums.EVENT_STOP
		})
		o.Destroy()
		So(hit0, ShouldEqual, true)
		So(o.IsValid(), ShouldEqual, true)
		hit0 = false
		o.Disconnect(SignalDestroy, "basic-destroy")
		o.Connect(SignalDestroy, "basic-destroy", func(data []interface{}, argv ...interface{}) enums.EventFlag {
			hit0 = true
			return enums.EVENT_PASS
		})
		o.Destroy()
		So(hit0, ShouldEqual, true)
		So(o.IsValid(), ShouldEqual, false)
	})
}