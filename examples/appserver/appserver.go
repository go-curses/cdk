package main

import (
	"time"

	"github.com/go-curses/cdk"
	"github.com/go-curses/cdk/env"
	"github.com/go-curses/cdk/lib/enums"
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
		clientStartup,
		serverStartup,
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

func clientStartup(d cdk.Display) error {
	log.DebugF("initFn hit")
	d.CaptureCtrlC()
	w := &AppWindow{}
	w.Init()
	w.SetTitle("Client-Side")
	d.SetActiveWindow(w)
	// draw the screen every second so the time displayed is now
	cdk.AddTimeout(time.Second, func() enums.EventFlag {
		d.RequestDraw()         // redraw the window, is buffered
		d.RequestShow()         // flag buffer for immediate show
		return enums.EVENT_PASS // keep looping every second
	})
	// this happens on the server-side?
	d.AddQuitHandler("appwindow", func() {
		log.DebugF("clientStartup - exited normally.\n")
	})
	return nil
}

func serverStartup(d cdk.Display) error {
	log.DebugF("initFn hit")
	d.CaptureCtrlC()
	w := &AppWindow{}
	w.Init()
	w.SetTitle("Server-Side")
	d.SetActiveWindow(w)
	// draw the screen every second so the time displayed is now
	cdk.AddTimeout(time.Second, func() enums.EventFlag {
		d.RequestDraw()         // redraw the window, is buffered
		d.RequestShow()         // flag buffer for immediate show
		return enums.EVENT_PASS // keep looping every second
	})
	// this happens on the server-side?
	d.AddQuitHandler("appwindow", func() {
		log.DebugF("serverStartup - exited normally.\n")
	})
	return nil
}
