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

package flag_sorter

import (
	"strings"

	"github.com/urfave/cli/v2"

	csort "github.com/go-curses/cdk/lib/sort"
)

type FlagSorter []cli.Flag

func (f FlagSorter) Len() int {
	return len(f)
}

func (f FlagSorter) Less(i, j int) bool {
	iIsCdk := strings.HasPrefix(f[i].Names()[0], "cdk-")
	jIsCdk := strings.HasPrefix(f[j].Names()[0], "cdk-")
	if iIsCdk {
		if jIsCdk {
			return csort.LexicographicLess(f[i].Names()[0], f[j].Names()[0])
		}
		return false
	} else if jIsCdk {
		return true
	}
	return !csort.LexicographicLess(f[i].Names()[0], f[j].Names()[0])
}

func (f FlagSorter) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
