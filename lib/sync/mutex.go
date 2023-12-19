package sync

import (
	goSync "sync"
)

type Mutex struct {
	goSync.Mutex

	lockStack []string
}
