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