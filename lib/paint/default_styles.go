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

type StyleName string

const (
	NilStyle   StyleName = "nil"
	MonoStyle  StyleName = "mono"
	ColorStyle StyleName = "color"
)

var (
	styleOverrides = map[StyleName]Style{}
)

func RegisterStyle(name StyleName, theme Style) {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	styleOverrides[name] = theme
}

func GetStyle(name StyleName) (theme Style, ok bool) {
	pkgLock.RLock()
	defer pkgLock.RUnlock()
	if theme, ok = styleOverrides[name]; !ok {
		switch name {
		case MonoStyle:
			return defaultMonoStyle, true
		case ColorStyle:
			return defaultColorStyle, true
		case NilStyle:
			return Style{}, true
		}
	}
	return
}

func GetDefaultMonoStyle() (theme Style) {
	theme, _ = GetStyle(MonoStyle)
	return
}

func GetDefaultColorStyle() (theme Style) {
	theme, _ = GetStyle(MonoStyle)
	return
}
