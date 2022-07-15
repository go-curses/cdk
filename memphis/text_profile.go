package memphis

import (
	"strings"

	"github.com/go-curses/cdk/lib/math"
	"github.com/go-curses/cdk/lib/ptypes"
	"github.com/go-curses/cdk/lib/sync"
)

type TextProfile struct {
	text    string
	textLen int

	data           [][]rune
	lookupPosition [][]int
	lookupPoint    []ptypes.Point2I

	width  int
	height int

	lock *sync.RWMutex
}

func NewTextProfile(input string) (tp *TextProfile) {
	tp = new(TextProfile)
	tp.init()
	tp.Set(input)
	return
}

func (tp *TextProfile) init() {
	tp.text = ""
	tp.textLen = 0
	tp.data = make([][]rune, 0)
	tp.lookupPosition = make([][]int, 0)
	tp.lookupPoint = make([]ptypes.Point2I, 0)
	tp.width = 0
	tp.height = 0
	tp.lock = &sync.RWMutex{}
}

func (tp *TextProfile) Len() int {
	tp.lock.RLock()
	defer tp.lock.RUnlock()
	return tp.textLen
}

func (tp *TextProfile) Width() int {
	tp.lock.RLock()
	defer tp.lock.RUnlock()
	return tp.width
}

func (tp *TextProfile) Height() int {
	tp.lock.RLock()
	defer tp.lock.RUnlock()
	return tp.height
}

func (tp *TextProfile) Get() string {
	tp.lock.RLock()
	defer tp.lock.RUnlock()
	return tp.text
}

func (tp *TextProfile) EndsWithNewLine() (ok bool) {
	tp.lock.RLock()
	defer tp.lock.RUnlock()
	ok = tp.endsWithNewLine()
	return
}

func (tp *TextProfile) endsWithNewLine() (ok bool) {
	switch rune(tp.text[tp.textLen-1]) {
	case 10, 13:
		ok = true
	}
	return
}

func (tp *TextProfile) Set(text string) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	tp.text = strings.ReplaceAll(text, "\r", "")
	tp.textLen = len(text)

	tp.data = make([][]rune, 0)
	tp.lookupPosition = make([][]int, 0)
	tp.lookupPoint = make([]ptypes.Point2I, 0)

	point := ptypes.NewPoint2I(0, 0)
	for idx, character := range text {

		if point.Y >= len(tp.lookupPosition) {
			tp.lookupPosition = append(tp.lookupPosition, []int{})
		}
		if point.Y >= len(tp.data) {
			tp.data = append(tp.data, []rune{})
		}

		// positions
		tp.lookupPosition[point.Y] = append(tp.lookupPosition[point.Y], idx)
		point.X = math.FloorI(len(tp.lookupPosition[point.Y])-1, 0)
		// points
		tp.lookupPoint = append(tp.lookupPoint, ptypes.MakePoint2I(point.X, point.Y))
		// crop data
		tp.data[point.Y] = append(tp.data[point.Y], character)

		switch character {
		case 10, 13:
			// newline, wrap point to starto of next line
			tp.lookupPosition = append(tp.lookupPosition, []int{})
			point.Y = math.FloorI(len(tp.lookupPosition)-1, 0)
			point.X = 0
		}
	}
}

func (tp *TextProfile) GetPointFromPosition(position int) (point ptypes.Point2I) {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	if tp.textLen == 0 {
		point = ptypes.MakePoint2I(0, 0)
		return
	}

	if position > tp.textLen {
		position = tp.textLen - 1
	}

	if position == tp.textLen {
		point = tp.lookupPoint[position-1].Clone()
		if tp.endsWithNewLine() {
			point.Y += 1
			point.X = 0
		} else {
			point.X += 1
		}
	} else {
		point = tp.lookupPoint[position].Clone()
	}
	return
}

func (tp *TextProfile) GetPositionFromPoint(point ptypes.Point2I) (position int) {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	if lookupLength := len(tp.lookupPosition); lookupLength > 0 {

		if point.Y < 0 || point.Y >= lookupLength {
			point.Y = lookupLength - 1
		}

		if xLookupLength := len(tp.lookupPosition[point.Y]); xLookupLength > 0 {
			if point.X < 0 || point.X >= xLookupLength {
				point.X = xLookupLength - 1

				switch tp.data[point.Y][point.X] {
				case 10, 13:
				default:
					position = tp.lookupPosition[point.Y][point.X] + 1
					return
				}
			}

			position = tp.lookupPosition[point.Y][point.X]
		} else {
			position = tp.textLen
		}
	}
	return
}

func (tp *TextProfile) Crop(region ptypes.Region) (cropped string) {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	if tp.textLen == 0 {
		return ""
	}

	far := region.FarPoint()

	for y, line := range tp.data {
		if y >= region.Y && y <= far.Y {
			var output string
			lineLength := len(line)

			if lineLength <= region.X {
				output = "\n"
			} else {
				if lineLength > far.X {
					output = string(line[region.X:far.X])
				} else {
					output = string(line[region.X:])
				}

				if outputLen := len(output); outputLen > 0 {
					if output[outputLen-1] != '\n' {
						output += "\n"
					}
				}
			}

			cropped += output
		}
	}
	return
}

func (tp *TextProfile) Select(start, end int) (selected string) {
	tp.lock.RLock()
	defer tp.lock.RUnlock()
	if tp.textLen == 0 {
		return
	}

	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = tp.textLen - 1
	} else {
		end += 1
	}

	if start < tp.textLen {
		if end != start && end > start && end < tp.textLen {
			selected = tp.text[start:end]
		} else if end == start {
			selected = string(tp.text[start])
		} else {
			selected = tp.text[start:]
		}
	}
	return
}

func (tp *TextProfile) Insert(text string, position int) (modified string, ok bool) {
	tp.lock.RLock()
	if position < 0 || position >= tp.textLen {
		modified = tp.text + text
	} else {
		modified = tp.text[:position] + text + tp.text[position:]
	}
	tp.lock.RUnlock()
	tp.Set(modified)
	ok = true
	return
}

func (tp *TextProfile) Delete(startPos int, endPos int) (modified string, ok bool) {
	tp.lock.RLock()
	if startPos >= tp.textLen {
		tp.lock.RUnlock()
		modified = tp.text
		return
	}
	if endPos >= tp.textLen {
		modified = tp.text[:startPos]
	} else {
		modified = tp.text[:startPos] + tp.text[endPos+1:]
	}
	tp.lock.RUnlock()
	tp.Set(modified)
	ok = true
	return
}

func (tp *TextProfile) Overwrite(text string, startPos, endPos int) (modified string, ok bool) {
	tp.lock.RLock()
	if startPos < 0 {
		startPos = tp.textLen
	}
	if startPos >= tp.textLen {
		modified = tp.text + text
	} else if endPos < 0 || endPos >= tp.textLen {
		modified = tp.text[:startPos] + text
	} else {
		modified = tp.text[:startPos] + text + tp.text[endPos+1:]
	}
	tp.lock.RUnlock()
	tp.Set(modified)
	ok = true
	return
}