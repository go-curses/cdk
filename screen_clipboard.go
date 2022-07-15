package cdk

import (
	"encoding/base64"

	"github.com/atotto/clipboard"

	"github.com/go-curses/cdk/log"
)

func (d *CScreen) CopyToClipboard(s string) {
	if err := clipboard.WriteAll(s); err != nil {
		log.Error(err)
	} else {
		log.DebugF("success (via github.com/atotto/clipboard): %v", s)
		return
	}
	b64 := base64.StdEncoding.EncodeToString([]byte(s))
	d.TPuts("\x1b]52;c;" + b64 + "\x07")
	log.DebugF("uncertain (via OSC-52 terminal sequence): %v", s)
}