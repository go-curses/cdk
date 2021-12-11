// +build lockStack

package sync

import (
	"fmt"
	"runtime"
)

func (m *RWMutex) makeTag(write bool, depth int) (tag string) {
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
	m.lockStack = append(m.lockStack, m.makeTag(true, 1))
}

func (m *RWMutex) Unlock() {
	m.RWMutex.Unlock()
	if len(m.lockStack) > 1 {
		m.lockStack = append([]string{}, m.lockStack[:len(m.lockStack)-1]...)
	} else {
		m.lockStack = []string{}
	}
}

func (m *RWMutex) RLock() {
	m.RWMutex.RLock()
	m.lockStack = append(m.lockStack, m.getLockStackTag(false, 1))
}

func (m *RWMutex) RUnlock() {
	m.RWMutex.RUnlock()
	if len(m.lockStack) > 1 {
		m.lockStack = append([]string{}, m.lockStack[:len(m.lockStack)-1]...)
	} else {
		m.lockStack = []string{}
	}
}
