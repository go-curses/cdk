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
	sync.Locker

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
}

type CTypeItem struct {
	sync.Mutex

	id       uuid.UUID
	typeTag  CTypeTag
	name     string
	ancestry []TypeTag
	valid    bool
	self     interface{}
}

func NewTypeItem(tag CTypeTag, name string) TypeItem {
	return &CTypeItem{
		id:       uuid.Nil,
		typeTag:  tag,
		name:     name,
		ancestry: make([]TypeTag, 0),
		valid:    false,
		self:     nil,
	}
}

func (o *CTypeItem) InitTypeItem(tag TypeTag, thing interface{}) (already bool) {
	o.Lock()
	defer o.Unlock()
	already = o.valid
	if !already {
		if o.typeTag == TypeNil {
			o.typeTag = tag.Tag()
		}
		if o.id == uuid.Nil {
			var err error
			if o.id, err = TypesManager.AddTypeItem(o.typeTag, thing); err != nil {
				log.FatalDF(1, "failed to add self to \"%v\" type: %v", o.typeTag, err)
			} else {
				o.self = thing
			}
		} else {
			o.ancestry = append(o.ancestry, tag)
		}
	}
	return
}

func (o *CTypeItem) Init() (already bool) {
	if o.valid {
		return true
	}
	if o.typeTag == TypeNil {
		log.FatalDF(1, "invalid object type: nil")
	}
	found := false
	if TypesManager.HasType(o.typeTag) {
		for _, i := range TypesManager.GetTypeItems(o.typeTag) {
			if c, ok := i.(TypeItem); ok {
				if c.ObjectID() == o.ObjectID() {
					found = true
					break
				}
			}
		}
	}
	if !found {
		log.FatalDF(1, "type or instance not found: %v (%v)", o.ObjectName(), o.typeTag)
	}
	o.valid = true
	return false
}

func (o *CTypeItem) IsValid() bool {
	return o.valid && o.self != nil
}

func (o *CTypeItem) Self() (this interface{}) {
	return o.self
}

func (o *CTypeItem) String() string {
	return o.ObjectName()
}

func (o *CTypeItem) GetTypeTag() TypeTag {
	return o.typeTag
}

func (o *CTypeItem) GetName() string {
	return o.name
}

func (o *CTypeItem) SetName(name string) {
	o.Lock()
	defer o.Unlock()
	o.name = name
}

func (o *CTypeItem) ObjectID() uuid.UUID {
	return o.id
}

func (o *CTypeItem) ObjectName() string {
	if len(o.name) > 0 {
		return fmt.Sprintf("%v-%v#%v", o.typeTag, o.ObjectID(), o.name)
	}
	return fmt.Sprintf("%v-%v", o.typeTag, o.ObjectID())
}

func (o *CTypeItem) DestroyObject() (err error) {
	if err = TypesManager.RemoveTypeItem(o.typeTag, o); err != nil {
		o.LogErr(err)
	}
	o.valid = false
	o.id = uuid.Nil
	return nil
}

func (o *CTypeItem) LogTag() string {
	if len(o.name) > 0 {
		return fmt.Sprintf("[%v.%v.%v]", o.typeTag, o.ObjectID(), o.name)
	}
	return fmt.Sprintf("[%v.%v]", o.typeTag, o.ObjectID())
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
