package paint

const (
	DefaultFillRune = ' '
	DefaultNilRune  = rune(0)
)

var (
	defaultMonoStyle  = StyleDefault.Foreground(ColorWhite).Background(ColorBlack).Dim(false)
	defaultColorStyle = StyleDefault.Foreground(ColorWhite).Background(ColorNavy).Dim(false)
)

var (
	stockBorderRune = BorderRuneSet{
		TopLeft:     RuneULCorner,
		Top:         RuneHLine,
		TopRight:    RuneURCorner,
		Left:        RuneVLine,
		Right:       RuneVLine,
		BottomLeft:  RuneLLCorner,
		Bottom:      RuneHLine,
		BottomRight: RuneLRCorner,
	}
	roundedBorderRune = BorderRuneSet{
		TopLeft:     RuneULCornerRounded,
		Top:         RuneHLine,
		TopRight:    RuneURCornerRounded,
		Left:        RuneVLine,
		Right:       RuneVLine,
		BottomLeft:  RuneLLCornerRounded,
		Bottom:      RuneHLine,
		BottomRight: RuneLRCornerRounded,
	}
	doubleBorderRune = BorderRuneSet{
		TopLeft:     RuneBoxDrawingsDoubleDownAndRight,
		Top:         RuneBoxDrawingsDoubleHorizontal,
		TopRight:    RuneBoxDrawingsDoubleDownAndLeft,
		Left:        RuneBoxDrawingsDoubleVertical,
		Right:       RuneBoxDrawingsDoubleVertical,
		BottomLeft:  RuneBoxDrawingsDoubleUpAndRight,
		Bottom:      RuneBoxDrawingsDoubleHorizontal,
		BottomRight: RuneBoxDrawingsDoubleUpAndLeft,
	}
	emptyBorderRune = BorderRuneSet{
		TopLeft:     ' ',
		Top:         ' ',
		TopRight:    ' ',
		Left:        ' ',
		Right:       ' ',
		BottomLeft:  ' ',
		Bottom:      ' ',
		BottomRight: ' ',
	}
	nilBorderRune = BorderRuneSet{
		TopLeft:     DefaultNilRune,
		Top:         DefaultNilRune,
		TopRight:    DefaultNilRune,
		Left:        DefaultNilRune,
		Right:       DefaultNilRune,
		BottomLeft:  DefaultNilRune,
		Bottom:      DefaultNilRune,
		BottomRight: DefaultNilRune,
	}

	stockArrowRune = ArrowRuneSet{
		Up:    RuneUArrow,
		Left:  RuneLArrow,
		Down:  RuneDArrow,
		Right: RuneRArrow,
	}
	wideArrowRune = ArrowRuneSet{
		Up:    RuneTriangleUp,
		Left:  RuneTriangleLeft,
		Down:  RuneTriangleDown,
		Right: RuneTriangleRight,
	}
)

var (
	defaultMonoThemeAspect = ThemeAspect{
		Normal:      defaultMonoStyle,
		Selected:    defaultMonoStyle.Dim(false),
		Active:      defaultMonoStyle.Dim(false).Reverse(true),
		Prelight:    defaultMonoStyle.Dim(false),
		Insensitive: defaultMonoStyle.Dim(true),
		FillRune:    DefaultFillRune,
		BorderRunes: stockBorderRune,
		ArrowRunes:  stockArrowRune,
		Overlay:     false,
	}
	defaultColorThemeAspect = ThemeAspect{
		Normal:      defaultColorStyle,
		Selected:    defaultColorStyle.Dim(false),
		Active:      defaultColorStyle.Dim(false).Reverse(true),
		Prelight:    defaultColorStyle.Dim(false),
		Insensitive: defaultColorStyle.Dim(true),
		FillRune:    DefaultFillRune,
		BorderRunes: stockBorderRune,
		ArrowRunes:  stockArrowRune,
		Overlay:     false,
	}
	defaultDisplayThemeAspect = ThemeAspect{
		Normal:      defaultMonoStyle.Dim(true),
		Selected:    defaultMonoStyle.Dim(false),
		Active:      defaultMonoStyle.Dim(false).Reverse(true),
		Prelight:    defaultMonoStyle.Dim(false),
		Insensitive: defaultMonoStyle.Dim(true),
		FillRune:    RuneBoxDrawingsLightDiagonalCross,
		BorderRunes: stockBorderRune,
		ArrowRunes:  stockArrowRune,
		Overlay:     false,
	}

	defaultMonoTheme = Theme{
		Content: defaultMonoThemeAspect,
		Border:  defaultMonoThemeAspect,
	}
	defaultColorTheme = Theme{
		Content: defaultColorThemeAspect,
		Border:  defaultColorThemeAspect,
	}
	defaultDisplayTheme = Theme{
		Content: defaultDisplayThemeAspect,
		Border:  defaultDisplayThemeAspect,
	}
)
