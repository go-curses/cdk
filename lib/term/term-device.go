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
	"syscall"

	"golang.org/x/exp/constraints"
)

func DeviceMajor[T constraints.Integer](rdev T) uint64 {
	return uint64(rdev) >> 8
}

func DeviceMinor[T constraints.Integer](rdev T) uint64 {
	return uint64(rdev) & 0xff
}

// CharDeviceInfo uses syscall.Stat to test if the path exists, is a character device (mode has S_IFCHR), parses the
// major and minor numbers out of syscall.Stat_t.Rdev and finally determines the TermType value based on the major and
// minor values
func CharDeviceInfo(path string) (major, minor uint64, ttyType TermType, yes bool) {
	stat := syscall.Stat_t{}
	if err := syscall.Stat(path, &stat); err == nil {
		if yes = stat.Mode&(syscall.S_IFCHR) == syscall.S_IFCHR; yes {
			major, minor, ttyType = ParseDeviceInfo(stat.Rdev)
		}
	}
	return
}