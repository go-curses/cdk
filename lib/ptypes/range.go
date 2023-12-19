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

package ptypes

// Range

import (
	"fmt"
)

type Range struct {
	Start, End int
}

func NewRange(start, end int) *Range {
	r := MakeRange(start, end)
	return &r
}

func MakeRange(start, end int) Range {
	return Range{Start: start, End: end}
}

func (r Range) String() string {
	return fmt.Sprintf("{start:%v,end:%v}", r.Start, r.End)
}

// Clone returns a new Range structure with the same values as this structure
func (r Range) Clone() (clone Range) {
	clone.Start = r.Start
	clone.End = r.End
	return
}

// NewClone returns a new Range instance with the same values as this structure
func (r Range) NewClone() (clone *Range) {
	clone = NewRange(r.Start, r.End)
	return
}

func (r Range) InRange(v int) bool {
	return r.Start <= v && v <= r.End
}

func (r Range) Width() int {
	return r.End - r.Start
}
