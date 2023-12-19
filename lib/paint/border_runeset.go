// Copyright (c) 2023  The Go-Curses Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
