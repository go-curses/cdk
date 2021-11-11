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
)

const (
	TypeTest = CTypeTag("test")
)

func init() {
	TypesManager.AddType(TypeTest, func() interface{} { return nil })
}

func TestCTypeRegistry(t *testing.T) {
	Convey("All Basic CType Features", t, func() {
		So(TypesManager.HasType(TypeTest), ShouldEqual, true)
		items := TypesManager.GetTypeItems(TypeTest)
		So(items, ShouldHaveLength, 0)
		testType, found := TypesManager.GetType(TypeTest)
		So(found, ShouldEqual, true)
		So(testType.Items(), ShouldHaveLength, 0)
		// firstItem := NewTypeItem(TypeTest, "first-test")
		// So(firstItem.Init(), ShouldEqual, false)
		// So(firstItem.Init(), ShouldEqual, true)
		// So(testType.Items(), ShouldHaveLength, 1)
		// So(testType.Items()[0], ShouldNotBeNil)
		// So(testType.Items()[0], ShouldEqual, firstItem)
		// So(firstItem.GetTypeTag(), ShouldEqual, TypeTest)
		// So(firstItem.GetName(), ShouldEqual, "first-test")
		// firstItem.SetName("first-test-renamed")
		// So(firstItem.GetName(), ShouldEqual, "first-test-renamed")
		// failItem := NewTypeItem(TypeTest, "")
		// So(TypesManager.RemoveTypeItem(TypeTest, failItem), ShouldNotBeNil)
		// So(TypesManager.RemoveTypeItem(TypeTest, firstItem), ShouldBeNil)
		// var err error
		// err = TypesManager.AddType(TypeNil)
		// So(err, ShouldNotBeNil)
		// So(err.Error(), ShouldEqual, "cannot add nil type")
		// err = TypesManager.AddType(TypeTest)
		// So(err, ShouldNotBeNil)
		// So(err.Error(), ShouldEqual, "type test exists already")
		// var id int
		// id, err = TypesManager.AddTypeItem(TypeNil, nil)
		// So(err, ShouldNotBeNil)
		// So(err.Error(), ShouldEqual, "cannot add to nil type")
		// So(id, ShouldEqual, -1)
		// nopeType := CTypeTag("nope")
		// id, err = TypesManager.AddTypeItem(nopeType, nil)
		// So(err, ShouldNotBeNil)
		// So(err.Error(), ShouldEqual, "unknown type: nope")
		// So(id, ShouldEqual, -1)
		// err = TypesManager.RemoveTypeItem(nopeType, firstItem)
		// So(err, ShouldNotBeNil)
		// So(err.Error(), ShouldEqual, "unknown type: nope")
		// err = TypesManager.RemoveTypeItem(TypeTest, nil)
		// So(err, ShouldNotBeNil)
		// So(err.Error(), ShouldEqual, "item not valid")
		// err = TypesManager.RemoveTypeItem(TypeTest, firstItem)
		// So(err, ShouldNotBeNil)
		// So(err.Error(), ShouldEqual, "item not found")
	})
}
