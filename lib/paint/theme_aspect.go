package paint

import (
	"fmt"
)

type ThemeAspect struct {
	Normal      Style
	Selected    Style
	Active      Style
	Prelight    Style
	Insensitive Style
	FillRune    rune
	BorderRunes BorderRuneSet
	ArrowRunes  ArrowRuneSet
	Overlay     bool // keep existing background
}

func (t ThemeAspect) String() string {
	return fmt.Sprintf(
		"{Normal=%v,Selected=%v,Active=%v,Prelight=%v,Insensitive=%v,FillRune=%v,BorderRunes=%v,ArrowRunes=%v,Overlay=%v}",
		t.Normal,
		t.Selected,
		t.Active,
		t.Prelight,
		t.Insensitive,
		t.FillRune,
		t.BorderRunes,
		t.ArrowRunes,
		t.Overlay,
	)
}