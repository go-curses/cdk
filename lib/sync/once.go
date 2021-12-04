package sync

import (
	"sync"
)

type Once struct {
	sync.Once
}
