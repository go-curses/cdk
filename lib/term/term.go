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
