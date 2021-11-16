// Copyright 2021  The CDK Authors
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

// AttrMask represents a mask of text attributes, apart from color.
// Note that support for attributes may vary widely across terminals.
type AttrMask int

// Attributes are not colors, but affect the display of text.  They can
// be combined.
const (
	AttrBold AttrMask = 1 << iota
	AttrBlink
	AttrReverse
	AttrUnderline
	AttrDim
	AttrItalic
	AttrStrike
	AttrInvalid              // Mark the style or attributes invalid
	AttrNone    AttrMask = 0 // Just normal text.
)

const attrAll = AttrBlink | AttrBold | AttrReverse | AttrUnderline | AttrDim | AttrItalic | AttrStrike

// check if the attributes are normal
func (m AttrMask) IsNormal() bool {
	return m == AttrNone
}

// check if the attributes include bold
func (m AttrMask) IsBold() bool {
	return m&AttrBold != 0
}

// check if the attributes include blink
func (m AttrMask) IsBlink() bool {
	return m&AttrBlink != 0
}

// check if the attributes include reverse
func (m AttrMask) IsReverse() bool {
	return m&AttrReverse != 0
}

// check if the attributes include underline
func (m AttrMask) IsUnderline() bool {
	return m&AttrUnderline != 0
}

// check if the attributes include dim
func (m AttrMask) IsDim() bool {
	return m&AttrDim != 0
}

// check if the attributes include italics
func (m AttrMask) IsItalic() bool {
	return m&AttrItalic != 0
}

// check if the attributes include italics
func (m AttrMask) IsStrike() bool {
	return m&AttrStrike != 0
}

// return a normal attribute mask
func (m AttrMask) Normal() AttrMask {
	return m &^ attrAll
}

// return the attributes with (true) or without (false) bold
func (m AttrMask) Bold(v bool) AttrMask {
	if v {
		return m | AttrBold
	}
	return m &^ AttrBold
}

// return the attributes with (true) or without (false) blink
func (m AttrMask) Blink(v bool) AttrMask {
	if v {
		return m | AttrBlink
	}
	return m &^ AttrBlink
}

// return the attributes with (true) or without (false) reverse
func (m AttrMask) Reverse(v bool) AttrMask {
	if v {
		return m | AttrReverse
	}
	return m &^ AttrReverse
}

// return the attributes with (true) or without (false) underline
func (m AttrMask) Underline(v bool) AttrMask {
	if v {
		return m | AttrUnderline
	}
	return m &^ AttrUnderline
}

// return the attributes with (true) or without (false) dim
func (m AttrMask) Dim(v bool) AttrMask {
	if v {
		return m | AttrDim
	}
	return m &^ AttrDim
}

// return the attributes with (true) or without (false) italic
func (m AttrMask) Italic(v bool) AttrMask {
	if v {
		return m | AttrItalic
	}
	return m &^ AttrItalic
}

// return the attributes with (true) or without (false) strikethrough
func (m AttrMask) Strike(v bool) AttrMask {
	if v {
		return m | AttrStrike
	}
	return m &^ AttrStrike
}
