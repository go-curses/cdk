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

// Point2I and Rectangle combined

import (
	"fmt"
	"regexp"
	"strconv"
)

type Region struct {
	Point2I
	Rectangle
}

func NewRegion(x, y, w, h int) *Region {
	r := MakeRegion(x, y, w, h)
	return &r
}

func MakeRegion(x, y, w, h int) Region {
	return Region{
		Point2I{X: x, Y: y},
		Rectangle{W: w, H: h},
	}
}

func ParseRegion(value string) (point Region, ok bool) {
	if rxParseRegion.MatchString(value) {
		m := rxParseRegion.FindStringSubmatch(value)
		if len(m) == 5 {
			x, _ := strconv.Atoi(m[1])
			y, _ := strconv.Atoi(m[2])
			w, _ := strconv.Atoi(m[3])
			h, _ := strconv.Atoi(m[4])
			return MakeRegion(x, y, w, h), true
		}
	}
	if rxParseFourDigits.MatchString(value) {
		m := rxParseFourDigits.FindStringSubmatch(value)
		if len(m) == 5 {
			x, _ := strconv.Atoi(m[1])
			y, _ := strconv.Atoi(m[2])
			w, _ := strconv.Atoi(m[3])
			h, _ := strconv.Atoi(m[4])
			return MakeRegion(x, y, w, h), true
		}
	}
	return Region{}, false
}

func (r Region) String() string {
	return fmt.Sprintf("{x:%v,y:%v,w:%v,h:%v}", r.X, r.Y, r.H, r.W)
}

func (r Region) Clone() (clone Region) {
	clone.X = r.X
	clone.Y = r.Y
	clone.W = r.W
	clone.H = r.H
	return
}

func (r Region) NewClone() (clone *Region) {
	clone = NewRegion(r.X, r.Y, r.W, r.H)
	return
}

func (r *Region) Set(x, y, w, h int) {
	r.X, r.Y, r.W, r.H = x, y, w, h
}

func (r *Region) SetRegion(region Region) {
	r.X, r.Y, r.W, r.H = region.X, region.Y, region.W, region.H
}

func (r Region) HasPoint(pos Point2I) bool {
	if r.X <= pos.X {
		if r.Y <= pos.Y {
			if (r.X + r.W) >= pos.X {
				if (r.Y + r.H) >= pos.Y {
					return true
				}
			}
		}
	}
	return false
}

func (r Region) Origin() Point2I {
	return Point2I{r.X, r.Y}
}

func (r Region) FarPoint() Point2I {
	return Point2I{
		r.X + r.W,
		r.Y + r.H,
	}
}

func (r Region) Size() Rectangle {
	return Rectangle{r.W, r.H}
}

func (r *Region) ClampToRegion(region Region) (clamped bool) {
	var x, y, w, h int
	fp := r.FarPoint()
	rfp := region.FarPoint()
	if r.X < region.X {
		// too far left
		x = region.X
	} else if r.X > rfp.X {
		// too far right
		x = rfp.X
	} else {
		// accept
		x = r.X
	}
	if r.Y < region.Y {
		// too far up
		y = region.Y
	} else if r.Y > rfp.Y {
		// too far down
		y = rfp.Y
	} else {
		// accept
		y = r.Y
	}
	if fp.X > rfp.X {
		// too wide
		w = rfp.X - region.X
	} else {
		// accept
		w = r.W
	}
	if fp.Y > rfp.Y {
		// to tall
		h = rfp.Y - region.Y
	} else {
		// accept
		h = r.W
	}
	// apply changes
	clamped = r.X != x || r.Y != y || r.W != w || r.H != h
	r.X, r.Y, r.W, r.H = x, y, w, h
	return
}

var (
	rxParseRegion     = regexp.MustCompile(`(?:i)^{??(?:x:)??(\d+),(?:y:)??(\d+),(?:w:)??(\d+),(?:h:)??(\d+)}??$`)
	rxParseFourDigits = regexp.MustCompile(`(?:i)^\s*(\d+)\s*(\d+)\s*(\d+)\s*(\d+)\s*$`)
)
