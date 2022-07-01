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