//go:build js || nacl || plan9 || windows
// +build js nacl plan9 windows

// Copyright (c) 2022-2023  The Go-Curses Authors
// Copyright 2021 The TCell Authors
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

// NB: We might someday wish to move Windows to this model.   However,
// that would probably mean sacrificing some of the richer key reporting
// that we can obtain with the console API present on Windows.

func (d *CScreen) engage() error {
	return ErrNoScreen
}

func (d *CScreen) reengage() error {
	return ErrNoScreen
}

func (d *CScreen) disengage() {
}

func (d *CScreen) initialize() error {
	return ErrNoScreen
}

func (d *CScreen) finalize() {
}

func (d *CScreen) getWinSize() (int, int, error) {
	return 0, 0, ErrNoScreen
}

func (d *CScreen) Beep() error {
	return ErrNoScreen
}
