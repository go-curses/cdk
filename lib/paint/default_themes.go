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
