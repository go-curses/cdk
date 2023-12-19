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

type ArrowName string

const (
	NilArrow   ArrowName = "nil"
	StockArrow ArrowName = "stock"
	WideArrow  ArrowName = "wide"
)

var (
	arrowOverrides = map[ArrowName]ArrowRuneSet{}
)

func RegisterArrows(name ArrowName, arrow ArrowRuneSet) {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	arrowOverrides[name] = arrow
}

func GetArrows(name ArrowName) (arrow ArrowRuneSet, ok bool) {
	pkgLock.RLock()
	defer pkgLock.RUnlock()
	if arrow, ok = arrowOverrides[name]; !ok {
		switch name {
		case StockArrow:
			return stockArrowRune, true
		case WideArrow:
			return wideArrowRune, true
		case NilArrow:
			return ArrowRuneSet{}, true
		}
	}
	return
}
