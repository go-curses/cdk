package exec

import (
	"os"
	"os/exec"

	"github.com/go-curses/cdk/log"
)

func Command(callTty *os.File, name string, argv ...string) (err error) {
	return CallWithTty(
		callTty,
		func(in, out *os.File) (err error) {
			log.DebugF("calling Command(%v,%v)", name, argv)
			cmd := exec.Command(name, argv...)
			cmd.Stdin = in
			cmd.Stdout = out
			cmd.Stderr = out
			return cmd.Run()
		},
	)
}
