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

type BorderName string

const (
	NilBorder     BorderName = "nil"
	EmptyBorder   BorderName = "empty"
	StockBorder   BorderName = "standard"
	RoundedBorder BorderName = "rounded"
	DoubleBorder  BorderName = "double"
)

var (
	borderOverrides = map[BorderName]BorderRuneSet{}
)

func RegisterBorderRunes(name BorderName, border BorderRuneSet) {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	borderOverrides[name] = border
}

func GetDefaultBorderRunes(name BorderName) (border BorderRuneSet, ok bool) {
	pkgLock.RLock()
	defer pkgLock.RUnlock()
	if border, ok = borderOverrides[name]; !ok {
		switch name {
		case EmptyBorder:
			return emptyBorderRune, true
		case StockBorder:
			return stockBorderRune, true
		case RoundedBorder:
			return roundedBorderRune, true
		case DoubleBorder:
			return doubleBorderRune, true
		case NilBorder:
			return nilBorderRune, true
		}
	}
	return
}
