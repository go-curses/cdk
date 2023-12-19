// Copyright (c) 2022-2023  The Go-Curses Authors
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
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofrs/uuid"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"

	"github.com/go-curses/cdk/env"
	"github.com/go-curses/cdk/lib/enums"
	"github.com/go-curses/cdk/lib/exec"
	"github.com/go-curses/cdk/lib/sync"

	cstrings "github.com/go-curses/cdk/lib/strings"
	cterm "github.com/go-curses/cdk/lib/term"
	"github.com/go-curses/cdk/log"
)

// TODO: cleanup clients when server RequestQuit

const TypeApplicationServer CTypeTag = "cdk-application-server"

func init() {
	_ = TypesManager.AddType(TypeApplicationServer, nil)
}

type ApplicationServer interface {
	Object

	Init() (already bool)
	GetClients() (clients []uuid.UUID)
	GetClient(id uuid.UUID) (*CApplicationServerClient, error)
	App() (app *CApplication)
	Display() (display *CDisplay)
	SetListenAddress(address string)
	GetListenAddress() (address string)
	SetListenPort(port int)
	GetListenPort() (port int)
	Stop() (err error)
	Daemon() (err error)
	Start() (err error)
	ClearAuthHandlers()
	InstallAuthHandler(handler ServerAuthHandler) (err error)
	UnInstallAuthHandler(handler ServerAuthHandler) (err error)
}

type CApplicationServer struct {
	CObject

	name         string
	usage        string
	description  string
	version      string
	tag          string
	title        string
	clientInitFn SignalListenerFn
	serverInitFn SignalListenerFn

	privateKeyPath string

	listenAddress string
	listenPort    int

	app     *CApplication
	display *CDisplay

	handlers []ServerAuthHandler
	config   *ssh.ServerConfig
	listener net.Listener
	clients  map[uuid.UUID]*CApplicationServerClient

	daemonize bool
}

func NewApplicationServer(name, usage, description, version, tag, title string, clientInitFn SignalListenerFn, serverInitFn SignalListenerFn, privateKeyPath string) *CApplicationServer {
	as := &CApplicationServer{
		name:           name,
		usage:          usage,
		description:    description,
		version:        version,
		tag:            tag,
		title:          title,
		clientInitFn:   clientInitFn,
		serverInitFn:   serverInitFn,
		listenAddress:  "0.0.0.0",
		listenPort:     2200,
		privateKeyPath: privateKeyPath,
	}
	as.Init()
	return as
}

func (s *CApplicationServer) Init() (already bool) {
	if s.InitTypeItem(TypeApplicationServer, s) {
		return true
	}
	s.CObject.Init()
	s.clients = make(map[uuid.UUID]*CApplicationServerClient)
	s.handlers = []ServerAuthHandler{
		NewDefaultServerAuthHandler(),
	}
	s.app = NewApplication(s.name, s.usage, s.description, s.version, s.tag, s.title, "/dev/tty")
	s.app.Connect(SignalStartup, "application-server-startup--server", func(data []interface{}, argv ...interface{}) enums.EventFlag {
		return s.serverInitFn(data, argv...)
	})
	s.app.runFn = s.runner
	s.display = s.app.display
	s.App().AddFlag(&cli.BoolFlag{
		Name:  "daemon",
		Usage: "start a server daemon instead of a server terminal",
		Value: false,
	})
	s.App().AddFlag(&cli.StringFlag{
		Name:        "listen-address",
		Usage:       "sets the address for the server to listen on",
		Value:       s.listenAddress,
		DefaultText: s.listenAddress,
	})
	s.App().AddFlag(&cli.IntFlag{
		Name:        "listen-port",
		Usage:       "sets the port for the server to listen on",
		Value:       s.listenPort,
		DefaultText: fmt.Sprintf("%d", s.listenPort),
	})
	s.App().AddFlag(&cli.StringFlag{
		Name:        "id-rsa",
		Usage:       "sets the path to the server id_rsa file",
		Value:       s.privateKeyPath,
		DefaultText: s.privateKeyPath,
	})
	return false
}

func (s *CApplicationServer) newClient(conn *ssh.ServerConn, channels <-chan ssh.NewChannel, requests <-chan *ssh.Request) (asc *CApplicationServerClient, err error) {
	if id, err := uuid.NewV4(); err != nil {
		return nil, err
	} else {
		s.Lock()
		defer s.Unlock()
		asc = &CApplicationServerClient{
			id:          id,
			conn:        conn,
			channels:    channels,
			requests:    requests,
			application: nil,
		}
		s.clients[id] = asc
		return asc, nil
	}
}

func (s *CApplicationServer) GetClients() (clients []uuid.UUID) {
	s.RLock()
	defer s.RUnlock()
	for id, _ := range s.clients {
		clients = append(clients, id)
	}
	return
}

func (s *CApplicationServer) GetClient(id uuid.UUID) (*CApplicationServerClient, error) {
	s.RLock()
	defer s.RUnlock()
	if asc, ok := s.clients[id]; ok {
		return asc, nil
	}
	return nil, fmt.Errorf("client not found: %v", id)
}

func (s *CApplicationServer) freeClient(id uuid.UUID) (err error) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.clients[id]; ok {
		delete(s.clients, id)
		return nil
	}
	return fmt.Errorf("client not found: %v", id)
}

func (s *CApplicationServer) App() (app *CApplication) {
	s.RLock()
	app = s.app
	s.RUnlock()
	return
}

func (s *CApplicationServer) Display() (display *CDisplay) {
	s.RLock()
	display = s.display
	s.RUnlock()
	return
}

func (s *CApplicationServer) SetListenAddress(address string) {
	s.Lock()
	s.listenAddress = address
	s.Unlock()
}

func (s *CApplicationServer) GetListenAddress() (address string) {
	s.RLock()
	address = s.listenAddress
	s.RUnlock()
	return
}

func (s *CApplicationServer) SetListenPort(port int) {
	s.Lock()
	s.listenPort = port
	s.Unlock()
}

func (s *CApplicationServer) GetListenPort() (port int) {
	s.RLock()
	port = s.listenPort
	s.RUnlock()
	return
}

func (s *CApplicationServer) Stop() (err error) {
	s.Lock()
	s.daemonize = false
	if s.display != nil {
		_ = s.display.AwaitCall(func(_ Display) error {
			s.display.RequestSync()
			return nil
		})
		s.Unlock()
		return
	}
	s.Unlock()
	return nil
}

func (s *CApplicationServer) Daemon() (err error) {
	s.Lock()
	s.daemonize = true
	app := s.app
	s.Unlock()
	err = app.Run(os.Args)
	return
}

func (s *CApplicationServer) Start() (err error) {
	s.Lock()
	s.daemonize = false
	app := s.app
	s.Unlock()
	err = app.Run(os.Args)
	return
}

func (s *CApplicationServer) ClearAuthHandlers() {
	s.Lock()
	handlers := s.handlers
	s.Unlock()
	for _, handler := range handlers {
		if err := s.UnInstallAuthHandler(handler); err != nil {
			log.Error(err)
		}
	}
	s.Lock()
	s.handlers = make([]ServerAuthHandler, 0)
	s.Unlock()
	return
}

func (s *CApplicationServer) InstallAuthHandler(handler ServerAuthHandler) (err error) {
	s.Lock()
	s.handlers = append(s.handlers, handler)
	s.Unlock()
	err = handler.Attach(s)
	return
}

func (s *CApplicationServer) UnInstallAuthHandler(handler ServerAuthHandler) (err error) {
	s.Lock()
	index := -1
	for idx, h := range s.handlers {
		if h.ID() == handler.ID() {
			index = idx
			break
		}
	}
	s.Unlock()
	if index > -1 {
		s.Lock()
		s.handlers = append(s.handlers[:index], s.handlers[index+1:]...)
		s.Unlock()
		err = handler.Detach()
	}
	return
}

func (s *CApplicationServer) handlerHasArg(arg string) (has bool) {
	s.RLock()
	handlers := s.handlers
	s.RUnlock()
	if arg[:2] == "--" {
		arg = arg[2:]
	}
	for _, h := range handlers {
		if h.HasArgument(arg) {
			has = true
			break
		}
	}
	return
}

func (s *CApplicationServer) runner(ctx *cli.Context) (err error) {
	if !s.daemonize {
		s.daemonize = ctx.Bool("daemon")
	}
	s.privateKeyPath = ctx.String("id-rsa")
	s.listenAddress = ctx.String("listen-address")
	s.listenPort = ctx.Int("listen-port")

	var args []string
	for _, arg := range os.Args {
		switch arg {
		case "--daemon":
		case "--listen-address":
		case "--listen-port":
		case "--id-rsa":
		default:
			if !s.handlerHasArg(arg) {
				args = append(args, arg)
			}
		}
	}
	os.Args = args

	var handler ServerAuthHandler
	if len(s.handlers) > 0 {
		handler = s.handlers[0]
		for _, handler := range s.handlers {
			if err := handler.Reload(ctx); err != nil {
				log.Error(err)
			}
		}
	}

	s.config = nil
	if handler != nil {
		if passwordHandler, ok := handler.(ServerAuthPasswordHandler); ok {
			s.config = &ssh.ServerConfig{
				PasswordCallback: passwordHandler.PasswordCallback,
			}
		}
	}
	if s.config == nil {
		s.config = &ssh.ServerConfig{}
	}

	var privateBytes []byte
	privateBytes, err = ioutil.ReadFile(s.privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load private key: %v - %v", err, s.privateKeyPath)
	}
	var private ssh.Signer
	private, err = ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v - %v", err, s.privateKeyPath)
	}
	s.config.AddHostKey(private)

	s.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.listenAddress, s.listenPort))
	if err != nil {
		return fmt.Errorf("failed to listen on %s:%d (%v)", s.listenAddress, s.listenPort, err)
	}

	done := make(chan bool, 1)

	s.app.Display().Connect(SignalDisplayShutdown, "application-server-display-shutdown-handler", func(data []interface{}, argv ...interface{}) enums.EventFlag {
		s.LogInfo("display shutting down")
		for _, client := range s.clients {
			if client.application != nil {
				if display := client.application.Display(); display != nil {
					s.LogInfo("shutting down client: %v", client.id)
					display.RequestQuit()
				}
			}
		}
		return enums.EVENT_PASS
	})

	// Accept all connections
	Go(func() {
		log.InfoF("Listening on %s:%d", s.listenAddress, s.listenPort)
	runnerListenerLoop:
		for {
			tcpConn, err := s.listener.Accept()
			if err != nil {
				log.ErrorF("Failed to accept incoming connection (%s)", err)
				continue
			}
			// Before use, a handshake must be performed on the incoming net.Conn.
			var conn *ssh.ServerConn
			var channels <-chan ssh.NewChannel
			var requests <-chan *ssh.Request
			conn, channels, requests, err = ssh.NewServerConn(tcpConn, s.config)
			if err != nil {
				log.ErrorF("Failed to handshake (%s)", err)
				continue
			}
			var asc *CApplicationServerClient
			if asc, err = s.newClient(conn, channels, requests); err != nil {
				log.Error(err)
				continue
			}
			log.InfoF("New SSH connection from %s (%s)", asc.String(), asc.conn.ClientVersion())
			// Discard all global out-of-band Requests
			// go ssh.DiscardRequests(requests)
			// Accept all channels
			Go(func() { s.handleChannels(asc) })

			select {
			case <-done:
				log.DebugF("breaking runner listener loop")
				break runnerListenerLoop
			default: // nop
			}
		}
	})

	if s.daemonize {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		select {
		case rx := <-sig:
			log.InfoF("daemon caught signal: %v", rx)
			done <- true
		}
		log.DebugF("daemon exiting")
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	display := s.app.Display()
	GoWithMainContext(
		env.Get("USER", "nil"),
		"/dev/tty",
		display,
		s.Self(),
		func() {
			display.Connect(SignalDisplayStartup, ApplicationServerDisplayStartupHandle, func(data []interface{}, argv ...interface{}) enums.EventFlag {
				_ = display.Disconnect(SignalDisplayStartup, ApplicationServerDisplayStartupHandle)
				if ctx, cancel, wg, ok := DisplaySignalDisplayStartupArgv(argv...); ok {
					if f := s.app.Emit(SignalStartup, s.app.Self(), display, ctx, cancel, wg); f == enums.EVENT_STOP {
						s.app.LogInfo("application server display startup signal listener requested EVENT_STOP")
						display.RequestQuit()
					}
					return enums.EVENT_PASS
				}
				return enums.EVENT_STOP
			})
			err = display.Run()
			wg.Done()
		},
	)
	wg.Wait()
	return err
}

func (s *CApplicationServer) handleChannels(asc *CApplicationServerClient) {
	// Service the incoming channel in goroutine
	for newChannel := range asc.channels {
		Go(func() { s.handleChannel(asc, newChannel) })
	}
}

func (s *CApplicationServer) handleChannel(asc *CApplicationServerClient, channel ssh.NewChannel) {
	// Since we're handling a shell, we expect a
	// channel type of "session". The also describes
	// "x11", "direct-tcpip" and "forwarded-tcpip"
	// channel types.
	if t := channel.ChannelType(); t != "session" {
		_ = channel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	connection, requests, err := channel.Accept()
	if err != nil {
		s.LogError("Could not accept channel (%s)", err)
		return
	}

	app := NewApplication(
		s.name,
		s.usage,
		s.description,
		s.version,
		s.tag,
		s.title,
		"",
	)
	asc.application = app
	app.Connect(SignalStartup, "application-server-startup--client", func(data []interface{}, argv ...interface{}) enums.EventFlag {
		return s.clientInitFn(data, argv...)
	})

	var display *CDisplay
	var cancel context.CancelFunc
	var resize exec.SpawnResize
	var wg *sync.WaitGroup
	valid := false
	once := &sync.Once{}

	if cancel, resize, wg, err = exec.Spawn(
		connection,
		func(in, out *os.File) (err error) {
			display = NewDisplayWithHandle("display-service", out)
			display.app = app
			username := env.Get("USER", "nil")
			displayname := cstrings.MakeObjectName(app.name, username, asc.conn.RemoteAddr().String())
			display.SetName(displayname)
			_ = display.SetStringProperty(PropertyDisplayName, displayname)
			_ = display.SetStringProperty(PropertyDisplayUser, username)
			_ = display.SetStringProperty(PropertyDisplayHost, asc.conn.RemoteAddr().String())
			valid = true
			return
		},
		func() (err error) {
			if !valid {
				return fmt.Errorf("shutdown called before startup")
			}
			log.DebugF("handleChannel, requesting quit on display")
			display.RequestQuit()
			wg.Wait() // hold until shutdown signal is received
			if ok, err := connection.SendRequest("exit-status", true, []byte{0, 0, 0, 0}); err != nil {
				log.ErrorF("error sending exit-status channel request")
			} else {
				log.InfoF("received exit-status response: %v, on connection: %s", ok, asc.String())
			}
			if err := connection.Close(); err != nil {
				log.ErrorF("error closing ssh channel: %v", err)
			}
			if err := asc.conn.Close(); err != nil {
				log.ErrorF("error closing ssh connection: %v", err)
			}
			if err := s.freeClient(asc.id); err != nil {
				log.ErrorF("error freeing app client: %v", err)
			}
			log.DebugF("Session closed")
			return
		},
	); err != nil {
		s.LogErr(err)
		return
	}

	wg.Add(1)
	app.Connect(SignalShutdown, "exit-handler", func(_ []interface{}, _ ...interface{}) enums.EventFlag {
		log.DebugF("exiting client connection: %s", asc.String())
		wg.Done()
		once.Do(cancel)
		return enums.EVENT_PASS
	})

	// start display service
	Go(func() {
		GoWithMainContext(
			asc.conn.User(),
			asc.conn.RemoteAddr().String(),
			display,
			app.Self(),
			func() {
				if err := app.SetDisplay(display); err != nil {
					log.Error(err)
					once.Do(cancel)
					return
				}
				display.Connect(SignalDisplayStartup, "application-signal-display-startup-handler", func(data []interface{}, argv ...interface{}) enums.EventFlag {
					if ctx, dcancel, wg, ok := DisplaySignalDisplayStartupArgv(argv...); ok {
						if f := app.Emit(SignalStartup, app.Self(), display, ctx, dcancel, wg); f == enums.EVENT_STOP {
							app.LogInfo("application startup signal listener requested EVENT_STOP")
							display.RequestQuit()
							return enums.EVENT_STOP
						}
						return enums.EVENT_PASS
					}
					return enums.EVENT_STOP
				})
				if err := app.cli.Run(os.Args); err != nil {
					log.Error(err)
				}
				once.Do(cancel)
			},
		)
	})

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	Go(func() {
		for req := range requests {
			switch req.Type {
			case "exec":
				cmd := cterm.ParseValue(req.Payload)
				log.DebugF("! out-of-band request: exec = %v", cmd)
				_ = req.Reply(true, nil)
			case "shell":
				// only accept the default shell...
				// (i.e. no command in the Payload)
				// if len(req.Payload) == 0 {
				// 	req.Reply(true, nil)
				// }
				//
				// accepting any shell, just not using the payload...
				log.DebugF("! out-of-band request: shell = %v", req.Payload)
				_ = req.Reply(true, nil)
			case "pty-req":
				termLen := req.Payload[3]
				w, h := cterm.ParseDims(req.Payload[termLen+4:])
				if err := resize(w, h); err != nil {
					log.Error(err)
				}
				_ = display.PostEvent(NewEventResize(int(w), int(h)))
				// Responding true (OK) here will let the client
				// know we have a pty ready for input
				log.DebugF("! pty-req: w:%d, h:%d", w, h)
				_ = req.Reply(true, nil)
			case "window-change":
				w, h := cterm.ParseDims(req.Payload)
				if err := resize(w, h); err != nil {
					log.Error(err)
				}
				_ = display.PostEvent(NewEventResize(int(w), int(h)))
				x, y := cterm.ParseDims(req.Payload[8:])
				log.DebugF("! window-change: w:%v h:%v x:%v y:%v len:%v payload:%v", w, h, x, y, len(req.Payload), req.Payload)
				_ = req.Reply(true, nil)
			case "env":
				k, v := cterm.ParseKeyValue(req.Payload)
				log.DebugF("! env: %s => \"%s\"", k, v)
				_ = req.Reply(true, nil)
			default:
				log.DebugF("! out-of-band request: %v - %v", req.Type, req.Payload)
				_ = req.Reply(true, nil)
			}
		}
	})
}

const ApplicationServerDisplayStartupHandle = "application-server-display-startup-handler"
