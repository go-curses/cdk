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
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func ResolveTTY() (ttyPath string, ttyType TermType, err error) {
	ttyType = InvalidTermType

	procPath := fmt.Sprintf("/proc/%d/stat", os.Getpid())
	if data, ee := os.ReadFile(procPath); ee == nil {
		parts := strings.Split(string(data), " ")
		tty_nr, _ := strconv.Atoi(parts[6])

		var major, minor uint64
		major, minor, ttyType = ParseDeviceInfo(tty_nr)

		switch ttyType {
		case PseudoTTY:
			ttyPath = fmt.Sprintf("/dev/pts/%d", minor)
			return

		case ConsoleTTY:
			sysDevPath := fmt.Sprintf("/sys/dev/char/%d:%d/uevent", major, minor)
			if data, ee = os.ReadFile(sysDevPath); ee == nil {
				lines := strings.Split(string(data), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "DEVNAME=") {
						trimmed := strings.TrimSpace(strings.TrimPrefix(line, "DEVNAME="))
						ttyPath = "/dev/" + trimmed
						return
					}
				}
			}
		}

	}

	err = syscall.ENOTTY
	return
}