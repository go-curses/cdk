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
	"time"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
	"github.com/go-curses/cdk/lib/sync"
)

const (
	TypeMetaData      CTypeTag = "cdk-metadata"
	SignalSetProperty Signal   = "set-property"
)

func init() {
	_ = TypesManager.AddType(TypeMetaData, nil)
}

type MetaData interface {
	Signaling

	Init() (already bool)
	InstallProperty(name Property, kind PropertyType, write bool, def interface{}) error
	InstallBuildableProperty(name Property, kind PropertyType, write bool, def interface{}) error
	OverloadProperty(name Property, kind PropertyType, write bool, buildable bool, def interface{}) error
	ListProperties() (properties []Property)
	ListBuildableProperties() (properties []Property)
	SetProperties(properties map[Property]string) (err error)
	IsProperty(name Property) bool
	IsBuildableProperty(name Property) (buildable bool)
	GetProperty(name Property) *CProperty
	SetPropertyFromString(name Property, value string) error
	SetProperty(name Property, value interface{}) error
	GetBoolProperty(name Property) (value bool, err error)
	SetBoolProperty(name Property, value bool) error
	GetStringProperty(name Property) (value string, err error)
	SetStringProperty(name Property, value string) error
	GetIntProperty(name Property) (value int, err error)
	SetIntProperty(name Property, value int) error
	GetFloat64Property(name Property) (value float64, err error)
	GetFloatProperty(name Property) (value float64, err error)
	SetFloatProperty(name Property, value float64) error
	GetColorProperty(name Property) (value paint.Color, err error)
	SetColorProperty(name Property, value paint.Color) error
	GetStyleProperty(name Property) (value paint.Style, err error)
	SetStyleProperty(name Property, value paint.Style) error
	GetThemeProperty(name Property) (value paint.Theme, err error)
	SetThemeProperty(name Property, value paint.Theme) error
	GetPointProperty(name Property) (value ptypes.Point2I, err error)
	SetPointProperty(name Property, value ptypes.Point2I) error
	GetRectangleProperty(name Property) (value ptypes.Rectangle, err error)
	SetRectangleProperty(name Property, value ptypes.Rectangle) error
	GetRegionProperty(name Property) (value ptypes.Region, err error)
	SetRegionProperty(name Property, value ptypes.Region) error
	GetStructProperty(name Property) (value interface{}, err error)
	SetStructProperty(name Property, value interface{}) error
	GetTimeProperty(name Property) (value time.Duration, err error)
	SetTimeProperty(name Property, value time.Duration) error
}

type CMetaData struct {
	CSignaling

	properties   []*CProperty
	propertyLock *sync.RWMutex
}

func (o *CMetaData) Init() (already bool) {
	if o.InitTypeItem(TypeMetaData, o) {
		return true
	}
	o.CSignaling.Init()
	o.properties = make([]*CProperty, 0)
	o.propertyLock = &sync.RWMutex{}
	return false
}

func (o *CMetaData) InstallProperty(name Property, kind PropertyType, write bool, def interface{}) error {
	existing := o.GetProperty(name)
	if existing != nil {
		return fmt.Errorf("property exists: %v", name)
	}
	o.propertyLock.Lock()
	o.properties = append(
		o.properties,
		NewProperty(name, kind, write, false, def),
	)
	o.propertyLock.Unlock()
	return nil
}

func (o *CMetaData) InstallBuildableProperty(name Property, kind PropertyType, write bool, def interface{}) error {
	existing := o.GetProperty(name)
	if existing != nil {
		return fmt.Errorf("property exists: %v", name)
	}
	o.propertyLock.Lock()
	o.properties = append(
		o.properties,
		NewProperty(name, kind, write, true, def),
	)
	o.propertyLock.Unlock()
	return nil
}

func (o *CMetaData) OverloadProperty(name Property, kind PropertyType, write bool, buildable bool, def interface{}) error {
	existing := o.GetProperty(name)
	if existing == nil {
		return fmt.Errorf("property not found: %v", name)
	}
	o.propertyLock.Lock()
	overload := Property(fmt.Sprintf("%v--overload", name))
	index := -1
	for idx, prop := range o.properties {
		if prop.name == overload {
			index = idx
			break
		}
	}
	if index == -1 {
		o.properties = append(
			o.properties,
			NewProperty(overload, kind, write, buildable, def),
		)
	} else {
		o.properties[index].kind = kind
		o.properties[index].write = write
		o.properties[index].def = def
	}
	o.propertyLock.Unlock()
	return nil
}

func (o *CMetaData) ListProperties() (properties []Property) {
	o.propertyLock.RLock()
	for _, prop := range o.properties {
		properties = append(properties, prop.Name())
	}
	o.propertyLock.RUnlock()
	return
}

func (o *CMetaData) ListBuildableProperties() (properties []Property) {
	o.propertyLock.RLock()
	for _, prop := range o.properties {
		if prop.Buildable() {
			properties = append(properties, prop.Name())
		}
	}
	o.propertyLock.RUnlock()
	return
}

func (o *CMetaData) SetProperties(properties map[Property]string) (err error) {
	for name, value := range properties {
		if prop := o.GetProperty(name); prop != nil {
			o.propertyLock.Lock()
			if prop.Buildable() {
				if err = prop.SetFromString(value); err != nil {
					o.LogError("error setting \"%v\" property from string: \"%v\" - %v", name, value, err)
				}
			} else {
				o.LogTrace("property not buildable: %v", name)
			}
			o.propertyLock.Unlock()
		} else {
			o.LogTrace("property not found: %v", name)
		}
	}
	return
}

func (o *CMetaData) IsProperty(name Property) bool {
	if prop := o.GetProperty(name); prop != nil {
		return true
	}
	return false
}

func (o *CMetaData) IsBuildableProperty(name Property) (buildable bool) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		buildable = prop.Buildable()
		o.propertyLock.RUnlock()
		return
	}
	return false
}

func (o *CMetaData) GetProperty(name Property) *CProperty {
	o.propertyLock.RLock()
	// check for overloaded properties first
	overload := Property(fmt.Sprintf("%v.overload", name))
	for _, prop := range o.properties {
		if prop.Name() == overload {
			o.propertyLock.RUnlock()
			return prop
		}
	}
	// check for regular named property
	for _, prop := range o.properties {
		if prop.Name() == name {
			o.propertyLock.RUnlock()
			return prop
		}
	}
	// o.LogError("property not found: %v", name)
	o.propertyLock.RUnlock()
	return nil
}

func (o *CMetaData) SetPropertyFromString(name Property, value string) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.ReadOnly() {
			return fmt.Errorf("error cannot update read-only property: %v", name)
		}
		if f := o.Emit(SignalSetProperty, o, name, value); f == enums.EVENT_PASS {
			o.propertyLock.Lock()
			if err := prop.SetFromString(value); err != nil {
				o.propertyLock.Unlock()
				return err
			}
			o.propertyLock.Unlock()
		}
	}
	return nil
}

func (o *CMetaData) SetProperty(name Property, value interface{}) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.ReadOnly() {
			return fmt.Errorf("error setting read-only property: %v", name)
		}
		if f := o.Emit(SignalSetProperty, o, name, value); f == enums.EVENT_PASS {
			o.propertyLock.Lock()
			if err := prop.Set(value); err != nil {
				o.propertyLock.Unlock()
				return err
			}
			o.propertyLock.Unlock()
		}
	}
	return nil
}

func (o *CMetaData) GetBoolProperty(name Property) (value bool, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == BoolProperty {
			if v, ok := prop.Value().(bool); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(bool); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return false, fmt.Errorf("%v.(%v) property is not a bool", name, prop.Type())
	}
	return false, fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetBoolProperty(name Property, value bool) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == BoolProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a bool", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetStringProperty(name Property) (value string, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == StringProperty {
			if v, ok := prop.Value().(string); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(string); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return "", fmt.Errorf("%v.(%v) property is not a string", name, prop.Type())
	}
	return "", fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetStringProperty(name Property, value string) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == StringProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a string", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetIntProperty(name Property) (value int, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == IntProperty {
			if v, ok := prop.Value().(int); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(int); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return 0, fmt.Errorf("%v.(%v) property is not an int", name, prop.Type())
	}
	return 0, fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetIntProperty(name Property, value int) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == IntProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not an int", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetFloat64Property(name Property) (value float64, err error) {
	return o.GetFloatProperty(name)
}

func (o *CMetaData) GetFloatProperty(name Property) (value float64, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == FloatProperty {
			if v, ok := prop.Value().(float64); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(float64); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return 0.0, fmt.Errorf("%v.(%v) property is not a float", name, prop.Type())
	}
	return 0.0, fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetFloatProperty(name Property, value float64) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == FloatProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a float64", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetColorProperty(name Property) (value paint.Color, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == ColorProperty {
			if v, ok := prop.Value().(paint.Color); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(paint.Color); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return paint.Color(0), fmt.Errorf("%v.(%v) property is not a Color", name, prop.Type())
	}
	return paint.Color(0), fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetColorProperty(name Property, value paint.Color) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == ColorProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a Color", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetStyleProperty(name Property) (value paint.Style, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == StyleProperty {
			if v, ok := prop.Value().(paint.Style); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(paint.Style); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return paint.Style{}, fmt.Errorf("%v.(%v) property is not a Style", name, prop.Type())
	}
	return paint.Style{}, fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetStyleProperty(name Property, value paint.Style) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == StyleProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a Style", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetThemeProperty(name Property) (value paint.Theme, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == ThemeProperty {
			if v, ok := prop.Value().(paint.Theme); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(paint.Theme); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return paint.Theme{}, fmt.Errorf("%v.(%v) property is not a Theme", name, prop.Type())
	}
	return paint.Theme{}, fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetThemeProperty(name Property, value paint.Theme) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == ThemeProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a Theme", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetPointProperty(name Property) (value ptypes.Point2I, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == PointProperty {
			if v, ok := prop.Value().(ptypes.Point2I); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(ptypes.Point2I); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return ptypes.Point2I{}, fmt.Errorf("%v.(%v) property is not a Point2I", name, prop.Type())
	}
	return ptypes.Point2I{}, fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetPointProperty(name Property, value ptypes.Point2I) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == PointProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a Point2I", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetRectangleProperty(name Property) (value ptypes.Rectangle, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == RectangleProperty {
			if v, ok := prop.Value().(ptypes.Rectangle); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(ptypes.Rectangle); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return ptypes.Rectangle{}, fmt.Errorf("%v.(%v) property is not a Rectangle", name, prop.Type())
	}
	return ptypes.Rectangle{}, fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetRectangleProperty(name Property, value ptypes.Rectangle) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == RectangleProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a Rectangle", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetRegionProperty(name Property) (value ptypes.Region, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == RegionProperty {
			if v, ok := prop.Value().(ptypes.Region); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(ptypes.Region); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return ptypes.Region{}, fmt.Errorf("%v.(%v) property is not a Region", name, prop.Type())
	}
	return ptypes.Region{}, fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetRegionProperty(name Property, value ptypes.Region) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == RegionProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a Region", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetStructProperty(name Property) (value interface{}, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == StructProperty {
			if v := prop.Value(); v != nil {
				o.propertyLock.RUnlock()
				return v, nil
			}
			o.propertyLock.RUnlock()
			return prop.Default(), nil
		}
		o.propertyLock.RUnlock()
		return 0, fmt.Errorf("%v.(%v) property is not a struct", name, prop.Type())
	}
	return 0, fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetStructProperty(name Property, value interface{}) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == StructProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a struct", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) GetTimeProperty(name Property) (value time.Duration, err error) {
	if prop := o.GetProperty(name); prop != nil {
		o.propertyLock.RLock()
		if prop.Type() == TimeProperty {
			if v, ok := prop.Value().(time.Duration); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
			if v, ok := prop.Default().(time.Duration); ok {
				o.propertyLock.RUnlock()
				return v, nil
			}
		}
		o.propertyLock.RUnlock()
		return 0, fmt.Errorf("%v.(%v) property is not a Time", name, prop.Type())
	}
	return 0, fmt.Errorf("property not found: %v", name)
}

func (o *CMetaData) SetTimeProperty(name Property, value time.Duration) error {
	if prop := o.GetProperty(name); prop != nil {
		if prop.Type() == TimeProperty {
			return o.SetProperty(name, value)
		}
		return fmt.Errorf("%v.(%v) property is not a Time", name, prop.Type())
	}
	return fmt.Errorf("property not found: %v", name)
}
