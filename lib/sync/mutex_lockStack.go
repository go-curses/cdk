//go:build lockStack
// +build lockStack

package sync

import (
	"fmt"
	"runtime"
	"strings"
)

func (m *Mutex) makeTag(write bool, depth int) (tag string) {
	depth += 1
	if pc, _, line, ok := runtime.Caller(depth); ok {
		details := runtime.FuncForPC(pc)
		name := details.Name()
		if strings.Contains(name, "(*CWidget).LockDraw") {
			if pc, _, line, ok = runtime.Caller(depth + 1); ok {
				details = runtime.FuncForPC(pc)
				name = details.Name()
			}
		}
		tag = fmt.Sprintf("[write:%v] %v:%d", write, name, line)
	} else {
		tag = fmt.Sprintf("invalid depth: %d", depth)
	}
	return
}

func (m *Mutex) Lock() {
	m.Mutex.Lock()
	m.lockStack = append(m.lockStack, m.makeTag(true, 1))
}

func (m *Mutex) Unlock() {
	m.Mutex.Unlock()
	if len(m.lockStack) > 1 {
		m.lockStack = append([]string{}, m.lockStack[:len(m.lockStack)-1]...)
	} else {
		m.lockStack = []string{}
	}
}
