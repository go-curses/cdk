package memphis

import (
	"github.com/go-curses/cdk/lib/paint"
)

type Renderer interface {
	GetContent(x, y int) (mainc rune, combc []rune, style paint.Style, width int)
	SetContent(x int, y int, mainc rune, combc []rune, style paint.Style)
}
