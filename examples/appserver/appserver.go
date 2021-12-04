package main

import (
	"context"
	"time"

	"github.com/go-curses/cdk"
	"github.com/go-curses/cdk/env"
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/sync"
	"github.com/go-curses/cdk/log"
)

func main() {
	env.Set("GO_CDK_LOG_FILE", "./appserver.log")
	env.Set("GO_CDK_LOG_LEVEL", "debug")
	cdk.Init()
	as := cdk.NewApplicationServer(
		"appserver",
		"run an application server",
		"",
		"",
		"",
		"",
		cdk.WithArgvApplicationSignalStartup(clientStartup),
		cdk.WithArgvApplicationSignalStartup(serverStartup),
		"./examples/appserver/id_rsa",
	)
	as.ClearAuthHandlers()
	if err := as.InstallAuthHandler(
		cdk.NewServerAuthHtpasswdHandler(
			"./examples/appserver/htpasswd",
		),
	); err != nil {
		log.Error(err)
	}
	if err := as.Start(); err != nil {
		panic(err)
	}
}

func clientStartup(app cdk.Application, display cdk.Display, ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) enums.EventFlag {
	log.DebugF("initFn hit")
	display.CaptureCtrlC()
	w := &AppWindow{}
	w.Init()
	w.SetTitle("Client-Side")
	display.FocusWindow(w)
	// draw the screen every second so the time displayed is now
	cdk.AddTimeout(time.Second, func() enums.EventFlag {
		display.RequestDraw()   // redraw the window, is buffered
		display.RequestShow()   // flag buffer for immediate show
		return enums.EVENT_PASS // keep looping every second
	})
	display.Connect(cdk.SignalShutdown, "appserver-client-shutdown", func(_ []interface{}, _ ...interface{}) enums.EventFlag {
		log.DebugF("clientStartup - exited normally.\n")
		return enums.EVENT_PASS
	})
	app.NotifyStartupComplete()
	return enums.EVENT_PASS
}

func serverStartup(app cdk.Application, display cdk.Display, ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) enums.EventFlag {
	log.DebugF("initFn hit")
	display.CaptureCtrlC()
	w := &AppWindow{}
	w.Init()
	w.SetTitle("Server-Side")
	display.FocusWindow(w)
	// draw the screen every second so the time displayed is now
	cdk.AddTimeout(time.Second, func() enums.EventFlag {
		display.RequestDraw()   // redraw the window, is buffered
		display.RequestShow()   // flag buffer for immediate show
		return enums.EVENT_PASS // keep looping every second
	})
	display.Connect(cdk.SignalShutdown, "appserver-server-shutdown", func(_ []interface{}, _ ...interface{}) enums.EventFlag {
		log.DebugF("serverStartup - exited normally.\n")
		return enums.EVENT_PASS
	})
	app.NotifyStartupComplete()
	return enums.EVENT_PASS
}
