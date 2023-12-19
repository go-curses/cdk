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

// Point2I is a 2-aspect vector represented by x and y coordinates.
type Point2I struct {
	X, Y int
}

// Construct a new instance of a Point2I structure
func NewPoint2I(x, y int) *Point2I {
	r := MakePoint2I(x, y)
	return &r
}

// Construct a new Point2I structure (non-pointer)
func MakePoint2I(x, y int) Point2I {
	return Point2I{X: x, Y: y}
}

// Parse a Point2I structure from a string representation. There are two valid
// formats supported by this parser function:
//
//	formal    "{x:0,y:0}"
//	plain     "0 0"
func ParsePoint2I(value string) (point Point2I, ok bool) {
	if rxParsePoint2I.MatchString(value) {
		m := rxParsePoint2I.FindStringSubmatch(value)
		if len(m) == 3 {
			x, _ := strconv.Atoi(m[1])
			y, _ := strconv.Atoi(m[2])
			return MakePoint2I(x, y), true
		}
	}
	if rxParseTwoDigits.MatchString(value) {
		m := rxParseTwoDigits.FindStringSubmatch(value)
		if len(m) == 3 {
			x, _ := strconv.Atoi(m[1])
			y, _ := strconv.Atoi(m[2])
			return MakePoint2I(x, y), true
		}
	}
	return Point2I{}, false
}

// returns a formal string representation of the Point2I structure, ie: "{x:0,y:0}"
func (p Point2I) String() string {
	return fmt.Sprintf("{x:%v,y:%v}", p.X, p.Y)
}

// returns a new Point2I structure with the same values as this structure
func (p Point2I) Clone() (clone Point2I) {
	clone.X = p.X
	clone.Y = p.Y
	return
}

// returns a new Point2I instance with the same values as this structure
func (p Point2I) NewClone() (clone *Point2I) {
	clone = NewPoint2I(p.X, p.Y)
	return
}

// Position returns the X and Y coordinates
func (p Point2I) Position() (x, y int) {
	x, y = p.X, p.Y
	return
}

// returns true if both the given x and y coordinates and this Point2I are
// equivalent, returns false otherwise
func (p Point2I) Equals(x, y int) bool {
	return p.X == x && p.Y == y
}

// returns true if both the given Point2I and this Point2I are equivalent and
// false otherwise
func (p Point2I) EqualsTo(o Point2I) bool {
	return p.X == o.X && p.Y == o.Y
}

// set this Point2I instance to the given x and y values
func (p *Point2I) Set(x, y int) {
	p.X = x
	p.Y = y
}

// set this Point2I instance to be equivalent to the given Point2I
func (p *Point2I) SetPoint(point Point2I) {
	p.X = point.X
	p.Y = point.Y
}

// add the given x and y values to this Point2I
func (p *Point2I) Add(x, y int) {
	p.X += x
	p.Y += y
}

// add the given Point2I to this Point2I
func (p *Point2I) AddPoint(point Point2I) {
	p.X += point.X
	p.Y += point.Y
}

// subtract the given x and y values from this Point2I instance
func (p *Point2I) Sub(x, y int) {
	p.X -= x
	p.Y -= y
}

// subtract the given Point2I's values from this Point2I instance
func (p *Point2I) SubPoint(point Point2I) {
	p.X -= point.X
	p.Y -= point.Y
}

func (p *Point2I) ClampMax(x, y int) {
	if p.X > x {
		p.X = x
	}
	if p.Y > y {
		p.Y = y
	}
}

func (p *Point2I) ClampMin(x, y int) {
	if p.X < x {
		p.X = x
	}
	if p.Y < y {
		p.Y = y
	}
}

// restrict this Point2I instance to be within the boundaries defined by the
// given region
func (p *Point2I) ClampToRegion(region Region) (clamped bool) {
	clamped = false
	min, max := region.Origin(), region.FarPoint()
	// is width within range?
	if p.X >= min.X && p.X <= max.X {
		// width is within range, NOP
	} else {
		// width is not within range
		if p.X < min.X {
			// width is too low, CLAMP
			p.X = min.X
		} else {
			// width is too high, CLAMP
			p.X = max.X
		}
		clamped = true
	}
	// is height within range?
	if p.Y >= min.Y && p.Y <= max.Y {
		// height is within range, NOP
	} else {
		// height is not within range
		if p.Y < min.Y {
			// height is too low, CLAMP
			p.Y = min.Y
		} else {
			// height is too high, CLAMP
			p.Y = max.Y
		}
		clamped = true
	}
	return
}

var (
	rxParsePoint2I = regexp.MustCompile(`(?:i)^{??(?:x:)??(\d+),(?:y:)??(\d+)}??$`)
)
