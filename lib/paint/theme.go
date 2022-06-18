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

package paint

import (
	"fmt"
)

var (
	DefaultFillRune   = ' '
	DefaultMonoStyle  = StyleDefault.Reverse(false).Dim(true)
	DefaultColorStyle = StyleDefault.Foreground(ColorWhite).Background(ColorNavy)
	DefaultBorderRune = BorderRuneSet{
		TopLeft:     RuneULCorner,
		Top:         RuneHLine,
		TopRight:    RuneURCorner,
		Left:        RuneVLine,
		Right:       RuneVLine,
		BottomLeft:  RuneLLCorner,
		Bottom:      RuneHLine,
		BottomRight: RuneLRCorner,
	}
	EmptyBorderRune = BorderRuneSet{
		TopLeft:     ' ',
		Top:         ' ',
		TopRight:    ' ',
		Left:        ' ',
		Right:       ' ',
		BottomLeft:  ' ',
		Bottom:      ' ',
		BottomRight: ' ',
	}
	NilBorderRune = BorderRuneSet{
		TopLeft:     rune(0),
		Top:         rune(0),
		TopRight:    rune(0),
		Left:        rune(0),
		Right:       rune(0),
		BottomLeft:  rune(0),
		Bottom:      rune(0),
		BottomRight: rune(0),
	}
	DefaultArrowRune = ArrowRuneSet{
		Up:    RuneUArrow,
		Left:  RuneLArrow,
		Down:  RuneDArrow,
		Right: RuneRArrow,
	}
)

var (
	DefaultNilTheme        = Theme{}
	DefaultMonoThemeAspect = ThemeAspect{
		Normal:      DefaultMonoStyle,
		Selected:    DefaultMonoStyle.Dim(false),
		Active:      DefaultMonoStyle.Dim(false).Reverse(true),
		Prelight:    DefaultMonoStyle.Dim(false),
		Insensitive: DefaultMonoStyle.Dim(true),
		FillRune:    DefaultFillRune,
		BorderRunes: DefaultBorderRune,
		ArrowRunes:  DefaultArrowRune,
		Overlay:     false,
	}
	DefaultColorThemeAspect = ThemeAspect{
		Normal:      DefaultColorStyle.Dim(true),
		Selected:    DefaultColorStyle.Dim(false),
		Active:      DefaultColorStyle.Dim(false).Reverse(true),
		Prelight:    DefaultColorStyle.Dim(false),
		Insensitive: DefaultColorStyle.Dim(true),
		FillRune:    DefaultFillRune,
		BorderRunes: DefaultBorderRune,
		ArrowRunes:  DefaultArrowRune,
		Overlay:     false,
	}
	DefaultMonoTheme = Theme{
		Content: DefaultMonoThemeAspect,
		Border:  DefaultMonoThemeAspect,
	}
	DefaultColorTheme = Theme{
		Content: DefaultColorThemeAspect,
		Border:  DefaultColorThemeAspect,
	}
)

type BorderRuneSet struct {
	TopLeft     rune
	Top         rune
	TopRight    rune
	Left        rune
	Right       rune
	BottomLeft  rune
	Bottom      rune
	BottomRight rune
}

func (b BorderRuneSet) String() string {
	return fmt.Sprintf(
		"{BorderRunes=%v,%v,%v,%v,%v,%v,%v,%v}",
		b.TopRight,
		b.Top,
		b.TopLeft,
		b.Left,
		b.BottomLeft,
		b.Bottom,
		b.BottomRight,
		b.Right,
	)
}

type ArrowRuneSet struct {
	Up    rune
	Left  rune
	Down  rune
	Right rune
}

func (b ArrowRuneSet) String() string {
	return fmt.Sprintf(
		"{ArrowRunes=%v,%v,%v,%v}",
		b.Up,
		b.Left,
		b.Down,
		b.Right,
	)
}

type ThemeAspect struct {
	Normal      Style
	Selected    Style
	Active      Style
	Prelight    Style
	Insensitive Style
	FillRune    rune
	BorderRunes BorderRuneSet
	ArrowRunes  ArrowRuneSet
	Overlay     bool // keep existing background
}

func (t ThemeAspect) String() string {
	return fmt.Sprintf(
		"{Normal=%v,Selected=%v,Active=%v,Prelight=%v,Insensitive=%v,FillRune=%v,BorderRunes=%v,ArrowRunes=%v,Overlay=%v}",
		t.Normal,
		t.Selected,
		t.Active,
		t.Prelight,
		t.Insensitive,
		t.FillRune,
		t.BorderRunes,
		t.ArrowRunes,
		t.Overlay,
	)
}

type Theme struct {
	Content ThemeAspect
	Border  ThemeAspect
}

func (t Theme) String() string {
	return fmt.Sprintf(
		"{Content=%v,Border=%v}",
		t.Content,
		t.Border,
	)
}