package cdk

import (
	"github.com/go-curses/cdk/lib/enums"
)

type Sensitive interface {
	ProcessEvent(evt Event) enums.EventFlag
}
