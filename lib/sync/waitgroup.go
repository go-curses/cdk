package sync

import (
	"sync"
)

type waitStackItem struct {
	Name  string
	Delta int
}

type WaitGroup struct {
	sync.WaitGroup

	waitStack []waitStackItem
}
