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
	"strconv"
	"strings"
	"sync"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
)

type Property string

func (p Property) String() string {
	return string(p)
}

type CProperty struct {
	name      Property
	kind      PropertyType
	write     bool
	buildable bool
	def       interface{}
	value     interface{}

	sync.RWMutex
}

func NewProperty(name Property, kind PropertyType, write bool, buildable bool, def interface{}) (property *CProperty) {
	property = new(CProperty)
	property.name = name
	property.kind = kind
	property.write = write
	property.buildable = buildable
	property.def = def
	property.value = def
	return
}

func (p *CProperty) Clone() *CProperty {
	return &CProperty{
		name:      p.name,
		kind:      p.kind,
		write:     p.write,
		buildable: p.buildable,
		def:       p.def,
		value:     p.value,
	}
}

func (p *CProperty) Name() Property {
	p.RLock()
	defer p.RUnlock()
	return p.name
}

func (p *CProperty) Type() PropertyType {
	p.RLock()
	defer p.RUnlock()
	return p.kind
}

func (p *CProperty) ReadOnly() bool {
	p.RLock()
	defer p.RUnlock()
	return !p.write
}

func (p *CProperty) Buildable() bool {
	p.RLock()
	defer p.RUnlock()
	return p.buildable
}

func (p *CProperty) Set(value interface{}) error {
	if p.ReadOnly() {
		return fmt.Errorf("cannot change read-only property: %v", p.name)
	}
	t := p.Type()
	p.Lock()
	defer p.Unlock()
	switch t {
	case BoolProperty:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%v value is not of bool type: %v (%T)", p.name, value, value)
		}
	case StringProperty:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%v value is not of string type: %v (%T)", p.name, value, value)
		}
	case IntProperty:
		if _, ok := value.(int); !ok {
			return fmt.Errorf("%v value is not of int type: %v (%T)", p.name, value, value)
		}
	case FloatProperty:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%v value is not of float64 type: %v (%T)", p.name, value, value)
		}
	case ColorProperty:
		if _, ok := value.(paint.Color); !ok {
			return fmt.Errorf("%v value is not of cdk.Color type: %v (%T)", p.name, value, value)
		}
	case ThemeProperty:
		if _, ok := value.(paint.Theme); !ok {
			return fmt.Errorf("%v value is not of cdk.Theme type: %v (%T)", p.name, value, value)
		}
	case PointProperty:
		if _, ok := value.(ptypes.Point2I); !ok {
			return fmt.Errorf("%v value is not of cdk.Point2I type: %v (%T)", p.name, value, value)
		}
	case RectangleProperty:
		if _, ok := value.(ptypes.Rectangle); !ok {
			return fmt.Errorf("%v value is not of cdk.Rectangle type: %v (%T)", p.name, value, value)
		}
	case RegionProperty:
		if _, ok := value.(ptypes.Region); !ok {
			return fmt.Errorf("%v value is not of cdk.Region type: %v (%T)", p.name, value, value)
		}
	case StructProperty:
		// no checks, just pass
	default:
		return fmt.Errorf("invalid property type for %v: %v", p.name, p.Type())
	}
	p.value = value
	return nil
}

func (p *CProperty) SetFromString(value string) error {
	if p.ReadOnly() {
		return fmt.Errorf("cannot change read-only property: %v", p.name)
	}
	switch p.Type() {
	case BoolProperty:
		switch strings.ToLower(value) {
		case "true", "t", "1":
			return p.Set(true)
		}
		return p.Set(false)
	case StringProperty:
		return p.Set(value)
	case IntProperty:
		if index := strings.Index(value, "px"); index > -1 {
			value = value[:index-1]
		} else if index := strings.Index(value, "pt"); index > -1 {
			value = value[:index-1]
		} else if index := strings.Index(value, "%"); index > -1 {
			value = value[:index-1]
		}
		if v, err := strconv.Atoi(value); err != nil {
			return err
		} else {
			return p.Set(v)
		}
	case FloatProperty:
		if v, err := strconv.ParseFloat(value, 64); err != nil {
			return err
		} else {
			return p.Set(v)
		}
	case ColorProperty:
		if c, ok := paint.ParseColor(value); ok {
			return p.Set(c)
		} else {
			return fmt.Errorf("invalid color value: %v", value)
		}
	case StyleProperty:
		if c, err := paint.ParseStyle(value); err != nil {
			return err
		} else {
			return p.Set(c)
		}
	case ThemeProperty:
		return fmt.Errorf("theme property not supported by builder features")
	case PointProperty:
		if v, ok := ptypes.ParsePoint2I(value); ok {
			return p.Set(v)
		} else {
			return fmt.Errorf("invalid point value: %v", value)
		}
	case RectangleProperty:
		if v, ok := ptypes.ParseRectangle(value); ok {
			return p.Set(v)
		} else {
			return fmt.Errorf("invalid rectangle value: %v", value)
		}
	case RegionProperty:
		if v, ok := ptypes.ParseRegion(value); ok {
			return p.Set(v)
		} else {
			return fmt.Errorf("invalid region value: %v", value)
		}
	case StructProperty:
		if efs, ok := p.Default().(enums.EnumFromString); ok {
			if nv, err := efs.FromString(value); err != nil {
				return err
			} else {
				return p.Set(nv)
			}
		}
		return fmt.Errorf("complex property %v not supported by builder features", p.Name())
	}
	return fmt.Errorf("error")
}

func (p *CProperty) Default() (def interface{}) {
	p.RLock()
	defer p.RUnlock()
	def = p.def
	return
}

func (p *CProperty) Value() (value interface{}) {
	p.RLock()
	defer p.RUnlock()
	if p.value == nil {
		value = p.def
	} else {
		value = p.value
	}
	return
}
