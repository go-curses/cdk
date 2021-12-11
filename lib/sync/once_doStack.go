// +build doStack

package sync

func (o *Once) makeTag(depth int) (tag string) {
	depth += 1
	if pc, _, line, ok := runtime.Caller(depth); ok {
		details := runtime.FuncForPC(pc)
		tag = fmt.Sprintf("%v:%d", details.Name(), line)
	} else {
		tag = fmt.Sprintf("invalid depth: %d", depth)
	}
	return
}

func (o *Once) Do(fn func()) {
	o.doStack = o.makeTag(1)
	o.Once.Do(fn)
}
