package sync

import (
	goSync "sync"
)

type RWMutex struct {
	goSync.RWMutex

	lockStack []string
}