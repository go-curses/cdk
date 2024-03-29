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

package memphis

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"

	"github.com/go-curses/cdk/lib/paint"
	"github.com/go-curses/cdk/lib/ptypes"
)

var (
	surfaces     = make(map[uuid.UUID]*CSurface)
	surfacesLock = &sync.RWMutex{}
)

func MakeSurface(id uuid.UUID, origin ptypes.Point2I, size ptypes.Rectangle, style paint.Style) (err error) {
	surfacesLock.Lock()
	defer surfacesLock.Unlock()
	if _, ok := surfaces[id]; ok {
		return fmt.Errorf("surface exists for id: %v", id)
	}
	surfaces[id] = NewSurface(origin, size, style)
	return nil
}

func HasSurface(id uuid.UUID) (ok bool) {
	surfacesLock.RLock()
	defer surfacesLock.RUnlock()
	_, ok = surfaces[id]
	return
}

func GetSurface(id uuid.UUID) (*CSurface, error) {
	surfacesLock.RLock()
	defer surfacesLock.RUnlock()
	if c, ok := surfaces[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("surface not found: %v", id)
}

func RemoveSurface(id uuid.UUID) {
	surfacesLock.Lock()
	defer surfacesLock.Unlock()
	if _, ok := surfaces[id]; ok {
		delete(surfaces, id)
	}
}

func FillSurface(id uuid.UUID, theme paint.Theme) (err error) {
	var s Surface
	if s, err = GetSurface(id); err == nil {
		s.Fill(theme)
	}
	return
}

func ConfigureSurface(id uuid.UUID, origin ptypes.Point2I, size ptypes.Rectangle, style paint.Style) (err error) {
	var s Surface
	if s, err = GetSurface(id); err == nil {
		s.SetOrigin(origin)
		s.SetStyle(style)
		s.Resize(size)
	}
	return
}

func MakeConfigureSurface(id uuid.UUID, origin ptypes.Point2I, size ptypes.Rectangle, style paint.Style) (err error) {
	var s Surface
	if s, err = GetSurface(id); err != nil {
		err = MakeSurface(id, origin, size, style)
	} else {
		s.SetOrigin(origin)
		s.SetStyle(style)
		s.Resize(size)
	}
	return
}
