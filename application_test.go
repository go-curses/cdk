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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCdk(t *testing.T) {
	Convey("Making a new app instance", t, func() {
		Convey("validating factory", func() {
			app := NewApplication(
				"AppName", "AppUsage",
				"AppDesc", "v0.0.0",
				"app-tag", "AppTitle",
				OffscreenTtyPath,
			)
			// app.Connect(SignalStartup, "test-cdk-handler", TestingMakesNoContent)
			So(app, ShouldNotBeNil)
			So(app.Name(), ShouldEqual, "AppName")
			So(app.Usage(), ShouldEqual, "AppUsage")
			So(app.Title(), ShouldEqual, "AppTitle")
			So(app.Version(), ShouldEqual, "v0.0.0")
			So(app.Tag(), ShouldEqual, "app-tag")
			So(app.GetContext(), ShouldBeNil)
			So(app.CLI(), ShouldNotBeNil)
			app.Destroy()
		})
		// Convey("with no content", WithApp(
		// 	TestingMakesNoContent,
		// 	func(d Application) {
		// 		// do tests here?
		// 		So(d.Display(), ShouldNotBeNil)
		// 	},
		// ))
	})
}
