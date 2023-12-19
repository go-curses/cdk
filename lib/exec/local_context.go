// Copyright (c) 2021-2023  The Go-Curses Authors
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

package exec

import (
	"fmt"

	"github.com/jtolio/gls"

	"github.com/go-curses/cdk/log"
)

var (
	localContextManager = gls.NewContextManager()
	localContextKey     = gls.GenSym()
)

type CLocalContext struct {
	Data interface{}
}

func Go(fn func()) {
	gls.Go(fn)
}

func GoWithMainContext(data interface{}, fn func()) {
	localContextManager.SetValues(
		gls.Values{
			localContextKey: &CLocalContext{
				Data: data,
			},
		},
		fn,
	)
}

func GoWithLocalContext(data interface{}, fn func()) {
	if local, err := GetLocalContext(); err != nil {
		log.Error(err)
	} else if local != nil {
		local.Data = data
		localContextManager.SetValues(
			gls.Values{
				localContextKey: local,
			},
			fn,
		)
	} else {
		log.ErrorDF(1, "missing local context")
	}
}

func IsLocalContextValid() (valid bool) {
	if v, ok := localContextManager.GetValue(localContextKey); ok {
		_, valid = v.(*CLocalContext)
	}
	return
}

func GetLocalContext() (ac *CLocalContext, err error) {
	if v, ok := localContextManager.GetValue(localContextKey); ok {
		if vd, vok := v.(*CLocalContext); vok {
			ac = vd
		} else {
			err = fmt.Errorf("not a cdk.CLocalContext: %T", v)
		}
	} else {
		err = fmt.Errorf("context not found for this goroutine")
	}
	return
}
