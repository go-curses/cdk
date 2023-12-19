// Copyright (c) 2021-2023  The Go-Curses Authors
// Copyright 2015 The TCell Authors
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

package paint

import (
	"math"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestColorBasics(t *testing.T) {
	Convey("Color Basics", t, func() {
		blue := GetColor("blue")
		So(blue.String(), ShouldEqual, "blue[#0000ff]")
		unknown := GetColor("unknown")
		So(unknown.String(), ShouldEqual, "unnamed[-1]")
	})
}

func TestColorValues(t *testing.T) {
	var values = []struct {
		color Color
		hex   int32
	}{
		{ColorRed, 0x00FF0000},
		{ColorGreen, 0x00008000},
		{ColorLime, 0x0000FF00},
		{ColorBlue, 0x000000FF},
		{ColorBlack, 0x00000000},
		{ColorWhite, 0x00FFFFFF},
		{ColorSilver, 0x00C0C0C0},
	}

	Convey("Color Values", t, func() {
		for _, tc := range values {
			So(tc.color.Hex() != tc.hex, ShouldEqual, false)
		}
	})
}

func TestColorFitting(t *testing.T) {
	pal := []Color{}
	for i := 0; i < 255; i++ {
		pal = append(pal, PaletteColor(i))
	}

	Convey("Color Fitting", t, func() {
		// Exact color fitting on ANSI colors
		for i := 0; i < 7; i++ {
			if FindColor(PaletteColor(i), pal[:8]) != PaletteColor(i) {
				t.Errorf("Color ANSI fit fail at %d", i)
			}
		}
		// Grey is closest to Silver
		if FindColor(PaletteColor(8), pal[:8]) != PaletteColor(7) {
			t.Errorf("Grey does not fit to silver")
		}
		// Color fitting of upper 8 colors.
		for i := 9; i < 16; i++ {
			if FindColor(PaletteColor(i), pal[:8]) != PaletteColor(i%8) {
				t.Errorf("Color fit fail at %d", i)
			}
		}
		// Imperfect fit
		if FindColor(ColorOrangeRed, pal[:16]) != ColorRed ||
			FindColor(ColorAliceBlue, pal[:16]) != ColorWhite ||
			FindColor(ColorPink, pal) != Color217 ||
			FindColor(ColorSienna, pal) != Color173 ||
			FindColor(GetColor("#00FD00"), pal) != ColorLime {
			t.Errorf("Imperfect color fit")
		}
		// if value is NaN, safe_nan produces '+Inf'
		nd := safe_nan(math.Log(-1.0))
		So(nd, ShouldEqual, math.Inf(+1))
	})
}

func TestColorNameLookup(t *testing.T) {
	var values = []struct {
		name  string
		color Color
		rgb   bool
	}{
		{"#FF0000", ColorRed, true},
		{"black", ColorBlack, false},
		{"orange", ColorOrange, false},
		{"door", ColorDefault, false},
	}
	Convey("Color Name Lookups", t, func() {
		for _, v := range values {
			c := GetColor(v.name)
			if c.Hex() != v.color.Hex() {
				t.Errorf("Wrong color for %v: %v", v.name, c.Hex())
			}
			if v.rgb {
				if c&ColorIsRGB == 0 {
					t.Errorf("Color should have RGB")
				}
			} else {
				if c&ColorIsRGB != 0 {
					t.Errorf("Named color should not be RGB")
				}
			}
			if c.TrueColor().Hex() != v.color.Hex() {
				t.Errorf("TrueColor did not match")
			}
		}
	})
}

func TestColorRGB(t *testing.T) {
	Convey("Color RGB", t, func() {
		r, g, b := GetColor("#112233").RGB()
		So(r, ShouldEqual, 0x11)
		So(g, ShouldEqual, 0x22)
		So(b, ShouldEqual, 0x33)
		c := Color(0xfffffff)
		So(c.IsRGB(), ShouldEqual, false)
		r, g, b = c.RGB()
		So(r, ShouldEqual, -1)
		So(g, ShouldEqual, -1)
		So(b, ShouldEqual, -1)
		c = NewRGBColor(0x11, 0x22, 0x33)
		r, g, b = c.RGB()
		So(c.IsRGB(), ShouldEqual, true)
		So(r, ShouldEqual, 0x11)
		So(g, ShouldEqual, 0x22)
		So(b, ShouldEqual, 0x33)
	})
}
