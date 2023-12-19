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
	"strings"

	"github.com/iancoleman/strcase"
)

type TypeTag interface {
	Tag() CTypeTag
	String() string
	GladeString() string
	ClassName() string
	Equals(tt TypeTag) bool
	Less(tt TypeTag) bool
	Valid() bool
}

// denotes a concrete type identity
type CTypeTag string

// constructs a new TypeTag instance
func MakeTypeTag(tag string) TypeTag {
	return NewTypeTag(tag)
}

// constructs a new Concrete TypeTag instance
func NewTypeTag(tag string) CTypeTag {
	return CTypeTag(tag)
}

// returns the underlying CTypeTag instance
func (tag CTypeTag) Tag() CTypeTag {
	return tag
}

// Stringer interface implementation
func (tag CTypeTag) String() string {
	return string(tag)
}

// returns a string representation of this type tag, translated for Gtk class
// naming conventions (ie: GtkCamelCase)
func (tag CTypeTag) GladeString() string {
	gt := strcase.ToCamel(tag.String())
	return strings.Replace(gt, "Ctk", "Gtk", 1)
}

// returns the CamelCase, or "Class Name", version of this type tag
func (tag CTypeTag) ClassName() string {
	return strcase.ToCamel(tag.String())
}

// returns true if the given type tag is the same as this type tag
func (tag CTypeTag) Equals(tt TypeTag) bool {
	return string(tag) == tt.String()
}

// returns true if this type tag is numerically less than the given type tag,
// used in sorting routines
func (tag CTypeTag) Less(tt TypeTag) bool {
	return string(tag) < tt.String()
}

func (tag CTypeTag) Valid() bool {
	return string(tag) != ""
}
