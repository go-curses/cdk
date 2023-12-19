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

package exec

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/go-curses/cdk/log"
	cterm "github.com/go-curses/term"
	"golang.org/x/term"
)

type Callback = func(in, out *os.File) (err error)

// CallWithTty will wrap the given *os.File with new pty/tty instances and call
// the fn Callback with the appropriate input and output *os.File handles
func CallWithTty(callTty *os.File, fn Callback) (err error) {
	log.DebugF("callTty = %v", callTty.Name())

	var ptmx, ptty *os.File
	if ptmx, ptty, err = pty.Open(); err != nil {
		return fmt.Errorf("pty.Open error: %v", err)
	}

	resize := make(chan os.Signal, 1)
	signal.Notify(resize, syscall.SIGWINCH)
	Go(func() {
		for range resize {
			if err := pty.InheritSize(callTty, ptmx); err != nil {
				log.ErrorF("error propagating resize %v->%v: %s", callTty.Name(), ptmx.Name(), err)
			}
		}
	})
	resize <- syscall.SIGWINCH

	var oldState *term.State
	if oldState, err = term.MakeRaw(int(callTty.Fd())); err != nil {
		return fmt.Errorf("term.MakeRaw error: %v", err)
	}

	var cancelPtmxToTty, cancelTtyToPtmx context.CancelFunc
	if cancelPtmxToTty, err = CopyWithCancel("OK", ptmx, callTty); err != nil {
		return fmt.Errorf("CopyWithCancel [OK] error: %v", err)
	}
	if cancelTtyToPtmx, err = CopyWithCancel("NOK", callTty, ptmx); err != nil {
		cancelPtmxToTty()
		return fmt.Errorf("CopyWithCancel [NOK] error: %v", err)
	}

	err = fn(ptty, ptty)               // blocking
	time.Sleep(time.Millisecond * 100) // let things catch up?

	signal.Stop(resize)
	close(resize)

	if e := ptmx.Close(); e != nil {
		log.ErrorF("ptmx.Close error: %v", e)
	}

	if e := ptty.Close(); e != nil {
		log.ErrorF("ptty.Close error: %v", e)
	}

	log.DebugF("sending Tiocsti: %v", callTty.Fd())
	if e := cterm.Tiocsti(callTty.Fd(), " "); e != nil {
		log.ErrorF("cterm.Tiocsti error: %v", e)
	}

	cancelPtmxToTty()
	cancelTtyToPtmx()

	if oldState != nil {
		log.DebugF("restoring term state: %v", callTty.Name())
		if e := term.Restore(int(callTty.Fd()), oldState); e != nil {
			log.ErrorF("term.Restore error: %v", e)
		}
	}

	return
}
