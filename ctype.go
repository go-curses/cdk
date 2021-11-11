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
	"sync"
)

const (
	TypeNil CTypeTag = ""
)

type Type interface {
	New() interface{}
	Buildable() (hasConstructor bool)
	Items() []interface{}
	Add(item interface{})
	Remove(item TypeItem) error
	Aliases() []string
}

type CType struct {
	tag   TypeTag
	items []interface{}
	new   func() interface{}
	alias []string

	sync.Mutex
}

func NewType(tag TypeTag, constructor func() interface{}, aliases ...string) Type {
	return &CType{
		tag:   tag,
		items: make([]interface{}, 0),
		new:   constructor,
		alias: aliases,
	}
}

func (t *CType) New() interface{} {
	if t.new != nil {
		return t.new()
	}
	return nil
}

func (t *CType) Aliases() (aliases []string) {

	return
}

func (t *CType) Buildable() (hasConstructor bool) {
	return t.new != nil
}

func (t *CType) Items() []interface{} {
	t.Lock()
	defer t.Unlock()
	return t.items
}

func (t *CType) Add(item interface{}) {
	t.Lock()
	defer t.Unlock()
	t.items = append(t.items, item)
	return
}

func (t *CType) Remove(item TypeItem) error {
	t.Lock()
	defer t.Unlock()
	var idx int
	var itm interface{}
	for idx, itm = range t.items {
		if tt, ok := itm.(TypeItem); ok {
			if tt.ObjectID() == item.ObjectID() {
				break
			}
		}
	}
	count := len(t.items)
	if count > 0 && idx >= count {
		return fmt.Errorf("item not found")
	} else if count > 1 {
		t.items = append(
			t.items[:idx],
			t.items[idx+1:]...,
		)
	} else {
		t.items = []interface{}{}
	}
	return nil
}
