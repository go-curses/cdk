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

// TODO: app determines if local and/or server?

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pkg/profile"
	"github.com/urfave/cli/v2"

	"github.com/go-curses/cdk/env"
	"github.com/go-curses/cdk/lib/enums"
	cfsorter "github.com/go-curses/cdk/lib/flag_sorter"
	cpaths "github.com/go-curses/cdk/lib/paths"
	cstrings "github.com/go-curses/cdk/lib/strings"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
)

const TypeApplication CTypeTag = "cdk-application"

func init() {
	_ = TypesManager.AddType(TypeApplication, nil)
}

var (
	cdkApps = make(map[uuid.UUID]*CApplication)

	DefaultGoProfilePath = os.TempDir() + string(os.PathSeparator) + "cdk.pprof"
	goProfile            interface{ Stop() }
)

type Application interface {
	Object

	Init() (already bool)
	SetupDisplay()
	Destroy()
	CLI() *cli.App
	GetContext() *cli.Context
	Tag() string
	Title() string
	Name() string
	Usage() string
	Description() string
	Version() string
	Reconfigure(name, usage, description, version, tag, title, ttyPath string)
	AddFlag(flag cli.Flag)
	RemoveFlag(flag cli.Flag) (removed bool)
	AddFlags(flags []cli.Flag)
	AddCommand(command *cli.Command)
	AddCommands(commands []*cli.Command)
	Display() *CDisplay
	SetDisplay(d *CDisplay) (err error)
	NotifyStartupComplete()
	Run(args []string) (err error)
	MainInit(argv ...interface{}) (ok bool)
	MainRun(runner ApplicationMain)
	MainEventsPending() (pending bool)
	MainIterateEvents()
	MainFinish()
	CliActionFn(ctx *cli.Context) (err error)
}

type CApplication struct {
	CObject

	id          uuid.UUID
	name        string
	usage       string
	description string
	version     string
	tag         string
	title       string
	ttyPath     string
	display     *CDisplay
	context     *cli.Context
	cli         *cli.App
	runFn       ApplicationRunFn
	valid       bool
	started     bool
}

func NewApplication(name, usage, description, version, tag, title, ttyPath string) *CApplication {
	id, _ := uuid.NewV4()
	app := &CApplication{
		id:          id,
		name:        name,
		usage:       usage,
		description: description,
		version:     version,
		tag:         tag,
		title:       title,
		ttyPath:     ttyPath,
		runFn:       nil,
	}
	app.Init()
	return app
}

func (app *CApplication) Init() (already bool) {
	if app.InitTypeItem(TypeApplicationServer, app) {
		return true
	}
	app.CObject.Init()
	app.started = false
	app.cli = &cli.App{
		Name:        app.name,
		Usage:       app.usage,
		Description: app.description,
		Version:     app.version,
		Flags:       GetApplicationCliFlags(),
		Commands:    []*cli.Command{},
		Action:      app.CliActionFn,
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{},
		Usage:   "display the version",
	}
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h", "usage"},
		Usage:   "display command-line usage information",
	}
	cdkApps[app.id] = app
	return false
}

func (app *CApplication) SetupDisplay() {
	if app.display == nil {
		display := NewDisplay(app.title, app.ttyPath)
		display.app = app
		username := env.Get("USER", "nil")
		display.SetName(cstrings.MakeObjectName(app.name, username, "/dev/tty"))
		_ = display.SetStringProperty(PropertyDisplayName, app.name)
		_ = display.SetStringProperty(PropertyDisplayUser, username)
		_ = display.SetStringProperty(PropertyDisplayHost, "/dev/tty")
		if err := app.SetDisplay(display); err != nil {
			app.LogErr(err)
		}
	}
	app.Emit(SignalSetupDisplay, app.display)
}

func (app *CApplication) Destroy() {
	app.started = false
	app.valid = false
	delete(cdkApps, app.id)
	if app.display != nil {
		app.display.Destroy()
	}
	app.display = nil
	app.context = nil
	app.cli = nil
}

func (app *CApplication) CLI() *cli.App {
	app.RLock()
	defer app.RUnlock()
	return app.cli
}

func (app *CApplication) GetContext() *cli.Context {
	app.RLock()
	defer app.RUnlock()
	return app.context
}

func (app *CApplication) Tag() string {
	app.RLock()
	defer app.RUnlock()
	return app.tag
}

func (app *CApplication) Title() string {
	app.RLock()
	defer app.RUnlock()
	return app.title
}

func (app *CApplication) Name() string {
	app.RLock()
	defer app.RUnlock()
	return app.name
}

func (app *CApplication) Usage() string {
	app.RLock()
	defer app.RUnlock()
	return app.usage
}

func (app *CApplication) Description() string {
	app.RLock()
	defer app.RUnlock()
	return app.description
}

func (app *CApplication) Version() string {
	app.RLock()
	defer app.RUnlock()
	return app.version
}

func (app *CApplication) Reconfigure(name, usage, description, version, tag, title, ttyPath string) {
	if f := app.Emit(SignalReconfigure, name, usage, description, version, tag, title, ttyPath); f == enums.EVENT_PASS {
		app.Lock()
		app.name = name
		app.usage = usage
		app.description = description
		app.version = version
		app.tag = tag
		app.title = title
		app.ttyPath = ttyPath
		if app.cli != nil {
			app.cli.Name = name
			app.cli.Usage = usage
			app.cli.Description = description
			app.cli.Version = version
		}
		app.Unlock()
		app.Emit(SignalChanged, name, usage, description, version, tag, title, ttyPath)
	}
}

func (app *CApplication) AddFlag(flag cli.Flag) {
	app.cli.Flags = append(app.cli.Flags, flag)
}

func (app *CApplication) RemoveFlag(flag cli.Flag) (removed bool) {
	index := -1
	for idx, f := range app.cli.Flags {
		if f.String() == flag.String() {
			index = idx
			break
		}
	}
	if index > -1 {
		app.cli.Flags = append(app.cli.Flags[:index], app.cli.Flags[index+1:]...)
		removed = true
	}
	return
}

func (app *CApplication) AddFlags(flags []cli.Flag) {
	for _, f := range flags {
		app.AddFlag(f)
	}
}

func (app *CApplication) AddCommand(command *cli.Command) {
	app.cli.Commands = append(app.cli.Commands, command)
}

func (app *CApplication) AddCommands(commands []*cli.Command) {
	for _, c := range commands {
		app.AddCommand(c)
	}
}

func (app *CApplication) Display() *CDisplay {
	app.RLock()
	defer app.RUnlock()
	return app.display
}

func (app *CApplication) SetDisplay(d *CDisplay) (err error) {
	app.RLock()
	if display := app.display; display != nil {
		app.RUnlock()
		if display.IsRunning() {
			return fmt.Errorf("cannot change a running Display")
		}
		_ = display.Disconnect(SignalDisplayStartup, ApplicationDisplayStartupHandle)
		_ = display.Disconnect(SignalDisplayShutdown, ApplicationDisplayShutdownHandle)
	} else {
		app.RUnlock()
	}
	app.Lock()
	app.display = d
	app.Unlock()
	app.display.Connect(
		SignalDisplayStartup,
		ApplicationDisplayStartupHandle,
		func(data []interface{}, argv ...interface{}) enums.EventFlag {
			_ = app.display.Disconnect(SignalDisplayStartup, ApplicationDisplayStartupHandle)
			if ctx, cancel, wg, ok := DisplaySignalDisplayStartupArgv(argv...); ok {
				if f := app.Emit(SignalStartup, app.Self(), app.display, ctx, cancel, wg); f == enums.EVENT_STOP {
					app.LogInfo("application startup signal listener requested EVENT_STOP")
					app.display.RequestQuit()
				}
				return enums.EVENT_PASS
			}
			return enums.EVENT_STOP
		},
	)
	app.display.Connect(
		SignalDisplayShutdown,
		ApplicationDisplayShutdownHandle,
		func(data []interface{}, argv ...interface{}) enums.EventFlag {
			_ = app.display.Disconnect(SignalDisplayShutdown, ApplicationDisplayShutdownHandle)
			return app.Emit(SignalShutdown)
		},
	)
	return
}

func (app *CApplication) NotifyStartupComplete() {
	if !app.started {
		app.started = app.display != nil
		if app.started {
			if f := app.Emit(SignalNotifyStartupComplete); f == enums.EVENT_STOP {
				app.LogInfo("application notify startup complete listener requested EVENT_STOP")
				app.display.RequestQuit()
				return
			}
			app.display.StartupComplete()
		}
	}
}

func (app *CApplication) Run(args []string) (err error) {
	if f := app.Emit(SignalPrepareStartup); f == enums.EVENT_STOP {
		app.LogDebug("application run SignalPrepareStartup requested EVENT_STOP")
		return
	}
	app.SetupDisplay()
	err = nil
	var wg sync.WaitGroup
	wg.Add(1)
	GoWithMainContext(
		env.Get("USER", "nil"),
		"localhost",
		app.display,
		app.Self(),
		func() {
			if app.cli.Commands != nil && len(app.cli.Commands) > 0 {
				sort.Sort(cli.CommandsByName(app.cli.Commands))
			} else {
				// prevent unnecessary help text output when there are no commands
				app.cli.Commands = nil
			}
			sort.Sort(cfsorter.FlagSorter(app.cli.Flags))
			err = app.cli.Run(args)
			wg.Done()
		},
	)
	wg.Wait()
	return err
}

// MainInit is used to initialize the Application based on the CLI arguments
// given at runtime.
//
// `argv` can be one of the following cases:
//  nil/empty     use only the environment variables, if any are set
//  *cli.Context  do not parse anything, just use existing context
//  ...string     parse the given strings as if it were os.Args
func (app *CApplication) MainInit(argv ...interface{}) (ok bool) {
	handled := false
	argc := len(argv)
	if argc > 1 {
		var args []string
		for _, arg := range argv {
			if v, ok := arg.(string); ok {
				args = append(args, v)
			}
		}
		app.cli.Action = func(ctx *cli.Context) error {
			app.context = ctx
			handled = true
			return nil
		}
		if err := app.cli.Run(args); err != nil {
			app.LogErr(err)
			return false
		}
	} else if argc == 1 {
		if ctx, ok := argv[0].(*cli.Context); ok {
			app.context = ctx
			handled = true
		}
	}
	if !handled {
		app.cli.Action = func(ctx *cli.Context) error {
			app.context = ctx
			return nil
		}
		if err := app.cli.Run([]string{app.name}); err != nil {
			app.LogErr(err)
			return false
		}
	}
	if Build.LogLevel {
		if v := app.context.String("cdk-log-level"); !cstrings.IsEmpty(v) {
			env.Set("GO_CDK_LOG_LEVEL", v)
		}
		if Build.LogLevels {
			if app.context.Bool("ctk-log-levels") {
				for i := len(log.LogLevels) - 1; i >= 0; i-- {
					fmt.Printf("%s\n", log.LogLevels[i])
				}
				return false
			}
		}
	}
	if Build.LogFile {
		if v := app.context.String("cdk-log-file"); !cstrings.IsEmpty(v) {
			env.Set("GO_CDK_LOG_OUTPUT", "file")
			env.Set("GO_CDK_LOG_FILE", v)
		}
	}
	if Build.LogTimestamps {
		if v := app.context.String("cdk-log-timestamps"); !cstrings.IsEmpty(v) && cstrings.IsBoolean(v) {
			env.Set("GO_CDK_LOG_TIMESTAMPS", v)
		}
	}
	if Build.LogTimestampFormat {
		if v := app.context.String("cdk-log-timestamp-format"); !cstrings.IsEmpty(v) {
			env.Set("GO_CDK_LOG_TIMESTAMP_FORMAT", v)
		}
	}
	profilePath := DefaultGoProfilePath
	if Build.Profiling {
		if v := app.context.String("cdk-profile-path"); !cstrings.IsEmpty(v) {
			if !cpaths.IsDir(v) {
				if err := cpaths.MakeDir(v, 0770); err != nil {
					log.Fatal(err)
				}
			}
			env.Set("GO_CDK_PROFILE_PATH", v)
			profilePath = v
		}
	}
	if err := log.StartRestart(); err != nil {
		panic(err)
	}
	if Build.Profiling {
		if v := app.context.String("cdk-profile"); !cstrings.IsEmpty(v) {
			v = strings.ToLower(v)
			var p goProfileFn
			env.Set("GO_CDK_PROFILE", v)
			// none, block, cpu, goroutine, mem, mutex, thread or trace
			switch v {
			case "block":
				p = profile.BlockProfile
			case "cpu":
				p = profile.CPUProfile
			case "goroutine":
				p = profile.GoroutineProfile
			case "mem":
				p = profile.MemProfile
			case "mutex":
				p = profile.MutexProfile
			case "thread":
				p = profile.ThreadcreationProfile
			case "trace":
				p = profile.TraceProfile
			default:
				p = nil
			}
			if p != nil {
				log.DebugF("starting profile of \"%v\" to path: %v", v, profilePath)
				// defer profile.Start(p, profile.ProfilePath(profilePath)).Stop()
				goProfile = profile.Start(p, profile.ProfilePath(profilePath))
			}
		}
	}
	return true
}

func (app *CApplication) MainRun(runner ApplicationMain) {
	app.SetupDisplay()
	display := app.Display()
	var wg *sync.WaitGroup
	GoWithMainContext(
		env.Get("USER", "nil"),
		"localhost",
		display,
		app.Self(),
		func() {
			var ctx context.Context
			var cancel context.CancelFunc
			ctx, cancel, wg = display.Startup()
			wg.Add(1)
			if f := app.Emit(SignalStartup, app.Self(), display, ctx, cancel, wg); f == enums.EVENT_STOP {
				app.LogInfo("application startup signal listener requested EVENT_STOP")
				app.display.RequestQuit()
			}
			if runner != nil {
				runner(ctx, cancel, wg)
			}
			wg.Done()
		},
	)
	wg.Wait()
	return
}

func (app *CApplication) MainEventsPending() (pending bool) {
	if d := app.Display(); d != nil {
		pending = d.HasPendingEvents()
	}
	return
}

func (app *CApplication) MainIterateEvents() {
	if d := app.Display(); d != nil {
		d.IterateBufferedEvents()
	}
}

func (app *CApplication) MainFinish() {
	if d := app.Display(); d != nil {
		d.MainFinish()
	}
}

func (app *CApplication) CliActionFn(ctx *cli.Context) (err error) {
	if !app.MainInit(ctx) {
		return nil
	}
	if app.runFn != nil {
		return app.runFn(ctx)
	}
	return app.Display().Run()
}

const SignalReconfigure Signal = "reconfigure"

const SignalChanged Signal = "changed"

const SignalPrepareStartup Signal = "prepare-startup"

const SignalSetupDisplay Signal = "setup-display"

const SignalStartup Signal = "startup"

const SignalActivate Signal = "activate"

const SignalShutdown Signal = "shutdown"

const SignalNotifyStartupComplete Signal = "notify-startup-complete"

const ApplicationDisplayStartupHandle = "application-display-startup-handler"

const ApplicationDisplayShutdownHandle = "application-display-shutdown-handler"

type goProfileFn = func(p *profile.Profile)

type ApplicationMain func(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup)

type ApplicationRunFn = func(ctx *cli.Context) error

type ApplicationInitFn = func(app Application)

type ApplicationStartupFn = func(
	app Application,
	display Display,
	ctx context.Context,
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
) enums.EventFlag

type ApplicationShutdownFn = func() enums.EventFlag

// TODO: need to go:generate WithArgv* wrappers from signals somehow, maybe the fn signature?

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

func WithArgvApplicationSignalStartup(startupFn ApplicationStartupFn) SignalListenerFn {
	return func(_ []interface{}, argv ...interface{}) enums.EventFlag {
		if app, display, ctx, cancel, wg, ok := ArgvApplicationSignalStartup(argv...); ok {
			return startupFn(app, display, ctx, cancel, wg)
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
