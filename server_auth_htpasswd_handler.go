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
	"fmt"

	cpaths "github.com/go-curses/cdk/lib/paths"
	"github.com/tg123/go-htpasswd"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
)

var (
	DefaultServerAuthHtpasswdPath = "./htpasswd"
)

type CServerAuthHtpasswdHandler struct {
	CServerAuthHandler

	defPath  string
	htpasswd *htpasswd.File
}

func NewServerAuthHtpasswdHandler(defaultHttpasswdFilePath string) (handler *CServerAuthHtpasswdHandler) {

	handler = &CServerAuthHtpasswdHandler{
		defPath: defaultHttpasswdFilePath,
	}
	handler.Init()
	return
}

func (h *CServerAuthHtpasswdHandler) Init() (already bool) {
	if h.CServerAuthHandler.Init() {
		return true
	}
	h.Lock()
	h.htpasswd = nil
	if h.defPath == "" {
		h.defPath = DefaultServerAuthHtpasswdPath
	}
	h.Unlock()
	h.RegisterArgument(&cli.PathFlag{
		Name:        "htpasswd",
		Usage:       "sets the path to the htpasswd file",
		Value:       h.defPath,
		DefaultText: h.defPath,
	})
	return
}

func (h *CServerAuthHtpasswdHandler) Reload(ctx *cli.Context) (err error) {
	h.Lock()
	htpasswdPath := ctx.String("htpasswd")
	if cpaths.IsFile(htpasswdPath) {
		if h.htpasswd, err = htpasswd.New(htpasswdPath, htpasswd.DefaultSystems, nil); err != nil {
			h.Unlock()
			return fmt.Errorf("failed to load htpasswd file: %v", err)
		}
	}
	h.Unlock()
	return
}

func (s *CServerAuthHtpasswdHandler) PasswordCallback(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	if s.htpasswd.Match(c.User(), string(pass)) {
		return nil, nil
	}
	return nil, fmt.Errorf("username or password rejected: %q", c.User())
}
