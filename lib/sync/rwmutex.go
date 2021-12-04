package sync

import (
	"sync"
)

type RWMutex struct {
	sync.RWMutex

	rwMutexLockStack []string
}
