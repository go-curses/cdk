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

	"github.com/go-curses/cdk/log"
	"github.com/gofrs/uuid"
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
	typeLock *sync.RWMutex

	sync.RWMutex
}

func NewTypeItem(tag CTypeTag, name string) TypeItem {
	return &CTypeItem{
		id:       uuid.Nil,
		typeTag:  tag,
		name:     name,
		ancestry: make([]TypeTag, 0),
		valid:    false,
		self:     nil,
		typeLock: &sync.RWMutex{},
	}
}

func (o *CTypeItem) InitTypeItem(tag TypeTag, thing interface{}) (already bool) {
	if o.typeLock == nil {
		o.typeLock = &sync.RWMutex{}
	}
	o.typeLock.RLock()
	already = o.valid
	if !already {
		if o.typeTag == TypeNil {
			o.typeTag = tag.Tag()
		}
		if o.id == uuid.Nil {
			o.typeLock.RUnlock()
			o.typeLock.Lock()
			var err error
			if o.id, err = TypesManager.AddTypeItem(o.typeTag, thing); err != nil {
				log.FatalDF(1, "failed to add self to \"%v\" type: %v", o.typeTag, err)
			} else {
				o.self = thing
			}
			o.typeLock.Unlock()
		} else {
			o.typeLock.RUnlock()
			o.typeLock.Lock()
			o.ancestry = append(o.ancestry, tag)
			o.typeLock.Unlock()
		}
	} else {
		o.typeLock.RUnlock()
	}
	return
}

func (o *CTypeItem) Init() (already bool) {
	if o.IsValid() {
		return true
	}
	o.typeLock.RLock()
	if o.typeTag == TypeNil {
		log.FatalDF(1, "invalid object type: nil")
	}
	found := false
	if TypesManager.HasType(o.typeTag) {
		for _, i := range TypesManager.GetTypeItems(o.typeTag) {
			if c, ok := i.(TypeItem); ok {
				o.typeLock.RUnlock()
				if c.ObjectID() == o.ObjectID() {
					o.typeLock.RLock()
					found = true
					break
				}
				o.typeLock.RLock()
			}
		}
	}
	if !found {
		log.FatalDF(1, "type or instance not found: %v (%v)", o.ObjectName(), o.typeTag)
	}
	o.typeLock.RUnlock()
	o.typeLock.Lock()
	defer o.typeLock.Unlock()
	o.valid = true
	return false
}

func (o *CTypeItem) IsValid() bool {
	o.typeLock.RLock()
	defer o.typeLock.RUnlock()
	return o.valid && o.self != nil
}

func (o *CTypeItem) Self() (this interface{}) {
	o.typeLock.RLock()
	defer o.typeLock.RUnlock()
	return o.self
}

func (o *CTypeItem) String() string {
	return o.ObjectName()
}

func (o *CTypeItem) GetTypeTag() TypeTag {
	o.typeLock.RLock()
	defer o.typeLock.RUnlock()
	return o.typeTag
}

func (o *CTypeItem) GetName() string {
	o.typeLock.RLock()
	defer o.typeLock.RUnlock()
	return o.name
}

func (o *CTypeItem) SetName(name string) {
	o.typeLock.Lock()
	defer o.typeLock.Unlock()
	o.name = name
}

func (o *CTypeItem) ObjectID() uuid.UUID {
	o.typeLock.RLock()
	defer o.typeLock.RUnlock()
	return o.id
}

func (o *CTypeItem) ObjectName() string {
	o.typeLock.RLock()
	tt, n := o.typeTag, o.name
	o.typeLock.RUnlock()
	if len(o.name) > 0 {
		return fmt.Sprintf("%v-%v#%v", tt, o.ObjectID(), n)
	}
	return fmt.Sprintf("%v-%v", n, o.ObjectID())
}

func (o *CTypeItem) DestroyObject() (err error) {
	o.RLock()
	if err = TypesManager.RemoveTypeItem(o.typeTag, o); err != nil {
		o.LogErr(err)
	}
	o.RUnlock()
	o.typeLock.Lock()
	defer o.typeLock.Unlock()
	o.valid = false
	o.id = uuid.Nil
	return nil
}

func (o *CTypeItem) LogTag() string {
	o.typeLock.RLock()
	tt, n := o.typeTag, o.name
	o.typeLock.RUnlock()
	if len(o.name) > 0 {
		return fmt.Sprintf("[%v.%v.%v]", tt, o.ObjectID(), n)
	}
	return fmt.Sprintf("[%v.%v]", tt, o.ObjectID())
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
