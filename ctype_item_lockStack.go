// +build lockStack

package cdk

import (
	"fmt"
	"runtime"
)

func (o *CTypeItem) getLockStackTag(write bool, depth int) (tag string) {
	depth += 1
	if pc, _, line, ok := runtime.Caller(depth); ok {
		details := runtime.FuncForPC(pc)
		tag = fmt.Sprintf("[write:%v] %v:%d", write, details.Name(), line)
	} else {
		tag = fmt.Sprintf("invalid depth: %d", depth)
	}
	return
}

func (o *CTypeItem) Lock() {
	o.lockStack = append(o.lockStack, o.getLockStackTag(true, 1))
	o.RWMutex.Lock()
}

func (o *CTypeItem) Unlock() {
	o.RWMutex.Unlock()
	if size := len(o.lockStack); size > 0 {
		o.lockStack = o.lockStack[:size-1]
	}
}

func (o *CTypeItem) RLock() {
	o.lockStack = append(o.lockStack, o.getLockStackTag(false, 1))
	o.RWMutex.RLock()
}

func (o *CTypeItem) RUnlock() {
	o.RWMutex.RUnlock()
	if size := len(o.lockStack); size > 0 {
		o.lockStack = o.lockStack[:size-1]
	}
}
