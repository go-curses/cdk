//go:build linux

// Copyright (c) 2023  The Go-Curses Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package term

import (
	"golang.org/x/exp/constraints"

	"github.com/go-curses/cdk/lib/math"
)

func ParseDeviceInfo[T constraints.Integer](dev T) (major, minor uint64, ttyType TermType) {
	major = DeviceMajor(dev)
	minor = DeviceMinor(dev)
	switch {
	case math.InRange(major, 136, 143):
		ttyType = PseudoTTY
	case math.InRange(major, 2, 5) || math.IsOneOf(major, 166, 188):
		ttyType = ConsoleTTY
	default:
		ttyType = NotATTY
	}
	return
}