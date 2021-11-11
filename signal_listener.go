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
	"github.com/go-curses/cdk/lib/enums"
)

type SignalListenerFn func(data []interface{}, argv ...interface{}) enums.EventFlag

type SignalListenerData []interface{}

type CSignalListener struct {
	s Signal
	n string
	c SignalListenerFn
	d SignalListenerData
}

func newSignalListener(s Signal, n string, c SignalListenerFn, d SignalListenerData) *CSignalListener {
	return &CSignalListener{
		s: s,
		n: n,
		c: c,
		d: d,
	}
}
