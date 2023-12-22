//go:build darwin

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
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func ResolveTTY() (ttyPath string, ttyType TermType, err error) {
	ttyType = InvalidTermType
	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)

	// TODO: figure out how to actually get the current processes controlling terminal
	//       - maybe something to do with using C and checking `proc_bsdinfo->e_tdev`
	//       - for now, use ps to get the tty

	var data []byte
	if data, err = exec.Command("ps", "-p", pidStr, "-o", "tty=").Output(); err != nil {
		return
	}
	ttyName := strings.TrimSpace(string(data))
	ttyPath = "/dev/" + ttyName
	_, _, ttyType, _ = CharDeviceInfo(ttyPath)
	return
}