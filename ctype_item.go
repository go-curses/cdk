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

	"github.com/gofrs/uuid"

	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
)

type TypeItem interface {
	InitTypeItem(tag TypeTag, thing interface{}) (already bool)
	Init() (already bool)
	IsValid() bool
	Self() (this interface{})
	String() string
	GetTypeTag() TypeTag
	GetName() string
	SetName(name string)
	ObjectID() uuid.UUID
	ObjectName() string
	DestroyObject() (err error)
	LogTag() string
	LogTrace(format string, argv ...interface{})
	LogDebug(format string, argv ...interface{})
	LogInfo(format string, argv ...interface{})
	LogWarn(format string, argv ...interface{})
	LogError(format string, argv ...interface{})
	LogErr(err error)

	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type CTypeItem struct {
	id       uuid.UUID
	typeTag  CTypeTag
	name     string
	ancestry []TypeTag
	valid    bool
	self     interface{}

	itemLock  *sync.RWMutex
	lockStack []string
	sync.RWMutex
}

func NewTypeItem(tag CTypeTag, name string) TypeItem {
	return &CTypeItem{
		id:        uuid.Nil,
		typeTag:   tag,
		name:      name,
		ancestry:  make([]TypeTag, 0),
		valid:     false,
		self:      nil,
		itemLock:  &sync.RWMutex{},
		lockStack: make([]string, 0),
	}
}

func (o *CTypeItem) InitTypeItem(tag TypeTag, thing interface{}) (already bool) {
	if o.itemLock == nil {
		o.itemLock = &sync.RWMutex{}
	}
	o.itemLock.RLock()
	already = o.valid
	if !already {
		if o.typeTag == TypeNil {
			o.typeTag = tag.Tag()
		}
		if o.id == uuid.Nil {
			o.itemLock.RUnlock()
			o.itemLock.Lock()
			var err error
			if o.id, err = TypesManager.AddTypeItem(o.typeTag, thing); err != nil {
				log.FatalDF(1, "failed to add self to \"%v\" type: %v", o.typeTag, err)
			} else {
				o.self = thing
			}
			o.itemLock.Unlock()
		} else {
			o.itemLock.RUnlock()
			o.itemLock.Lock()
			o.ancestry = append(o.ancestry, tag)
			o.itemLock.Unlock()
		}
	} else {
		o.itemLock.RUnlock()
	}
	return
}

func (o *CTypeItem) Init() (already bool) {
	if o.IsValid() {
		return true
	}
	o.itemLock.RLock()
	if o.typeTag == TypeNil {
		log.FatalDF(1, "invalid object type: nil")
	}
	found := false
	if TypesManager.HasType(o.typeTag) {
		for _, i := range TypesManager.GetTypeItems(o.typeTag) {
			if c, ok := i.(TypeItem); ok {
				o.itemLock.RUnlock()
				if c.ObjectID() == o.ObjectID() {
					o.itemLock.RLock()
					found = true
					break
				}
				o.itemLock.RLock()
			}
		}
	}
	if !found {
		log.FatalDF(1, "type or instance not found: %v (%v)", o.ObjectName(), o.typeTag)
	}
	o.itemLock.RUnlock()
	o.itemLock.Lock()
	defer o.itemLock.Unlock()
	o.valid = true
	return false
}

func (o *CTypeItem) IsValid() bool {
	o.itemLock.RLock()
	defer o.itemLock.RUnlock()
	return o.valid && o.self != nil
}

func (o *CTypeItem) Self() (this interface{}) {
	o.itemLock.RLock()
	defer o.itemLock.RUnlock()
	return o.self
}

func (o *CTypeItem) String() string {
	return o.ObjectName()
}

func (o *CTypeItem) GetTypeTag() TypeTag {
	o.itemLock.RLock()
	defer o.itemLock.RUnlock()
	return o.typeTag
}

func (o *CTypeItem) GetName() string {
	o.itemLock.RLock()
	defer o.itemLock.RUnlock()
	return o.name
}

func (o *CTypeItem) SetName(name string) {
	o.itemLock.Lock()
	defer o.itemLock.Unlock()
	o.name = name
}

func (o *CTypeItem) ObjectID() uuid.UUID {
	o.itemLock.RLock()
	defer o.itemLock.RUnlock()
	return o.id
}

func (o *CTypeItem) ObjectName() string {
	o.itemLock.RLock()
	tt, n := o.typeTag, o.name
	o.itemLock.RUnlock()
	if len(n) > 0 {
		return fmt.Sprintf("%v.%v#%v", tt, o.ObjectID(), n)
	}
	return fmt.Sprintf("%v.%v", tt, o.ObjectID())
}

func (o *CTypeItem) DestroyObject() (err error) {
	o.RLock()
	if err = TypesManager.RemoveTypeItem(o.typeTag, o); err != nil {
		o.LogErr(err)
	}
	o.RUnlock()
	o.itemLock.Lock()
	defer o.itemLock.Unlock()
	o.valid = false
	o.id = uuid.Nil
	return nil
}

func (o *CTypeItem) LogTag() string {
	return fmt.Sprintf("[%v]", o.ObjectName())
}

func (o *CTypeItem) LogTrace(format string, argv ...interface{}) {
	log.TraceDF(1, fmt.Sprintf("%s %s", o.LogTag(), format), argv...)
}

func (o *CTypeItem) LogDebug(format string, argv ...interface{}) {
	log.DebugDF(1, fmt.Sprintf("%s %s", o.LogTag(), format), argv...)
}

func (o *CTypeItem) LogInfo(format string, argv ...interface{}) {
	log.InfoDF(1, fmt.Sprintf("%s %s", o.LogTag(), format), argv...)
}

func (o *CTypeItem) LogWarn(format string, argv ...interface{}) {
	log.WarnDF(1, fmt.Sprintf("%s %s", o.LogTag(), format), argv...)
}

func (o *CTypeItem) LogError(format string, argv ...interface{}) {
	log.ErrorDF(1, fmt.Sprintf("%s %s", o.LogTag(), format), argv...)
}

func (o *CTypeItem) LogErr(err error) {
	log.ErrorDF(1, err.Error())
}
