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