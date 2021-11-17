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
	"sort"
	"sync"

	"github.com/go-curses/cdk/log"
	"github.com/gofrs/uuid"
)

var (
	TypesManager = NewTypeRegistry()
)

type TypeRegistry interface {
	GetTypeTags() (tags []TypeTag)
	GetBuildableInfo() (info map[string]TypeTag)
	MakeType(tag TypeTag) (thing interface{}, err error)
	AddType(tag TypeTag, constructor func() interface{}, aliases ...string) error
	HasType(tag TypeTag) (exists bool)
	GetType(tag TypeTag) (t Type, ok bool)
	AddTypeAlias(tag TypeTag, alias ...string)
	GetTypeTagByAlias(alias string) (tt TypeTag, ok bool)
	AddTypeItem(tag TypeTag, item interface{}) (id uuid.UUID, err error)
	HasID(index uuid.UUID) bool
	GetTypeItems(tag TypeTag) []interface{}
	GetTypeItemByID(id uuid.UUID) interface{}
	GetTypeItemByName(name string) interface{}
	RemoveTypeItem(tag TypeTag, item TypeItem) error
}

type CTypeRegistry struct {
	register     map[TypeTag]Type
	aliases      map[string]TypeTag
	registryLock *sync.RWMutex
}

func NewTypeRegistry() TypeRegistry {
	return &CTypeRegistry{
		register:     make(map[TypeTag]Type),
		aliases:      make(map[string]TypeTag),
		registryLock: &sync.RWMutex{},
	}
}

func (r *CTypeRegistry) GetTypeTags() (tags []TypeTag) {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	for tt, _ := range r.register {
		tags = append(tags, tt)
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].String() < tags[j].String()
	})
	return
}

func (r *CTypeRegistry) GetBuildableInfo() (info map[string]TypeTag) {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	var tmp []TypeTag
	for tt, tType := range r.register {
		if tType.Buildable() {
			tmp = append(tmp, tt)
		}
	}
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].String() < tmp[j].String()
	})
	info = make(map[string]TypeTag)
	for _, tt := range tmp {
		gt := tt.GladeString()
		info[gt] = tt
	}
	for alias, tt := range r.aliases {
		gt := CTypeTag(alias).GladeString()
		info[gt] = tt
		info[alias] = tt
	}
	return
}

func (r *CTypeRegistry) MakeType(tag TypeTag) (thing interface{}, err error) {
	if t, ok := r.GetType(tag); ok {
		r.registryLock.Lock()
		defer r.registryLock.Unlock()
		if t.Buildable() {
			if thing = t.New(); thing == nil {
				err = fmt.Errorf("buildable produced nil: %v", tag)
			}
		}
	} else {
		err = fmt.Errorf("type not found: %v", tag)
	}
	return
}

func (r *CTypeRegistry) AddType(tag TypeTag, constructor func() interface{}, aliases ...string) error {
	r.registryLock.RLock()
	if tag == TypeNil {
		r.registryLock.RUnlock()
		return fmt.Errorf("cannot add nil type")
	}
	if _, ok := r.register[tag]; ok {
		r.registryLock.RUnlock()
		return fmt.Errorf("type %v exists already", tag)
	}
	r.registryLock.RUnlock()
	r.registryLock.Lock()
	r.register[tag] = NewType(tag, constructor)
	r.registryLock.Unlock()
	r.AddTypeAlias(tag, aliases...)
	return nil
}

func (r *CTypeRegistry) HasType(tag TypeTag) (exists bool) {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	_, exists = r.register[tag]
	return
}

func (r *CTypeRegistry) GetType(tag TypeTag) (t Type, ok bool) {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	t, ok = r.register[tag]
	return
}

func (r *CTypeRegistry) AddTypeAlias(tag TypeTag, aliases ...string) {
	for _, alias := range aliases {
		if r.HasType(CTypeTag(alias)) {
			log.ErrorF("error, invalid alias: %v (concrete type)", alias)
			continue
		}
		r.registryLock.Lock()
		if t, ok := r.aliases[alias]; ok {
			log.WarnF("overwriting alias %v - was: %v, now: %v", alias, t.Tag(), tag)
		}
		r.aliases[alias] = tag
		r.registryLock.Unlock()
	}
}

func (r *CTypeRegistry) GetTypeTagByAlias(alias string) (tt TypeTag, ok bool) {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	for a, t := range r.aliases {
		if alias == a {
			return t, true
		}
	}
	return
}

func (r *CTypeRegistry) AddTypeItem(tag TypeTag, item interface{}) (id uuid.UUID, err error) {
	r.registryLock.RLock()
	if tag == TypeNil {
		r.registryLock.RUnlock()
		id, err = uuid.Nil, fmt.Errorf("cannot add to nil type")
		return
	}
	if _, ok := r.register[tag]; !ok {
		r.registryLock.RUnlock()
		id, err = uuid.Nil, fmt.Errorf("unknown type: %v", tag)
		return
	}
	r.registryLock.RUnlock()
	r.registryLock.Lock()
	r.register[tag].Add(item)
	id, _ = uuid.NewV4()
	r.registryLock.Unlock()
	return
}

func (r *CTypeRegistry) HasID(index uuid.UUID) bool {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	for _, t := range r.register {
		for _, item := range t.Items() {
			if ci, ok := item.(TypeItem); ok {
				if index == ci.ObjectID() {
					return true
				}
			}
		}
	}
	return false
}

func (r *CTypeRegistry) GetTypeItems(tag TypeTag) []interface{} {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	if t, ok := r.register[tag]; ok {
		return t.Items()
	}
	return nil
}

func (r *CTypeRegistry) GetTypeItemByID(id uuid.UUID) interface{} {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	for _, t := range r.register {
		for _, i := range t.Items() {
			if c, ok := i.(TypeItem); ok {
				if c.ObjectID() == id {
					return i
				}
			}
		}
	}
	return nil
}

func (r *CTypeRegistry) GetTypeItemByName(name string) interface{} {
	r.registryLock.RLock()
	defer r.registryLock.RUnlock()
	for _, t := range r.register {
		for _, i := range t.Items() {
			if c, ok := i.(TypeItem); ok {
				if c.GetName() == name {
					return i
				}
			}
		}
	}
	return nil
}

func (r *CTypeRegistry) RemoveTypeItem(tag TypeTag, item TypeItem) error {
	r.registryLock.RLock()
	if item == nil || !item.IsValid() {
		r.registryLock.RUnlock()
		return fmt.Errorf("item is nil or not valid")
	}
	if _, ok := r.register[tag]; !ok {
		r.registryLock.RUnlock()
		return fmt.Errorf("unknown type: %v", tag)
	}
	r.registryLock.RUnlock()
	r.registryLock.Lock()
	if err := r.register[tag].Remove(item); err != nil {
		r.registryLock.Unlock()
		return err
	}
	r.registryLock.Unlock()
	return nil
}
