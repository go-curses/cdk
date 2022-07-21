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

package charset

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCharsetUnix(t *testing.T) {
	Convey("Unix Charset checks", t, func() {
		os.Setenv("LC_ALL", "")
		os.Setenv("LC_CTYPE", "")
		os.Setenv("LANG", "POSIX")
		c := Get()
		So(c, ShouldEqual, "US-ASCII")
		os.Setenv("LC_ALL", "")
		os.Setenv("LC_CTYPE", "")
		os.Setenv("LANG", "this.other@thing")
		c = Get()
		So(c, ShouldEqual, "other")
		os.Setenv("LANG", "en_CA.UTF-8")
		c = Get()
		So(c, ShouldEqual, "UTF-8")
	})
}
