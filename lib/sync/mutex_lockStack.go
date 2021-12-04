// +build lockStack

package sync

import (
	"fmt"
	"runtime"
)

func (m *Mutex) getLockStackTag(write bool, depth int) (tag string) {
	depth += 1
	if pc, _, line, ok := runtime.Caller(depth); ok {
		details := runtime.FuncForPC(pc)
		tag = fmt.Sprintf("[write:%v] %v:%d", write, details.Name(), line)
	} else {
		tag = fmt.Sprintf("invalid depth: %d", depth)
	}
	return
}

func (m *Mutex) Lock() {
	m.Mutex.Lock()
	m.mutexLockStack = append(m.mutexLockStack, m.getLockStackTag(true, 1))
}

func (m *Mutex) Unlock() {
	m.Mutex.Unlock()
	if len(m.mutexLockStack) > 1 {
		m.mutexLockStack = append([]string{}, m.mutexLockStack[:len(m.mutexLockStack)-1]...)
	} else {
		m.mutexLockStack = []string{}
	}
}
