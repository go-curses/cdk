package term

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
)

// ParseDims extracts terminal dimensions (width x height) from the provided buffer.
func ParseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

func ParseValue(b []byte) (value string) {
	for _, c := range b[4:] {
		if c < 32 || c > 126 {
			break
		}
		value += string(c)
	}
	return
}

func ParseKeyValue(b []byte) (key, value string) {
	idx := 0
	for i, c := range b[4:] {
		if c < 32 || c > 126 {
			idx = i + 4
			break
		}
		key += string(c)
	}
	for _, c := range b[idx:] {
		if c < 32 || c > 126 {
			continue
		}
		value += string(c)
	}
	// fmt.Printf("key:%s, value:%s, bytes: %v\n", key, value, b)
	return
}

// SetWinSz sets the width and height for the given tty fd
func SetWinSz(fd uintptr, w, h uint32) (err error) {
	ws := &struct {
		Height uint16
		Width  uint16
		x      uint16 // unused
		y      uint16 // unused
	}{
		Width:  uint16(w),
		Height: uint16(h),
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd, uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if errno > 0 {
		return fmt.Errorf("set tiocgwinsz error: [%v] %s", errno, errno.Error())
	}
	return nil
}
