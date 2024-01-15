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
	"context"

	"github.com/urfave/cli/v2"

	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/sync"
)

// TODO: need to go:generate WithArgv* wrappers from signals somehow, maybe
//       the fn signature? better yet would be to find the Emit(signal,...)
//       and determine both Argv* and WithArgv* funcs

type ApplicationMain func(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup)

type ApplicationRunFn = func(ctx *cli.Context) error

type ApplicationInitFn = func(app Application)

type ApplicationPrepareStartupFn = func(app Application, args []string) enums.EventFlag

type ApplicationStartupFn = func(
	app Application,
	display Display,
	ctx context.Context,
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
) enums.EventFlag

type ApplicationShutdownFn = func() enums.EventFlag

func WithArgvNoneWithFlagsSignal(fn func() enums.EventFlag) SignalListenerFn {
	return func(_ []interface{}, _ ...interface{}) enums.EventFlag {
		return fn()
	}
}

func WithArgvNoneSignal(fn func(), eventFlag enums.EventFlag) SignalListenerFn {
	return func(_ []interface{}, _ ...interface{}) enums.EventFlag {
		fn()
		return eventFlag
	}
}

func WithArgvApplicationSignalPrepareStartup(fn ApplicationPrepareStartupFn) SignalListenerFn {
	return func(_ []interface{}, argv ...interface{}) enums.EventFlag {
		if app, args, ok := ArgvApplicationSignalPrepareStartup(argv...); ok {
			return fn(app, args)
		}
		return enums.EVENT_STOP
	}
}

func ArgvApplicationSignalPrepareStartup(argv ...interface{}) (app Application, args []string, ok bool) {
	if ok = len(argv) == 2; ok {
		if app, ok = argv[0].(Application); ok {
			if args, ok = argv[1].([]string); ok {
				return
			}
		}
		app = nil
		args = nil
	}
	return
}

func WithArgvApplicationSignalStartup(fn ApplicationStartupFn) SignalListenerFn {
	return func(_ []interface{}, argv ...interface{}) enums.EventFlag {
		if app, display, ctx, cancel, wg, ok := ArgvApplicationSignalStartup(argv...); ok {
			return fn(app, display, ctx, cancel, wg)
		}
		return enums.EVENT_STOP
	}
}

func ArgvApplicationSignalStartup(argv ...interface{}) (app Application, display Display, ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, ok bool) {
	if len(argv) == 5 {
		if app, ok = argv[0].(Application); ok {
			if display, ok = argv[1].(Display); ok {
				if ctx, ok = argv[2].(context.Context); ok {
					if cancel, ok = argv[3].(context.CancelFunc); ok {
						if wg, ok = argv[4].(*sync.WaitGroup); ok {
							return
						}
						cancel = nil
					}
					ctx = nil
				}
				display = nil
			}
			app = nil
		}
	}
	return
}