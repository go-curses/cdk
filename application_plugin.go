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
	"plugin"
)

// LoadApplicationFromPlugin is a wrapper around
// LoadApplicationFromPluginWithExport, using a default exported symbol name of
// `CdkApp`.
func LoadApplicationFromPlugin(path string) (app Application, err error) {
	return LoadApplicationFromPluginWithExport("CdkApp", path)
}

// LoadApplicationFromPluginWithExport opens the shared object file indicated by
// the path argument, looks for the exported symbol (which needs to be to a
// valid cdk.Application instance), recasts and returns the result.
func LoadApplicationFromPluginWithExport(exported, path string) (app Application, err error) {
	var plug *plugin.Plugin
	if plug, err = plugin.Open(path); err != nil {
		return nil, err
	}

	var appSymbol plugin.Symbol
	if appSymbol, err = plug.Lookup(exported); err != nil {
		return nil, err
	}

	if appPointer, ok := appSymbol.(*Application); !ok {
		return nil, fmt.Errorf(
			"CdkApp value stored in plugin is not of cdk.Application type: %v (%T)\n",
			appSymbol,
			appSymbol,
		)
	} else {
		app = *appPointer
	}
	return
}
