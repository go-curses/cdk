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

	"github.com/go-curses/cdk/log"
	cterm "github.com/go-curses/term"
)

func CopyWithCancel(tag string, src, dst *os.File) (cancel context.CancelFunc, err error) {
	stop, waiting := false, false
	cancel = func() {
		log.DebugF("cancel copy: [%s] %v->%v", tag, src.Name(), dst.Name())
		stop = true
		if waiting {
			time.Sleep(time.Millisecond * 10)
			log.DebugF("sending Tiocsti to: %v", src.Name())
			if err := cterm.Tiocsti(src.Fd(), " "); err != nil {
				log.Error(err)
				log.DebugF("[trying again] writing Tiocsti: %v", src.Name())
				if _, err := src.Write([]byte(" ")); err != nil {
					log.Error(err)
				}
			}
		}
	}
	Go(func() {
		log.DebugF("start copy: [%s] %v->%v", tag, src.Name(), dst.Name())
		n := 0
		buf := make([]byte, 1)
		for !stop {
			// log.DebugF("waiting for read: %s", tag)
			waiting = true
			n, err = src.Read(buf)
			waiting = false
			// log.DebugF("read %v for: %s", buf[:n], tag)
			if err != nil && err != io.EOF {
				break
			}
			if n == 0 {
				break
			}
			if _, err = dst.Write(buf[:n]); err != nil {
				break
			}
		}
		log.DebugF("finish copy: [%s] %v->%v", tag, src.Name(), dst.Name())
	})
	return
}

func CopyIoWithCancel(tag string, src io.ReadWriter, dst io.Writer) (cancel context.CancelFunc, err error) {
	stop, waiting := false, false
	cancel = func() {
		log.DebugF("cancel copy: [%s]", tag)
		stop = true
		if waiting {
			time.Sleep(time.Millisecond * 10)
			log.DebugF("writing Tiocsti: %v", tag)
			_, err = src.Write([]byte(" "))
		}
	}
	Go(func() {
		log.DebugF("start copy: [%s]", tag)
		n := 0
		buf := make([]byte, 1)
		for !stop {
			// log.DebugF("waiting for read: %s", tag)
			waiting = true
			n, err = src.Read(buf)
			waiting = false
			// log.DebugF("read %v for: %s", buf[:n], tag)
			if err != nil && err != io.EOF {
				break
			}
			if n == 0 {
				break
			}
			if _, err = dst.Write(buf[:n]); err != nil {
				break
			}
		}
		log.DebugF("finish copy: [%s]", tag)
	})
	return
}
