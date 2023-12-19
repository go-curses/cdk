package paint

import (
	"sync"
)

var (
	pkgLock = &sync.RWMutex{}
)
