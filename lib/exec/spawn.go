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
