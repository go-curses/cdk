// Copyright (c) 2021-2023  The Go-Curses Authors
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

package paint

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTheme(t *testing.T) {
	Convey("Basic Theme Features", t, func() {
		So(
			GetDefaultMonoTheme().String(),
			ShouldEqual,
			"{Content={Normal={white[#ffffff],black[#000000],0},Selected={white[#ffffff],black[#000000],0},Active={white[#ffffff],black[#000000],4},Prelight={white[#ffffff],black[#000000],0},Insensitive={white[#ffffff],black[#000000],16},FillRune=32,BorderRunes={BorderRunes=9488,9472,9484,9474,9492,9472,9496,9474},ArrowRunes={ArrowRunes=8593,8592,8595,8594},Overlay=false},Border={Normal={white[#ffffff],black[#000000],0},Selected={white[#ffffff],black[#000000],0},Active={white[#ffffff],black[#000000],4},Prelight={white[#ffffff],black[#000000],0},Insensitive={white[#ffffff],black[#000000],16},FillRune=32,BorderRunes={BorderRunes=9488,9472,9484,9474,9492,9472,9496,9474},ArrowRunes={ArrowRunes=8593,8592,8595,8594},Overlay=false}}",
		)
	})
}
