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

package paint

import (
	"fmt"
)

type Theme struct {
	Content ThemeAspect
	Border  ThemeAspect
}

func (t Theme) String() string {
	return fmt.Sprintf(
		"{Content=%v,Border=%v}",
		t.Content,
		t.Border,
	)
}

func (t Theme) Clone() Theme {
	return Theme{
		Content: t.Content,
		Border:  t.Border,
	}
}
