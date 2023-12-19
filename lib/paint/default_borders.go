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
