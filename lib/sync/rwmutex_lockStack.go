// +build lockStack

package sync

import (
	"fmt"
	"runtime"
)

func (m *RWMutex) getLockStackTag(write bool, depth int) (tag string) {
	depth += 1
	if pc, _, line, ok := runtime.Caller(depth); ok {
		details := runtime.FuncForPC(pc)
		tag = fmt.Sprintf("[write:%v] %v:%d", write, details.Name(), line)
	} else {
		tag = fmt.Sprintf("invalid depth: %d", depth)
	}
	return
}

func (m *RWMutex) Lock() {
	m.RWMutex.Lock()
	m.rwMutexLockStack = append(m.rwMutexLockStack, m.getLockStackTag(true, 1))
}

func (m *RWMutex) Unlock() {
	m.RWMutex.Unlock()
	if len(m.rwMutexLockStack) > 1 {
		m.rwMutexLockStack = append([]string{}, m.rwMutexLockStack[:len(m.rwMutexLockStack)-1]...)
	} else {
		m.rwMutexLockStack = []string{}
	}
}

func (m *RWMutex) RLock() {
	m.RWMutex.RLock()
	m.rwMutexLockStack = append(m.rwMutexLockStack, m.getLockStackTag(false, 1))
}

func (m *RWMutex) RUnlock() {
	m.RWMutex.RUnlock()
	if len(m.rwMutexLockStack) > 1 {
		m.rwMutexLockStack = append([]string{}, m.rwMutexLockStack[:len(m.rwMutexLockStack)-1]...)
	} else {
		m.rwMutexLockStack = []string{}
	}
}
