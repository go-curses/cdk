// Copyright 2021  The CDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cdk

import (
	"fmt"

	"github.com/go-curses/cdk/log"
	"github.com/jtolio/gls"
)

var (
	cdkContextManager = gls.NewContextManager()
	cdkContextKey     = gls.GenSym()
)

type AppContext struct {
	Display *CDisplay
	Host    string
	User    string
	Data    interface{}
}

func newGlsValuesWithContext(user, host string, display *CDisplay, data interface{}) (values gls.Values) {
	values = gls.Values{
		cdkContextKey: &AppContext{
			Display: display,
			Host:    host,
			User:    user,
			Data:    data,
		},
	}
	return
}

func IsLocalContextValid() (valid bool) {
	if v, ok := cdkContextManager.GetValue(cdkContextKey); ok {
		_, valid = v.(*AppContext)
	}
	return
}

func GetLocalContext() (ac *AppContext, err error) {
	if v, ok := cdkContextManager.GetValue(cdkContextKey); ok {
		if vd, vok := v.(*AppContext); vok {
			ac = vd
		} else {
			err = fmt.Errorf("not an AppContext: %T", v)
		}
	} else {
		err = fmt.Errorf("context not found for this goroutine")
	}
	return
}

// GetDefaultDisplay returns the default display for the current app context
func GetDefaultDisplay() (dm *CDisplay) {
	if ac, err := GetLocalContext(); err == nil {
		dm = ac.Display
	} else {
		if len(cdkApps) == 1 {
			for _, app := range cdkApps {
				dm = app.display
			}
		}
		log.ErrorDF(1, "app context error: %v", err)
	}
	return
}
