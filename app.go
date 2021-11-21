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
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/profile"
	"github.com/urfave/cli/v2"

	"github.com/go-curses/cdk/env"
	cfsorter "github.com/go-curses/cdk/lib/flag_sorter"
	cpaths "github.com/go-curses/cdk/lib/paths"
	cstrings "github.com/go-curses/cdk/lib/strings"
	"github.com/go-curses/cdk/log"
)

type ScreenStateReq uint64

const (
	NullRequest ScreenStateReq = 1 << iota
	DrawRequest
	ShowRequest
	SyncRequest
	QuitRequest
)

var (
	cdkApps = make(map[uuid.UUID]*CApp)

	DefaultGoProfilePath = os.TempDir() + string(os.PathSeparator) + "cdk.pprof"
	goProfile            interface{ Stop() }
)

type goProfileFn = func(p *profile.Profile)

type DisplayInitFn = func(d Display) error

type AppMainFn func(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup)

type App interface {
	GetContext() *cli.Context
	Tag() string
	Title() string
	Name() string
	Usage() string
	Description() string
	Display() *CDisplay
	CLI() *cli.App
	Version() string
	SetupDisplay()
	InitUI() error
	AddFlag(f cli.Flag)
	AddCommand(c *cli.Command)
	Run(args []string) error
	CliActionFn(ctx *cli.Context) error
}

type CApp struct {
	id          uuid.UUID
	name        string
	usage       string
	description string
	version     string
	tag         string
	title       string
	ttyPath     string
	display     *CDisplay
	dispLock    *sync.RWMutex
	context     *cli.Context
	cli         *cli.App
	initFn      DisplayInitFn
	runFn       func(ctx *cli.Context) error
	valid       bool
}

func NewApp(name, usage, description, version, tag, title, ttyPath string, initFn DisplayInitFn) *CApp {
	id, _ := uuid.NewV4()
	app := &CApp{
		id:          id,
		name:        name,
		usage:       usage,
		description: description,
		version:     version,
		tag:         tag,
		title:       title,
		ttyPath:     ttyPath,
		initFn:      initFn,
		runFn:       nil,
	}
	app.init()
	return app
}

func (app *CApp) init() {
	app.dispLock = &sync.RWMutex{}
	app.cli = &cli.App{
		Name:        app.name,
		Usage:       app.usage,
		Description: app.description,
		Version:     app.version,
		Flags:       getAppCliFlags(),
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
	app.valid = true
	return
}

func (app *CApp) SetupDisplay() {
	if app.display == nil {
		display := NewDisplay(app.title, app.ttyPath)
		display.app = app
		username := env.Get("USER", "nil")
		display.SetName(cstrings.MakeObjectName(app.name, username, "/dev/tty"))
		_ = display.SetStringProperty(PropertyDisplayName, app.name)
		_ = display.SetStringProperty(PropertyDisplayUser, username)
		_ = display.SetStringProperty(PropertyDisplayHost, "/dev/tty")
		app.display = display
	}
}

func (app *CApp) Destroy() {
	app.valid = false
	delete(cdkApps, app.id)
	if app.display != nil {
		app.display.Destroy()
	}
	app.display = nil
	app.context = nil
	app.cli = nil
}

func (app *CApp) GetContext() *cli.Context {
	return app.context
}

func (app *CApp) Tag() string {
	return app.tag
}

func (app *CApp) Title() string {
	return app.title
}

func (app *CApp) Name() string {
	return app.name
}

func (app *CApp) Usage() string {
	return app.usage
}

func (app *CApp) Description() string {
	return app.description
}

func (app *CApp) Display() *CDisplay {
	app.dispLock.RLock()
	defer app.dispLock.RUnlock()
	return app.display
}

func (app *CApp) setDisplay(d *CDisplay) {
	app.dispLock.Lock()
	defer app.dispLock.Unlock()
	app.display = d
}

func (app *CApp) CLI() *cli.App {
	return app.cli
}

func (app *CApp) Version() string {
	return app.version
}

func (app *CApp) InitUI() error {
	return app.initFn(app.Display())
}

func (app *CApp) AddFlag(flag cli.Flag) {
	app.cli.Flags = append(app.cli.Flags, flag)
}

func (app *CApp) AddFlags(flags []cli.Flag) {
	for _, f := range flags {
		app.AddFlag(f)
	}
}

func (app *CApp) AddCommand(command *cli.Command) {
	app.cli.Commands = append(app.cli.Commands, command)
}

func (app *CApp) AddCommands(commands []*cli.Command) {
	for _, c := range commands {
		app.AddCommand(c)
	}
}

func (app *CApp) Run(args []string) (err error) {
	app.SetupDisplay()
	err = nil
	var wg sync.WaitGroup
	wg.Add(1)
	cdkContextManager.SetValues(
		newGlsValuesWithContext(env.Get("USER", "nil"), "localhost", app.display, nil),
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

func (app *CApp) MainInit(ctx *cli.Context) (ok bool) {
	if ctx == nil {
		app.context = cli.NewContext(app.cli, nil, nil)
	} else {
		app.context = ctx
	}
	if Build.LogLevel {
		if v := ctx.String("cdk-log-level"); !cstrings.IsEmpty(v) {
			env.Set("GO_CDK_LOG_LEVEL", v)
		}
		if Build.LogLevels {
			if ctx.Bool("ctk-log-levels") {
				for i := len(log.LogLevels) - 1; i >= 0; i-- {
					fmt.Printf("%s\n", log.LogLevels[i])
				}
				return false
			}
		}
	}
	if Build.LogFile {
		if v := ctx.String("cdk-log-file"); !cstrings.IsEmpty(v) {
			env.Set("GO_CDK_LOG_OUTPUT", "file")
			env.Set("GO_CDK_LOG_FILE", v)
		}
	}
	if Build.LogTimestamps {
		if v := ctx.String("cdk-log-timestamps"); !cstrings.IsEmpty(v) && cstrings.IsBoolean(v) {
			env.Set("GO_CDK_LOG_TIMESTAMPS", v)
		}
	}
	if Build.LogTimestampFormat {
		if v := ctx.String("cdk-log-timestamp-format"); !cstrings.IsEmpty(v) {
			env.Set("GO_CDK_LOG_TIMESTAMP_FORMAT", v)
		}
	}
	profilePath := DefaultGoProfilePath
	if Build.Profiling {
		if v := ctx.String("cdk-profile-path"); !cstrings.IsEmpty(v) {
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
		if v := ctx.String("cdk-profile"); !cstrings.IsEmpty(v) {
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

func (app *CApp) MainRun(fn AppMainFn) {
	app.SetupDisplay()
	var wg *sync.WaitGroup
	cdkContextManager.SetValues(
		newGlsValuesWithContext(env.Get("USER", "nil"), "localhost", app.display, nil),
		func() {
			var ctx context.Context
			var cancel context.CancelFunc
			ctx, cancel, wg = app.Display().MainInit()
			wg.Add(1)
			fn(ctx, cancel, wg)
			wg.Done()
		},
	)
	wg.Wait()
	return
}

func (app *CApp) MainEventsPending() (pending bool) {
	if d := app.Display(); d != nil {
		pending = d.HasPendingEvents()
	}
	return
}

func (app *CApp) MainIterateEvents() {
	if d := app.Display(); d != nil {
		d.IterateBufferedEvents()
	}
}

func (app *CApp) MainFinish() {
	if d := app.Display(); d != nil {
		d.MainFinish()
	}
}

func (app *CApp) CliActionFn(ctx *cli.Context) error {
	if !app.MainInit(ctx) {
		return nil
	}
	if app.runFn != nil {
		return app.runFn(ctx)
	}
	return app.Display().Run()
}
