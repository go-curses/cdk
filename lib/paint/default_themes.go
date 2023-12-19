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

type ThemeName string

const (
	NilTheme     ThemeName = "nil"
	MonoTheme    ThemeName = "mono"
	ColorTheme   ThemeName = "color"
	DisplayTheme ThemeName = "display"
)

var themeOverrides = map[ThemeName]Theme{}

func RegisterTheme(name ThemeName, theme Theme) {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	themeOverrides[name] = theme
}

func GetTheme(name ThemeName) (theme Theme, ok bool) {
	pkgLock.RLock()
	defer pkgLock.RUnlock()
	if theme, ok = themeOverrides[name]; !ok {
		switch name {
		case MonoTheme:
			return defaultMonoTheme, true
		case ColorTheme:
			return defaultColorTheme, true
		case DisplayTheme:
			return defaultDisplayTheme, true
		case NilTheme:
			return Theme{}, true
		}
	}
	return
}

func GetDefaultMonoTheme() (theme Theme) {
	theme, _ = GetTheme(MonoTheme)
	return
}

func GetDefaultColorTheme() (theme Theme) {
	theme, _ = GetTheme(ColorTheme)
	return
}

func MakeStyledColorFillTheme(style Style) (theme Theme) {
	theme = GetDefaultColorTheme()
	theme.Content.Normal = style
	theme.Content.Active = style
	theme.Content.Prelight = style
	theme.Content.Selected = style
	theme.Content.Insensitive = style
	theme.Content.FillRune = DefaultNilRune
	return
}
