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
	"io"
	"os"
	"time"

	"github.com/creack/pty"
	cterm "github.com/go-curses/cdk/lib/term"

	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
)

type SpawnCallback = func() (err error)
type SpawnResize = func(w, h uint32) (err error)

func Spawn(device io.ReadWriter, startup Callback, shutdown SpawnCallback) (cancel context.CancelFunc, resize SpawnResize, wg *sync.WaitGroup, err error) {

	var ptmx, ptty *os.File
	if ptmx, ptty, err = pty.Open(); err != nil {
		return
	}

	wg = &sync.WaitGroup{}

	var cancelPtmxToTty, cancelTtyToPtmx context.CancelFunc
	if cancelPtmxToTty, err = CopyIoWithCancel("OK", ptmx, device); err != nil {
		_ = ptmx.Close()
		_ = ptty.Close()
		wg = nil
		return
	}
	if cancelTtyToPtmx, err = CopyIoWithCancel("NOK", device, ptmx); err != nil {
		_ = ptmx.Close()
		_ = ptty.Close()
		cancelPtmxToTty()
		wg = nil
		return
	}

	if err = startup(ptty, ptty); err != nil {
		cancelPtmxToTty()
		cancelTtyToPtmx()
		_ = ptmx.Close()
		_ = ptty.Close()
		wg = nil
		return
	}

	wg.Add(1)
	cancel = func() {
		time.Sleep(time.Millisecond * 100) // let things catch up?
		cancelPtmxToTty()
		cancelTtyToPtmx()
		_ = ptmx.Close()
		_ = ptty.Close()
		wg.Done()
		if err := shutdown(); err != nil {
			log.Error(err)
		}
	}
	resize = func(w, h uint32) (err error) {
		err = cterm.SetWinSz(ptmx.Fd(), w, h)
		return
	}
	return
}
