package sync

import (
	"sync"
)

type Mutex struct {
	sync.Mutex

	lockStack []string
}
