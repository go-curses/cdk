// Copyright (c) 2023  The Go-Curses Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cdk

import (
	"fmt"

	"github.com/jtolio/gls"

	"github.com/go-curses/cdk/lib/exec"
	"github.com/go-curses/cdk/log"
)

type CLocalContextData struct {
	User    string
	Host    string
	Display *CDisplay
	Data    interface{}
}

func Go(fn func()) {
	gls.Go(fn)
}

func GetLocalContext() (acd *CLocalContextData, err error) {
	var lc *exec.CLocalContext
	if lc, err = exec.GetLocalContext(); err != nil {
		return
	} else if lc != nil {
		var ok bool
		if acd, ok = lc.Data.(*CLocalContextData); ok {
			return
		}
		err = fmt.Errorf("value stored in local context data is not *cdk.CLocalContextData: %v (%T)", lc.Data, lc.Data)
	} else {
		err = fmt.Errorf("local context not found")
	}
	return
}

// GetDefaultDisplay returns the default display for the current app context
func GetDefaultDisplay() (display *CDisplay) {
	if acd, err := GetLocalContext(); err == nil {
		display = acd.Display
	} else {
		log.ErrorDF(1, "error getting local context: %v", err)
		if len(cdkApps) == 1 {
			for _, app := range cdkApps {
				display = app.Display()
				break
			}
		}
	}
	if display == nil {
		log.ErrorF("default display not found")
	}
	return
}

func GoWithMainContext(user, host string, display *CDisplay, data interface{}, fn func()) {
	exec.GoWithMainContext(
		&CLocalContextData{
			User:    user,
			Host:    host,
			Display: display,
			Data:    data,
		},
		fn,
	)
}
