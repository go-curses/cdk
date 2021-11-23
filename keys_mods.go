package cdk

import (
	"fmt"
	"regexp"
	"strings"
)

// Key is a generic value for representing keys, and especially special
// keys (function keys, cursor movement keys, etc.)  For normal keys, like
// ASCII letters, we use KeyRune, and then expect the application to
// inspect the Rune() member of the EventKey.
type Key int16

// This is the list of named keys.  KeyRune is special however, in that it is
// a place holder key indicating that a printable character was sent.  The
// actual value of the rune will be transported in the Rune of the associated
// EventKey.
const (
	KeyRune Key = iota + 256
	KeyUp
	KeyDown
	KeyRight
	KeyLeft
	KeyUpLeft
	KeyUpRight
	KeyDownLeft
	KeyDownRight
	KeyCenter
	KeyPgUp
	KeyPgDn
	KeyHome
	KeyEnd
	KeyInsert
	KeyDelete
	KeyHelp
	KeyExit
	KeyClear
	KeyCancel
	KeyPrint
	KeyPause
	KeyBacktab
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
	KeyF21
	KeyF22
	KeyF23
	KeyF24
	KeyF25
	KeyF26
	KeyF27
	KeyF28
	KeyF29
	KeyF30
	KeyF31
	KeyF32
	KeyF33
	KeyF34
	KeyF35
	KeyF36
	KeyF37
	KeyF38
	KeyF39
	KeyF40
	KeyF41
	KeyF42
	KeyF43
	KeyF44
	KeyF45
	KeyF46
	KeyF47
	KeyF48
	KeyF49
	KeyF50
	KeyF51
	KeyF52
	KeyF53
	KeyF54
	KeyF55
	KeyF56
	KeyF57
	KeyF58
	KeyF59
	KeyF60
	KeyF61
	KeyF62
	KeyF63
	KeyF64
)

const (
	// These key codes are used internally, and will never appear to applications.
	keyPasteStart Key = iota + 16384
	keyPasteEnd
)

// These are the control keys.  Note that they overlap with other keys,
// perhaps.  For example, KeyCtrlH is the same as KeyBackspace.
const (
	KeyCtrlSpace Key = iota
	KeyCtrlA
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlE
	KeyCtrlF
	KeyCtrlG
	KeyCtrlH
	KeyCtrlI
	KeyCtrlJ
	KeyCtrlK
	KeyCtrlL
	KeyCtrlM
	KeyCtrlN
	KeyCtrlO
	KeyCtrlP
	KeyCtrlQ
	KeyCtrlR
	KeyCtrlS
	KeyCtrlT
	KeyCtrlU
	KeyCtrlV
	KeyCtrlW
	KeyCtrlX
	KeyCtrlY
	KeyCtrlZ
	KeyCtrlLeftSq // Escape
	KeyCtrlBackslash
	KeyCtrlRightSq
	KeyCtrlCarat
	KeyCtrlUnderscore
)

var decodedCtrlKeys = map[Key]Key{
	KeyCtrlA: KeySmallA,
	KeyCtrlB: KeySmallB,
	KeyCtrlC: KeySmallC,
	KeyCtrlD: KeySmallD,
	KeyCtrlE: KeySmallE,
	KeyCtrlF: KeySmallF,
	KeyCtrlG: KeySmallG,
	KeyCtrlH: KeySmallH,
	KeyCtrlI: KeySmallI,
	KeyCtrlJ: KeySmallJ,
	KeyCtrlK: KeySmallK,
	KeyCtrlL: KeySmallL,
	KeyCtrlM: KeySmallM,
	KeyCtrlN: KeySmallN,
	KeyCtrlO: KeySmallO,
	KeyCtrlP: KeySmallP,
	KeyCtrlQ: KeySmallQ,
	KeyCtrlR: KeySmallR,
	KeyCtrlS: KeySmallS,
	KeyCtrlT: KeySmallT,
	KeyCtrlU: KeySmallU,
	KeyCtrlV: KeySmallV,
	KeyCtrlW: KeySmallW,
	KeyCtrlX: KeySmallX,
	KeyCtrlY: KeySmallY,
	KeyCtrlZ: KeySmallZ,
}

func DecodeCtrlKey(in Key) (key Key, mods ModMask, ok bool) {
	if key, ok = decodedCtrlKeys[in]; ok {
		mods = ModCtrl
		return
	}
	ok = false
	key = in
	mods = ModNone
	return
}

// Special values - these are fixed in an attempt to make it more likely
// that aliases will encode the same way.

// These are the defined ASCII values for key codes.  They generally match
// with KeyCtrl values.
const (
	KeyNUL Key = iota
	KeySOH
	KeySTX
	KeyETX
	KeyEOT
	KeyENQ
	KeyACK
	KeyBEL
	KeyBS
	KeyTAB
	KeyLF
	KeyVT
	KeyFF
	KeyCR
	KeySO
	KeySI
	KeyDLE
	KeyDC1
	KeyDC2
	KeyDC3
	KeyDC4
	KeyNAK
	KeySYN
	KeyETB
	KeyCAN
	KeyEM
	KeySUB
	KeyESC
	KeyFS
	KeyGS
	KeyRS
	KeyUS
	KeyDEL Key = 0x7F
)

// These keys are aliases for other names.
const (
	KeyBackspace  = KeyBS
	KeyTab        = KeyTAB
	KeyEsc        = KeyESC
	KeyEscape     = KeyESC
	KeyEnter      = KeyCR
	KeyBackspace2 = KeyDEL
)

// ASCII Keys
const (
	KeySpacebar           Key = KeyRune
	KeySpace              Key = 32
	KeyExclamationMark    Key = 33
	KeyDoubleQuote        Key = 34
	KeyNumber             Key = 35
	KeyDollarSign         Key = 36
	KeyPercent            Key = 37
	KeyAmpersand          Key = 38
	KeySingleQuote        Key = 39
	KeyLeftParenthesis    Key = 40
	KeyRightParenthesis   Key = 41
	KeyAsterisk           Key = 42
	KeyPlus               Key = 43
	KeyComma              Key = 44
	KeyMinus              Key = 45
	KeyPeriod             Key = 46
	KeySlash              Key = 47
	KeyZero               Key = 48
	KeyOne                Key = 49
	KeyTwo                Key = 50
	KeyThree              Key = 51
	KeyFour               Key = 52
	KeyFive               Key = 53
	KeySix                Key = 54
	KeySeven              Key = 55
	KeyEight              Key = 56
	KeyNine               Key = 57
	KeyColon              Key = 58
	KeySemicolon          Key = 59
	KeyLessThan           Key = 60
	KeyEqualitySign       Key = 61
	KeyGreaterThan        Key = 62
	KeyQuestionMark       Key = 63
	KeyAtSign             Key = 64
	KeyCapitalA           Key = 65
	KeyCapitalB           Key = 66
	KeyCapitalC           Key = 67
	KeyCapitalD           Key = 68
	KeyCapitalE           Key = 69
	KeyCapitalF           Key = 70
	KeyCapitalG           Key = 71
	KeyCapitalH           Key = 72
	KeyCapitalI           Key = 73
	KeyCapitalJ           Key = 74
	KeyCapitalK           Key = 75
	KeyCapitalL           Key = 76
	KeyCapitalM           Key = 77
	KeyCapitalN           Key = 78
	KeyCapitalO           Key = 79
	KeyCapitalP           Key = 80
	KeyCapitalQ           Key = 81
	KeyCapitalR           Key = 82
	KeyCapitalS           Key = 83
	KeyCapitalT           Key = 84
	KeyCapitalU           Key = 85
	KeyCapitalV           Key = 86
	KeyCapitalW           Key = 87
	KeyCapitalX           Key = 88
	KeyCapitalY           Key = 89
	KeyCapitalZ           Key = 90
	KeyLeftSquareBracket  Key = 91
	KeyBackslash          Key = 92
	KeyRightSquareBracket Key = 93
	KeyCaretCircumflex    Key = 94
	KeyUnderscore         Key = 95
	KeyGraveAccent        Key = 96
	KeySmallA             Key = 97
	KeySmallB             Key = 98
	KeySmallC             Key = 99
	KeySmallD             Key = 100
	KeySmallE             Key = 101
	KeySmallF             Key = 102
	KeySmallG             Key = 103
	KeySmallH             Key = 104
	KeySmallI             Key = 105
	KeySmallJ             Key = 106
	KeySmallK             Key = 107
	KeySmallL             Key = 108
	KeySmallM             Key = 109
	KeySmallN             Key = 110
	KeySmallO             Key = 111
	KeySmallP             Key = 112
	KeySmallQ             Key = 113
	KeySmallR             Key = 114
	KeySmallS             Key = 115
	KeySmallT             Key = 116
	KeySmallU             Key = 117
	KeySmallV             Key = 118
	KeySmallW             Key = 119
	KeySmallX             Key = 120
	KeySmallY             Key = 121
	KeySmallZ             Key = 122
	KeyLeftCurlyBracket   Key = 123
	KeyVerticalBar        Key = 124
	KeyRightCurlyBracket  Key = 125
	KeyTilde              Key = 126
)

func LookupKeyRune(r rune) Key {
	if r >= 32 && r <= 126 {
		return Key(r)
	}
	return 0
}

// KeyNames holds the written names of special keys. Useful to echo back a key
// name, or to look up a key from a string value.
var KeyNames = map[Key]string{
	KeyEnter:          "Enter",
	KeyBackspace:      "Backspace",
	KeyTab:            "Tab",
	KeyBacktab:        "Backtab",
	KeyEsc:            "Esc",
	KeyBackspace2:     "Backspace2",
	KeyDelete:         "Delete",
	KeyInsert:         "Insert",
	KeyUp:             "Up",
	KeyDown:           "Down",
	KeyLeft:           "Left",
	KeyRight:          "Right",
	KeyHome:           "Home",
	KeyEnd:            "End",
	KeyUpLeft:         "UpLeft",
	KeyUpRight:        "UpRight",
	KeyDownLeft:       "DownLeft",
	KeyDownRight:      "DownRight",
	KeyCenter:         "Center",
	KeyPgDn:           "PgDn",
	KeyPgUp:           "PgUp",
	KeyClear:          "Clear",
	KeyExit:           "Exit",
	KeyCancel:         "Cancel",
	KeyPause:          "Pause",
	KeyPrint:          "Print",
	KeyF1:             "F1",
	KeyF2:             "F2",
	KeyF3:             "F3",
	KeyF4:             "F4",
	KeyF5:             "F5",
	KeyF6:             "F6",
	KeyF7:             "F7",
	KeyF8:             "F8",
	KeyF9:             "F9",
	KeyF10:            "F10",
	KeyF11:            "F11",
	KeyF12:            "F12",
	KeyF13:            "F13",
	KeyF14:            "F14",
	KeyF15:            "F15",
	KeyF16:            "F16",
	KeyF17:            "F17",
	KeyF18:            "F18",
	KeyF19:            "F19",
	KeyF20:            "F20",
	KeyF21:            "F21",
	KeyF22:            "F22",
	KeyF23:            "F23",
	KeyF24:            "F24",
	KeyF25:            "F25",
	KeyF26:            "F26",
	KeyF27:            "F27",
	KeyF28:            "F28",
	KeyF29:            "F29",
	KeyF30:            "F30",
	KeyF31:            "F31",
	KeyF32:            "F32",
	KeyF33:            "F33",
	KeyF34:            "F34",
	KeyF35:            "F35",
	KeyF36:            "F36",
	KeyF37:            "F37",
	KeyF38:            "F38",
	KeyF39:            "F39",
	KeyF40:            "F40",
	KeyF41:            "F41",
	KeyF42:            "F42",
	KeyF43:            "F43",
	KeyF44:            "F44",
	KeyF45:            "F45",
	KeyF46:            "F46",
	KeyF47:            "F47",
	KeyF48:            "F48",
	KeyF49:            "F49",
	KeyF50:            "F50",
	KeyF51:            "F51",
	KeyF52:            "F52",
	KeyF53:            "F53",
	KeyF54:            "F54",
	KeyF55:            "F55",
	KeyF56:            "F56",
	KeyF57:            "F57",
	KeyF58:            "F58",
	KeyF59:            "F59",
	KeyF60:            "F60",
	KeyF61:            "F61",
	KeyF62:            "F62",
	KeyF63:            "F63",
	KeyF64:            "F64",
	KeyCtrlA:          "Ctrl-A",
	KeyCtrlB:          "Ctrl-B",
	KeyCtrlC:          "Ctrl-C",
	KeyCtrlD:          "Ctrl-D",
	KeyCtrlE:          "Ctrl-E",
	KeyCtrlF:          "Ctrl-F",
	KeyCtrlG:          "Ctrl-G",
	KeyCtrlJ:          "Ctrl-J",
	KeyCtrlK:          "Ctrl-K",
	KeyCtrlL:          "Ctrl-L",
	KeyCtrlN:          "Ctrl-N",
	KeyCtrlO:          "Ctrl-O",
	KeyCtrlP:          "Ctrl-P",
	KeyCtrlQ:          "Ctrl-Q",
	KeyCtrlR:          "Ctrl-R",
	KeyCtrlS:          "Ctrl-S",
	KeyCtrlT:          "Ctrl-T",
	KeyCtrlU:          "Ctrl-U",
	KeyCtrlV:          "Ctrl-V",
	KeyCtrlW:          "Ctrl-W",
	KeyCtrlX:          "Ctrl-X",
	KeyCtrlY:          "Ctrl-Y",
	KeyCtrlZ:          "Ctrl-Z",
	KeyCtrlSpace:      "Ctrl-Space",
	KeyCtrlUnderscore: "Ctrl-_",
	KeyCtrlRightSq:    "Ctrl-]",
	KeyCtrlBackslash:  "Ctrl-\\",
	KeyCtrlCarat:      "Ctrl-^",
}

func LookupKeyName(key Key) string {
	if s, ok := KeyNames[key]; ok {
		return s
	}
	return fmt.Sprintf("%v", key)
}

// ModMask is a mask of modifier keys.  Note that it will not always be
// possible to report modifier keys.
type ModMask int16

// These are the modifiers keys that can be sent either with a key press,
// or a mouse event.  Note that as of now, due to the confusion associated
// with Meta, and the lack of support for it on many/most platforms, the
// current implementations never use it.  Instead, they use ModAlt, even for
// events that could possibly have been distinguished from ModAlt.
const (
	ModShift ModMask = 1 << iota
	ModCtrl
	ModAlt
	ModMeta
	ModNone ModMask = 0
)

func (m ModMask) Has(mask ModMask) bool {
	return m&mask != 0
}

func (m ModMask) String() string {
	v := ""
	if m.Has(ModCtrl) {
		v += "<Control>"
	}
	if m.Has(ModAlt) {
		v += "<Alt>"
	}
	if m.Has(ModMeta) {
		v += "<Meta>"
	}
	if m.Has(ModShift) {
		v += "<Shift>"
	}
	return v
}

var rxParseKeyMods = regexp.MustCompile(`^\s*((?:\s*<[a-zA-Z][a-zA-Z0-9]+>\s*)*[a-zA-Z0-9])\s*$`)
var rxParseMods = regexp.MustCompile(`\s*<([a-zA-Z][a-zA-Z0-9]+)>\s*`)

func ParseKeyMods(input string) (key Key, mods ModMask, err error) {
	if rxParseKeyMods.MatchString(input) {
		match := rxParseKeyMods.FindAllString(input, -1)
		remainder := strings.TrimSpace(match[0])
		if rxParseMods.MatchString(match[0]) {
			remainder = rxParseMods.ReplaceAllString(remainder, "")
			matches := rxParseMods.FindAllStringSubmatch(match[0], -1)
			if len(matches) == 1 {
				for i := 1; i < len(matches[0]); i++ {
					switch strings.ToLower(matches[0][i]) {
					case "control", "ctrl", "ctl":
						mods |= ModCtrl
					case "alternate", "alt":
						mods |= ModAlt
					case "meta":
						mods |= ModMeta
					case "shift":
						mods |= ModShift
					default:
						key = KeyNUL
						mods = ModNone
						err = fmt.Errorf("error parsing modifier: %q", matches[0][i])
						return
					}
				}
			}
		}
		if remainder = strings.TrimSpace(remainder); remainder == "" {
			key = KeyNUL
			mods = ModNone
			err = fmt.Errorf("error parsing key: %q", match[0])
			return
		}
		key = LookupKeyRune(rune(remainder[0]))
		return
	}
	key = KeyNUL
	mods = ModNone
	err = fmt.Errorf("error parsing string: %q", input)
	return
}
