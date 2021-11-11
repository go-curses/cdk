// Copyright 2021  The CDK Authors
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

	gencoding "github.com/gdamore/encoding"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
)

func TestEncoding(t *testing.T) {
	Convey("Encoding Customizations", t, func() {
		enc := GetEncoding("eucjp")
		So(enc, ShouldEqual, nil)
		RegisterEncoding("EUCJP", japanese.EUCJP)
		enc = GetEncoding("eucjp")
		So(enc, ShouldEqual, japanese.EUCJP)
		SetEncodingFallback(EncodingFallbackASCII)
		aenc := GetEncoding("notathing")
		So(aenc, ShouldEqual, gencoding.ASCII)
		SetEncodingFallback(EncodingFallbackUTF8)
		uenc := GetEncoding("notathing")
		So(uenc, ShouldResemble, encoding.Nop)
		list := ListEncodings()
		So(len(list), ShouldEqual, 6)
		So(list, ShouldContain, "eucjp")
		UnregisterEncoding("EUCJP")
		list = ListEncodings()
		So(len(list), ShouldEqual, 5)
	})

}
