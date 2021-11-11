// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris zos

// Copyright 2021 The TCell Authors
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
	// "fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-curses/cdk/log"
	"github.com/go-curses/term"
)

// engage is used to place the terminal in raw mode and establish screen size, etc.
// Thing of this is as CDK "engaging" the clutch, as it's going to be driving the
// terminal interface.
func (d *CScreen) engage() error {
	if err := term.RawMode(d.term); err != nil {
		return err
	}
	if w, h, err := d.term.Winsz(); err == nil && w > 0 && h > 0 {
		d.cells.Resize(w, h)
		_ = d.PostEvent(NewEventResize(w, h))
	}
	return nil
}

// disengage is used to release the terminal back to support from the caller.
// Think of this as CDK disengaging the clutch, so that another application
// can take over the terminal interface.  This restores the TTY mode that was
// present when the application was first started.
func (d *CScreen) disengage() {
	if err := d.term.Restore(); err != nil {
		log.ErrorF("error restoring terminal: %v", err)
	}
}

// initialize is used at application startup, and sets up the initial values
// including file descriptors used for terminals and saving the initial state
// so that it can be restored when the application terminates.
func (d *CScreen) initialize() error {
	var err error
	if d.ttyFile != nil {
		if d.term, err = term.Open(d.ttyFile.Name()); err != nil {
			return err
		}
	} else {
		if d.ttyPath == "" {
			d.ttyPath = "/dev/tty"
		}
		if d.term, err = term.Open(d.ttyPath); err != nil {
			return err
		}
	}
	if err = term.RawMode(d.term); err != nil {
		return err
	}
	signal.Notify(d.sigWinch, syscall.SIGWINCH)
	if w, h, e := d.getWinSize(); e == nil && w != 0 && h != 0 {
		d.cells.Resize(w, h)
		_ = d.PostEvent(NewEventResize(w, h))
	}
	return nil
}

// finalize is used to at application shutdown, and restores the terminal
// to it's initial state.  It should not be called more than once.
func (d *CScreen) finalize() {
	signal.Stop(d.sigWinch)
	<-d.inDoneQ
	if d.term != nil {
		if err := term.CBreakMode(d.term); err != nil {
			log.ErrorF("error setting CBreakMode: %v", err)
		}
		if err := d.term.Restore(); err != nil {
			log.ErrorF("error restoring terminal: %v", err)
		}
		if err := d.term.Close(); err != nil {
			log.ErrorF("error closing terminal: %v", err)
		}
	}
}

// getWinSize is called to obtain the terminal dimensions.
func (d *CScreen) getWinSize() (w, h int, err error) {
	w, h, err = d.term.Winsz()
	if err != nil {
		w, h = -1, -1
		return
	}
	if w == 0 {
		colsEnv := os.Getenv("COLUMNS")
		if colsEnv != "" {
			if w, err = strconv.Atoi(colsEnv); err != nil {
				w, h = -1, -1
				return
			}
		} else {
			w = d.ti.Columns
		}
	}
	if h == 0 {
		rowsEnv := os.Getenv("LINES")
		if rowsEnv != "" {
			if h, err = strconv.Atoi(rowsEnv); err != nil {
				w, h = -1, -1
				return
			}
		} else {
			h = d.ti.Lines
		}
	}
	return
}

// Beep emits a beep to the terminal.
func (d *CScreen) Beep() error {
	d.writeString(string(byte(7)))
	return nil
}
