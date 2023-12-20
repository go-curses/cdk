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

var (
	SafeSpinnerRuneSet = SpinnerRuneSet{
		'/',
		'|',
		'\\',
		'-',
	}
	OneDotSpinnerRuneSet = SpinnerRuneSet{
		RuneOneDotSpinner0,
		RuneOneDotSpinner1,
		RuneOneDotSpinner2,
		RuneOneDotSpinner3,
		RuneOneDotSpinner4,
		RuneOneDotSpinner5,
		RuneOneDotSpinner6,
		RuneOneDotSpinner7,
	}
	OrbitDotSpinnerRuneSet = SpinnerRuneSet{
		RuneOrbitDotSpinner0,
		RuneOrbitDotSpinner1,
		RuneOrbitDotSpinner2,
		RuneOrbitDotSpinner3,
	}
	SevenDotSpinnerRuneSet = SpinnerRuneSet{
		RuneSevenDotSpinner0,
		RuneSevenDotSpinner1,
		RuneSevenDotSpinner2,
		RuneSevenDotSpinner3,
		RuneSevenDotSpinner4,
		RuneSevenDotSpinner5,
		RuneSevenDotSpinner6,
		RuneSevenDotSpinner7,
	}
)

type SpinnerRuneSet []rune

func (b SpinnerRuneSet) String() string {
	return fmt.Sprintf("{SpinnerRunes=%+v}", b[:])
}

func (b SpinnerRuneSet) Strings() (s []string) {
	for _, c := range b {
		s = append(s, string(c))
	}
	return
}