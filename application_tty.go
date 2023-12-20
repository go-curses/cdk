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

package cdk

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-curses/cdk/lib/paths"
	"github.com/go-curses/cdk/log"
)

func (app *CApplication) ttyPathSetup(ttyPath string) {
	if ttyPath != "" && ttyPath != "auto" {
		// developer specified something specific
		app.ttyPath = ttyPath
		log.InfoDF(1, "using ttyPath: %q", app.ttyPath)
		return
	}

	// developer omitted anything specific (empty or auto ttyPath)
	app.ttyPath = "/dev/tty"

	// let's auto-detect if we can
	if ttyCmd, err := exec.LookPath("tty"); err == nil {
		if ttyCmd, err = filepath.Abs(ttyCmd); err == nil {
			cmd := exec.Command(ttyCmd)
			cmd.Stdin = os.Stdin
			cmd.Dir, _ = os.Getwd()
			cmd.Env = os.Environ()

			var sob, seb bytes.Buffer
			cmd.Stdout = &sob
			cmd.Stderr = &seb

			if err = cmd.Run(); err != nil {
				format := "error running %q: %v"
				argv := []interface{}{
					ttyCmd,
					err,
				}
				stdout := sob.String()
				if stdout = strings.TrimSpace(stdout); stdout != "" {
					format += "\n[stdout]\n%v\n[/stdout]"
					argv = append(argv, stdout)
				}
				stderr := sob.String()
				if stderr = strings.TrimSpace(stderr); stderr != "" {
					format += "\n[stderr]\n%v\n[/stderr]"
					argv = append(argv, stderr)
				}
				log.ErrorF(format, argv...)
			} else {
				output := sob.String()
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if path := strings.TrimSpace(line); paths.IsDevice(path) {
						app.ttyPath = path
						return
					}
				}
			}
		}
	}

	log.InfoDF(1, "using ttyPath (auto): %q", app.ttyPath)
	return
}