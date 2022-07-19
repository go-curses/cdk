package paint

import (
	"fmt"
)

type ArrowRuneSet struct {
	Up    rune
	Left  rune
	Down  rune
	Right rune
}

func (b ArrowRuneSet) String() string {
	return fmt.Sprintf(
		"{ArrowRunes=%v,%v,%v,%v}",
		b.Up,
		b.Left,
		b.Down,
		b.Right,
	)
}