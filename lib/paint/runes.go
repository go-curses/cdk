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

// The names of these constants are chosen to match Terminfo names,
// modulo case, and changing the prefix from ACS_ to Rune.  These are
// the runes we provide extra special handling for, with ASCII fallbacks
// for terminals that lack them.
const (
	RuneBoxLineX = '╳'
	RuneCheckbox = '✔'
	RuneSterling = '£'
	RuneDArrow   = '↓'
	RuneLArrow   = '←'
	RuneRArrow   = '→'
	RuneUArrow   = '↑'
	RuneBullet   = '·'
	RuneBoard    = '░'
	RuneCkBoard  = '▒'
	RuneDegree   = '°'
	RuneDiamond  = '◆'
	RuneGEqual   = '≥'
	RunePi       = 'π'
	RuneHLine    = '─'
	RuneLantern  = '§'
	RunePlus     = '┼'
	RuneLEqual   = '≤'
	RuneLLCorner = '└'
	RuneLRCorner = '┘'
	RuneNEqual   = '≠'
	RunePlMinus  = '±'
	RuneS1       = '⎺'
	RuneS3       = '⎻'
	RuneS7       = '⎼'
	RuneS9       = '⎽'
	RuneBlock    = '█'
	RuneTTee     = '┬'
	RuneRTee     = '┤'
	RuneLTee     = '├'
	RuneBTee     = '┴'
	RuneULCorner = '┌'
	RuneURCorner = '┐'
	RuneVLine    = '│'

	RuneLLCornerRounded = '╰'
	RuneLRCornerRounded = '╯'
	RuneULCornerRounded = '╭'
	RuneURCornerRounded = '╮'

	RuneLeftwardsTwoHeadedArrowWithTriangleArrowheads  = '⯬'
	RuneUpwardsTwoHeadedArrowWithTriangleArrowheads    = '⯭'
	RuneRightwardsTwoHeadedArrowWithTriangleArrowheads = '⯮'
	RuneDownwardsTwoHeadedArrowWithTriangleArrowheads  = '⯯'
	RuneFilledSquareCentred                            = '⯀'
	RuneFilledMediumUpPointingTriangleCentred          = '⯅'
	RuneFilledMediumDownPointingTriangleCentred        = '⯆'
	RuneFilledMediumLeftPointingTriangleCentred        = '⯇'
	RuneFilledMediumRightPointingTriangleCentred       = '⯈'
	RuneLeftwardsFilledCircledHollowArrow              = '⮈'
	RuneUpwardsFilledCircledHollowArrow                = '⮉'
	RuneRightwardsFilledCircledHollowArrow             = '⮊'
	RuneDownwardsFilledCircledHollowArrow              = '⮋'

	RuneTriangleUp    = '▲'
	RuneTriangleLeft  = '◀'
	RuneTriangleDown  = '▼'
	RuneTriangleRight = '▶'

	RuneEllipsis = '…'

	RuneBraille1235678 = '⣷'
	RuneBraille2345678 = '⣾'
	RuneBraille1345678 = '⣽'
	RuneBraille1245678 = '⣻'
	RuneBraille1234568 = '⢿'
	RuneBraille1234567 = '⡿'
	RuneBraille1234578 = '⣟'
	RuneBraille1234678 = '⣯'

	RuneBraille1 = '⠁'
	RuneBraille2 = '⠂'
	RuneBraille3 = '⠄'
	RuneBraille4 = '⠈'
	RuneBraille5 = '⠐'
	RuneBraille6 = '⠠'
	RuneBraille7 = '⡀'
	RuneBraille8 = '⢀'

	RuneBraille47 = '⡈'
	RuneBraille18 = '⢁'
	RuneBraille26 = '⠢'
	RuneBraille35 = '⠔'

	RuneSevenDotSpinner0 = RuneBraille1235678
	RuneSevenDotSpinner1 = RuneBraille2345678
	RuneSevenDotSpinner2 = RuneBraille1345678
	RuneSevenDotSpinner3 = RuneBraille1245678
	RuneSevenDotSpinner4 = RuneBraille1234568
	RuneSevenDotSpinner5 = RuneBraille1234567
	RuneSevenDotSpinner6 = RuneBraille1234578
	RuneSevenDotSpinner7 = RuneBraille1234678

	RuneOneDotSpinner0 = RuneBraille4
	RuneOneDotSpinner1 = RuneBraille1
	RuneOneDotSpinner2 = RuneBraille2
	RuneOneDotSpinner3 = RuneBraille3
	RuneOneDotSpinner4 = RuneBraille7
	RuneOneDotSpinner5 = RuneBraille8
	RuneOneDotSpinner6 = RuneBraille6
	RuneOneDotSpinner7 = RuneBraille5

	RuneOrbitDotSpinner0 = RuneBraille47
	RuneOrbitDotSpinner1 = RuneBraille18
	RuneOrbitDotSpinner2 = RuneBraille26
	RuneOrbitDotSpinner3 = RuneBraille35
)

// RuneFallbacks is the default map of fallback strings that will be
// used to replace a rune when no other more appropriate transformation
// is available, and the rune cannot be displayed directly.
//
// New entries may be added to this map over time, as it becomes clear
// that such is desirable.  Characters that represent either letters or
// numbers should not be added to this list unless it is certain that
// the meaning will still convey unambiguously.
//
// As an example, it would be appropriate to add an ASCII mapping for
// the full width form of the letter 'A', but it would not be appropriate
// to do so a glyph representing the country China.
//
// Programs that desire richer fallbacks may register additional ones,
// or change or even remove these mappings with Screen.RegisterRuneFallback
// Screen.UnregisterRuneFallback methods.
//
// Note that Unicode is presumed to be able to display all glyphs.
// This is a pretty poor assumption, but there is no easy way to
// figure out which glyphs are supported in a given font.  Hence,
// some care in selecting the characters you support in your application
// is still appropriate.
var RuneFallbacks = map[rune]string{
	RuneBoxLineX: "X",
	RuneCheckbox: "*",
	RuneSterling: "f",
	RuneDArrow:   "v",
	RuneLArrow:   "<",
	RuneRArrow:   ">",
	RuneUArrow:   "^",
	RuneBullet:   "o",
	RuneBoard:    "#",
	RuneCkBoard:  ":",
	RuneDegree:   "\\",
	RuneDiamond:  "+",
	RuneGEqual:   ">",
	RunePi:       "*",
	RuneHLine:    "-",
	RuneLantern:  "#",
	RunePlus:     "+",
	RuneLEqual:   "<",
	RuneLLCorner: "+",
	RuneLRCorner: "+",
	RuneNEqual:   "!",
	RunePlMinus:  "#",
	RuneS1:       "~",
	RuneS3:       "-",
	RuneS7:       "-",
	RuneS9:       "_",
	RuneBlock:    "#",
	RuneTTee:     "+",
	RuneRTee:     "+",
	RuneLTee:     "+",
	RuneBTee:     "+",
	RuneULCorner: "+",
	RuneURCorner: "+",
	RuneVLine:    "|",

	RuneLLCornerRounded: "+",
	RuneLRCornerRounded: "+",
	RuneULCornerRounded: "+",
	RuneURCornerRounded: "+",

	RuneLeftwardsTwoHeadedArrowWithTriangleArrowheads:  "<",
	RuneUpwardsTwoHeadedArrowWithTriangleArrowheads:    "^",
	RuneRightwardsTwoHeadedArrowWithTriangleArrowheads: ">",
	RuneDownwardsTwoHeadedArrowWithTriangleArrowheads:  "v",
	RuneFilledSquareCentred:                            "#",
	RuneFilledMediumUpPointingTriangleCentred:          "^",
	RuneFilledMediumDownPointingTriangleCentred:        "v",
	RuneFilledMediumLeftPointingTriangleCentred:        "<",
	RuneFilledMediumRightPointingTriangleCentred:       ">",
	RuneLeftwardsFilledCircledHollowArrow:              "<",
	RuneUpwardsFilledCircledHollowArrow:                "^",
	RuneRightwardsFilledCircledHollowArrow:             ">",
	RuneDownwardsFilledCircledHollowArrow:              "v",

	RuneTriangleUp:    "^",
	RuneTriangleLeft:  "<",
	RuneTriangleDown:  "v",
	RuneTriangleRight: ">",
}

var RuneConsoleFallbacks = map[rune]rune{
	RuneSevenDotSpinner0: '/',
	RuneSevenDotSpinner1: '|',
	RuneSevenDotSpinner2: '\\',
	RuneSevenDotSpinner3: '-',
	RuneSevenDotSpinner4: '/',
	RuneSevenDotSpinner5: '|',
	RuneSevenDotSpinner6: '\\',
	RuneSevenDotSpinner7: '-',

	RuneOneDotSpinner0: '/',
	RuneOneDotSpinner1: '|',
	RuneOneDotSpinner2: '\\',
	RuneOneDotSpinner3: '-',
	RuneOneDotSpinner4: '/',
	RuneOneDotSpinner5: '|',
	RuneOneDotSpinner6: '\\',
	RuneOneDotSpinner7: '-',

	RuneOrbitDotSpinner0: '/',
	RuneOrbitDotSpinner1: '|',
	RuneOrbitDotSpinner2: '\\',
	RuneOrbitDotSpinner3: '-',
}