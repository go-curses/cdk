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

import (
	"fmt"
	"regexp"
	"strconv"
)

// Rectangle is a 2-aspect vector represented by width and height values
type Rectangle struct {
	W, H int
}

// Construct a new instance of a Point2I structure
func NewRectangle(w, h int) *Rectangle {
	r := MakeRectangle(w, h)
	return &r
}

// Construct a new Point2I structure (non-pointer)
func MakeRectangle(w, h int) Rectangle {
	return Rectangle{W: w, H: h}
}

// Parse a Point2I structure from a string representation. There are two valid
// formats supported by this parser function:
//   formal    "{w:0,h:0}"
//   plain     "0 0"
func ParseRectangle(value string) (point Rectangle, ok bool) {
	if rxParseRectangle.MatchString(value) {
		m := rxParseRectangle.FindStringSubmatch(value)
		if len(m) == 3 {
			w, _ := strconv.Atoi(m[1])
			h, _ := strconv.Atoi(m[2])
			return MakeRectangle(w, h), true
		}
	}
	if rxParseTwoDigits.MatchString(value) {
		m := rxParseTwoDigits.FindStringSubmatch(value)
		if len(m) == 3 {
			w, _ := strconv.Atoi(m[1])
			h, _ := strconv.Atoi(m[2])
			return MakeRectangle(w, h), true
		}
	}
	return Rectangle{}, false
}

// returns a formal string representation of the Point2I structure, ie: "{w:0,h:0}"
func (r Rectangle) String() string {
	return fmt.Sprintf("{w:%v,h:%v}", r.W, r.H)
}

// returns a new Point2I structure with the same values as this structure
func (r Rectangle) Clone() (clone Rectangle) {
	clone.W = r.W
	clone.H = r.H
	return
}

// returns a new Point2I instance with the same values as this structure
func (r Rectangle) NewClone() (clone *Rectangle) {
	clone = NewRectangle(r.W, r.H)
	return
}

// returns true if both the given x and y coordinates and this Point2I are
// equivalent, returns false otherwise
func (r Rectangle) Equals(w, h int) bool {
	return r.W == w && r.H == h
}

// returns true if both the given Point2I and this Point2I are equivalent and
// false otherwise
func (r Rectangle) EqualsTo(o Rectangle) bool {
	return r.W == o.W && r.H == o.H
}

// set this Point2I instance to the given x and y values
func (r *Rectangle) Set(w, h int) {
	r.W = w
	r.H = h
}

// set this Point2I instance to be equivalent to the given Point2I
func (r *Rectangle) SetRectangle(size Rectangle) {
	r.W = size.W
	r.H = size.H
}

// add the given w and h values to this Point2I
func (r *Rectangle) Add(w, h int) {
	r.W += w
	r.H += h
}

// add the given Point2I to this Point2I
func (r *Rectangle) AddRectangle(size Rectangle) {
	r.W += size.W
	r.H += size.H
}

// subtract the given x and y values from this Point2I instance
func (r *Rectangle) Sub(w, h int) {
	r.W -= w
	r.H -= h
}

// subtract the given Point2I's values from this Point2I instance
func (r *Rectangle) SubRectangle(size Rectangle) {
	r.W -= size.W
	r.H -= size.H
}

// returns the volume of the Rectangle (w * h)
func (r Rectangle) Volume() int {
	return r.W * r.H
}

// constrain the width and height values to be at least the given values or
// greater
func (r *Rectangle) Floor(minWidth, minHeight int) {
	if r.W < minWidth {
		r.W = minWidth
	}
	if r.H < minHeight {
		r.H = minHeight
	}
}

// constrain the width and height values to be within the given ranges of
// min and max values
func (r *Rectangle) Clamp(minWidth, minHeight, maxWidth, maxHeight int) {
	if r.W < minWidth {
		r.W = minWidth
	}
	if r.H < minHeight {
		r.H = minHeight
	}
	if r.W > maxWidth {
		r.W = maxWidth
	}
	if r.H > maxHeight {
		r.H = maxHeight
	}
}

func (r *Rectangle) ClampToRegion(region Region) (clamped bool) {
	clamped = false
	min, max := region.Origin(), region.FarPoint()
	// is width within range?
	if r.W >= min.X && r.W <= max.X {
		// width is within range, NOP
	} else {
		// width is not within range
		if r.W < min.X {
			// width is too low, CLAMP
			r.W = min.X
		} else {
			// width is too high, CLAMP
			r.W = max.X
		}
		clamped = true
	}
	// is height within range?
	if r.H >= min.Y && r.H <= max.Y {
		// height is within range, NOP
	} else {
		// height is not within range
		if r.H < min.Y {
			// height is too low, CLAMP
			r.H = min.Y
		} else {
			// height is too high, CLAMP
			r.H = max.Y
		}
		clamped = true
	}
	return
}

var (
	rxParseRectangle = regexp.MustCompile(`(?:i)^{??(?:w:)??(\d+),(?:h:)??(\d+)}??$`)
	rxParseTwoDigits = regexp.MustCompile(`(?:i)^\s*(\d+)\s*(\d+)\s*$`)
)
