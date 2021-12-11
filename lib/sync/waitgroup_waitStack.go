// +build waitStack

package sync

import (
	"fmt"
	"runtime"
)

func (wg *WaitGroup) makeTag(depth int) (tag string) {
	depth += 1
	if pc, _, line, ok := runtime.Caller(depth); ok {
		details := runtime.FuncForPC(pc)
		tag = fmt.Sprintf("%v:%d", details.Name(), line)
	} else {
		tag = fmt.Sprintf("invalid depth: %d", depth)
	}
	return
}

func (wg *WaitGroup) Add(delta int) {
	wg.WaitGroup.Add(delta)
	wg.waitStack = append(wg.waitStack, waitStackItem{
		Name:  wg.makeTag(1),
		Delta: delta,
	})
}

func (wg *WaitGroup) Done() {
	if len(wg.waitStack) > 1 {
		wg.waitStack = append([]waitStackItem{}, wg.waitStack[:len(wg.waitStack)-1]...)
	} else {
		wg.waitStack = []waitStackItem{}
	}
}
