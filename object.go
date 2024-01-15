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
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
)

const TypeObject CTypeTag = "cdk-object"

func init() {
	_ = TypesManager.AddType(TypeObject, nil)
}

// This is the base type for all complex CDK object types. The Object type
// provides a means of installing properties, getting and setting property
// values
type Object interface {
	MetaData

	Init() (already bool)
	InitWithProperties(properties map[Property]string) (already bool, err error)
	Destroy()
	GetName() (name string)
	SetName(name string)
	GetTheme() (theme paint.Theme)
	SetTheme(theme paint.Theme)
}

type CObject struct {
	CMetaData
}

func (o *CObject) Init() (already bool) {
	if o.InitTypeItem(TypeObject, o) {
		return true
	}
	o.CMetaData.Init()
	o.properties = make([]*CProperty, 0)
	_ = o.InstallProperty(PropertyDebug, BoolProperty, true, false)
	_ = o.InstallProperty(PropertyName, StringProperty, true, "")
	_ = o.InstallProperty(PropertyTheme, ThemeProperty, true, paint.GetDefaultColorTheme())
	return false
}

func (o *CObject) InitWithProperties(properties map[Property]string) (already bool, err error) {
	if o.Init() {
		return true, nil
	}
	if err = o.SetProperties(properties); err != nil {
		return false, err
	}
	return false, nil
}

func (o *CObject) Destroy() {
	if f := o.Emit(SignalDestroy, o); f == enums.EVENT_PASS {
		if err := o.DestroyObject(); err != nil {
			o.LogErr(err)
		}
	}
}

func (o *CObject) GetName() (name string) {
	var err error
	if name, err = o.GetStringProperty(PropertyName); err != nil {
		return ""
	}
	if name == "" {
		name = o.CTypeItem.GetName()
	}
	return
}

func (o *CObject) SetName(name string) {
	if err := o.SetStringProperty(PropertyName, name); err != nil {
		o.LogErr(err)
	} else {
		o.CTypeItem.SetName(name)
	}
}

func (o *CObject) GetTheme() (theme paint.Theme) {
	var err error
	if theme, err = o.GetThemeProperty(PropertyTheme); err != nil {
		o.LogErr(err)
	}
	return
}

func (o *CObject) SetTheme(theme paint.Theme) {
	if err := o.SetThemeProperty(PropertyTheme, theme); err != nil {
		o.LogErr(err)
	}
}

// emitted when the object instance is destroyed
const SignalDestroy Signal = "destroy"

// request that the object be rendered with additional features useful to
// debugging custom Widget development
const PropertyDebug Property = "debug"

// property wrapper around the CTypeItem name field
const PropertyName Property = "name"

const PropertyTheme Property = "theme"

const ObjectSetPropertyHandle = "object-set-property-handle"
