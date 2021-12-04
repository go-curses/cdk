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
	"github.com/gofrs/uuid"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"

	"github.com/go-curses/cdk/lib/sync"
)

type ServerAuthHandler interface {
	Init() (already bool)
	ID() (id uuid.UUID)
	Attach(server ApplicationServer) (err error)
	Detach() (err error)
	Reload(ctx *cli.Context) (err error)
	HasArgument(arg string) (has bool)
	RegisterArgument(flag cli.Flag)
}

// CServerAuthHandler is the base type for application server authentication
// handler implementations. This handler does no authentication and just accepts
// all connections.
type CServerAuthHandler struct {
	id          uuid.UUID
	server      ApplicationServer
	arguments   []cli.Flag
	initialized bool

	sync.RWMutex
}

func NewDefaultServerAuthHandler() (handler ServerAuthHandler) {
	handler = &CServerAuthHandler{}
	handler.Init()
	return
}

func (h *CServerAuthHandler) Init() (already bool) {
	if h.initialized {
		return
	}
	h.Lock()
	h.id, _ = uuid.NewV4()
	h.server = nil
	h.arguments = make([]cli.Flag, 0)
	h.initialized = true
	h.Unlock()
	return
}

func (h *CServerAuthHandler) ID() (id uuid.UUID) {
	h.RLock()
	id = h.id
	h.RUnlock()
	return
}

func (h *CServerAuthHandler) Attach(server ApplicationServer) (err error) {
	h.RLock()
	if h.server != nil {
		h.RUnlock()
		if err = h.Detach(); err != nil {
			return err
		}
	}
	h.RUnlock()
	h.Lock()
	h.server = server
	for _, flag := range h.arguments {
		h.server.App().AddFlag(flag)
	}
	h.Unlock()
	return
}

func (h *CServerAuthHandler) Detach() (err error) {
	h.Lock()
	if h.server != nil {
		for _, flag := range h.arguments {
			h.server.App().RemoveFlag(flag)
		}
	}
	h.server = nil
	h.Unlock()
	return
}

func (h *CServerAuthHandler) Reload(ctx *cli.Context) (err error) {
	// nothing to reload
	return
}

func (h *CServerAuthHandler) HasArgument(arg string) (has bool) {
	for _, flag := range h.arguments {
		if flag.String() == arg {
			has = true
			break
		}
	}
	return
}

func (h *CServerAuthHandler) RegisterArgument(flag cli.Flag) {
	h.Lock()
	h.arguments = append(h.arguments, flag)
	attached := h.server != nil
	h.Unlock()
	if attached {
		h.server.App().AddFlag(flag)
	}
}

func (h *CServerAuthHandler) PasswordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	// if h.server.htpasswd.Match(conn.User(), string(password)) {
	// 	return nil, nil
	// }
	// return nil, fmt.Errorf("password rejected for %q", c.User())
	return nil, nil
}
