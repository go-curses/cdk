// Copyright 2021  The CDK Authors
// Copyright 2018 The TCell Authors
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

package cdk

import (
	"testing"

	"github.com/go-curses/cdk/lib/paint"
)

func TestInitScreen(t *testing.T) {
	s := NewTestingScreen(t, "")
	defer s.Close()
	if x, y := s.Size(); x != 80 || y != 25 {
		t.Fatalf("Size should be 80, 25, was %v, %v", x, y)
	}
	if s.CharacterSet() != "UTF-8" {
		t.Fatalf("Character Set (%v) not UTF-8", s.CharacterSet())
	}
	if b, x, y := s.GetContents(); len(b) != x*y || x != 80 || y != 25 {
		t.Fatalf("Contents (%v, %v, %v) wrong", len(b), x, y)
	}
}

func TestClearScreen(t *testing.T) {
	s := NewTestingScreen(t, "")
	defer s.Close()
	s.Clear()
	b, x, y := s.GetContents()
	if len(b) != x*y || x != 80 || y != 25 {
		t.Fatalf("Contents (%v, %v, %v) wrong", len(b), x, y)
	}
	for i := 0; i < x*y; i++ {
		if len(b[i].Runes) == 1 && b[i].Runes[0] != ' ' {
			t.Errorf("Incorrect contents at %v: %v", i, b[i].Runes)
		}
		if b[i].Style != paint.StyleDefault {
			t.Errorf("Incorrect style at %v: %v", i, b[i].Style)
		}
	}
}

func TestSetCell(t *testing.T) {
	st := paint.StyleDefault.Background(paint.ColorRed).Blink(true)
	s := NewTestingScreen(t, "")
	defer s.Close()
	s.SetCell(2, 5, st, '@')
	b, _, _ := s.GetContents()
	s.Show()
	if len(b) != 80*25 {
		s.Close()
		t.Fatalf("Wrong content size")
	}
	cell := &b[5*80+2]
	if len(cell.Runes) != 1 || len(cell.Bytes) != 1 ||
		cell.Runes[0] != '@' || cell.Bytes[0] != '@' ||
		cell.Style != st {
		t.Errorf("Incorrect cell content: %v", cell)
	}
}

func TestResize(t *testing.T) {
	st := paint.StyleDefault.Background(paint.ColorYellow).Underline(true)
	s := NewTestingScreen(t, "")
	defer s.Close()
	s.SetCell(2, 5, st, '&')
	b, x, y := s.GetContents()
	s.Show()

	cell := &b[5*80+2]
	if len(cell.Runes) != 1 || len(cell.Bytes) != 1 ||
		cell.Runes[0] != '&' || cell.Bytes[0] != '&' ||
		cell.Style != st {
		t.Errorf("Incorrect cell content: %v", cell)
	}
	s.SetSize(30, 10)
	s.Show()
	b2, x2, y2 := s.GetContents()
	if len(b2) == len(b) || x2 == x || y2 == y {
		t.Errorf("Display parameters should not match")
	}

	cell2 := &b[5*80+2]
	if len(cell2.Runes) != 1 || len(cell2.Bytes) != 1 ||
		cell2.Runes[0] != '&' || cell2.Bytes[0] != '&' ||
		cell2.Style != st {
		t.Errorf("Incorrect cell content after resize: %v", cell2)
	}
}

func TestBeep(t *testing.T) {
	s := NewTestingScreen(t, "")
	defer s.Close()

	b0, x0, y0 := s.GetContents()

	if err := s.Beep(); err != nil {
		t.Errorf("could not beep: %v", err)
	}
	s.Show()

	b1, x1, y1 := s.GetContents()
	if x0 != x1 {
		t.Fatalf("display width changed unexpectedly from %d to %d", x0, x1)
	}
	if y0 != y1 {
		t.Fatalf("display height changed unexpectedly from %d to %d", y0, y1)
	}
	if len(b0) != len(b1) {
		t.Fatalf("display size changed unexpectedly (had %d cells, has %d cells)",
			len(b0), len(b1))
	}
	for i := 0; i < len(b0); i++ {
		cell0 := b0[i]
		cell1 := b1[i]
		if len(cell0.Runes) != len(cell1.Runes) {
			t.Errorf("incorrect cell content: had %d runes, has %d runes",
				len(cell0.Runes), len(cell1.Runes))
			continue
		}
		for j := 0; j < len(cell0.Runes); j++ {
			r0 := cell0.Runes[j]
			r1 := cell1.Runes[j]
			if r0 != r1 {
				t.Errorf("incorrect content: cell %d rune %d changed from %v to %v",
					i, j, r0, r1)
			}
		}
	}
}
