package sync

import (
	"sync"
)

type WaitGroup struct {
	sync.WaitGroup
}
