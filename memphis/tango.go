// Copyright 2022  The CDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memphis

// Text-based version of Pango

/*
<span
  style=[normal,italic]
  weight=[dim,normal,bold]
  foreground=[color]
  background=[color]
  underline=[bool]
  strikethrough=[bool]
>
 CONTENT
</span>
<b></b>
<i></i>
<s></s>
<u></u>
<d></d>

*/

import (
	"encoding/xml"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/log"
)

type Tango interface {
	Raw() string
	TextBuffer(mnemonic bool) TextBuffer
}

type CTango struct {
	raw    string
	style  paint.Style
	marked []TextCell
	input  WordLine
}

func NewMarkup(text string, style paint.Style) (markup Tango, err error) {
	if !strings.HasPrefix(text, "<markup") {
		text = "<markup>" + text + "</markup>"
	}
	m := &CTango{
		raw:   text,
		style: style,
	}
	if err = m.init(); err == nil {
		markup = m
	} else {
		markup = nil
	}
	return
}

func (m *CTango) Raw() string {
	return m.raw
}

func (m *CTango) TextBuffer(mnemonic bool) TextBuffer {
	tb := NewEmptyTextBuffer(m.style, mnemonic)
	tb.SetInput(m.input)
	return tb
}

func (m *CTango) init() error {
	m.marked = []TextCell{}
	m.input = NewEmptyWordLine()
	r := strings.NewReader(m.raw)
	parser := xml.NewDecoder(r)

	wid := 0

	cStyle := m.style         // current style
	var pStyles []paint.Style // previous styles

	isWord := false
	var err error
	var token xml.Token
	for {
		token, err = parser.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "markup":
				pStyles = append(pStyles, cStyle)
				cStyle = m.parseStyleAttrs(t.Attr)
			case "span":
				pStyles = append(pStyles, cStyle)
				cStyle = m.parseStyleAttrs(t.Attr)
			case "b":
				cStyle = cStyle.Bold(true)
			case "i":
				cStyle = cStyle.Italic(true)
			case "s":
				cStyle = cStyle.Strike(true)
			case "u":
				cStyle = cStyle.Underline(true)
			case "d":
				cStyle = cStyle.Dim(true)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "markup":
				last := len(pStyles) - 1
				cStyle = pStyles[last]
				pStyles = pStyles[:last]
			case "span":
				last := len(pStyles) - 1
				cStyle = pStyles[last]
				pStyles = pStyles[:last]
			case "b":
				cStyle = cStyle.Bold(false)
			case "i":
				cStyle = cStyle.Italic(false)
			case "s":
				cStyle = cStyle.Strike(false)
			case "u":
				cStyle = cStyle.Underline(false)
			case "d":
				cStyle = cStyle.Dim(false)
			}
		case xml.CharData:
			for idx := 0; idx < len(t); idx++ {
				v, _ := utf8.DecodeRune(t[idx:])
				m.marked = append(m.marked, NewTextCellFromRune(v, cStyle))
				if unicode.IsSpace(v) {
					if isWord {
						isWord = false
						m.input.AppendWordCell(NewEmptyWordCell())
						wid = m.input.Len() - 1
					}
				} else {
					if !isWord {
						isWord = true
						m.input.AppendWordCell(NewEmptyWordCell())
						wid = m.input.Len() - 1
					}
				}
				if wid >= m.input.Len() {
					m.input.AppendWordCell(NewEmptyWordCell())
				}
				if err := m.input.AppendWordRune(wid, v, cStyle); err != nil {
					log.ErrorDF(1, "error appending word rune: %v", err)
				}
			} // for idx len(content)
		case xml.Comment:
		case xml.ProcInst:
		case xml.Directive:
		default:
		}
	}
	return nil
}

func (m *CTango) parseStyleAttrs(attrs []xml.Attr) (style paint.Style) {
	style = m.style
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "style":
			switch attr.Value {
			case "normal":
				style = style.Italic(false)
			case "italic":
				style = style.Italic(true)
			}
		case "weight":
			switch attr.Value {
			case "dim":
				style = style.Dim(true)
			case "normal":
				style = style.Dim(false).Bold(false)
			case "bold":
				style = style.Bold(true)
			}
		case "foreground":
			style = style.Foreground(paint.GetColor(attr.Value))
		case "background":
			style = style.Background(paint.GetColor(attr.Value))
		case "underline":
			style = style.Underline(attr.Value == "true" || attr.Value == "1")
		case "strikethrough":
			style = style.Strike(attr.Value == "true" || attr.Value == "1")
		}
	}
	return
}
