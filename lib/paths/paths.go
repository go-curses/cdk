// Copyright (c) 2021-2023  The Go-Curses Authors
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

package paths

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func FileWritable(path string) (writable bool) {
	if fi, err := os.Stat(path); err == nil {
		if fi.Mode().IsRegular() {
			writable = unix.Access(path, unix.W_OK) == nil
		}
	}
	return
}

func IsPipe(fh *os.File) (piped bool) {
	if fi, err := fh.Stat(); err == nil {
		if mode := fi.Mode(); mode > 0 {
			piped = mode&os.ModeNamedPipe != 0
		}
	}
	return
}

func IsDevice(path string) (present bool) {
	if fi, err := os.Stat(path); err == nil {
		if mode := fi.Mode(); mode > 0 {
			present = mode&(os.ModeDevice|os.ModeCharDevice) != 0
		}
	}
	return
}

func IsFile(path string) (found bool) {
	if fi, err := os.Stat(path); err == nil {
		found = fi.Mode().IsRegular()
	}
	return
}

func IsDir(path string) (found bool) {
	if fi, err := os.Stat(path); err == nil {
		found = fi.Mode().IsDir()
	}
	return
}

func MakeDir(path string, perm os.FileMode) error {
	if IsDir(path) {
		return fmt.Errorf("directory exists")
	}
	if IsFile(path) {
		return fmt.Errorf("given path is a file")
	}
	return os.MkdirAll(path, perm)
}

func ReadFile(path string) (content string, err error) {
	if IsFile(path) {
		var bytes []byte
		if bytes, err = ioutil.ReadFile(path); err == nil {
			content = string(bytes)
		}
		return
	}
	return "", fmt.Errorf("not a file or file not found: %v", path)
}

func WriteFile(path string, content string) (err error) {
	err = WriteFileWithPerms(path, content, 0644)
	return
}

func WriteFileWithPerms(path string, content string, perm fs.FileMode) (err error) {
	err = ioutil.WriteFile(path, []byte(content), perm)
	return
}

func CopyFile(src, dst string) (nBytes int64, err error) {
	// see: https://opensource.com/article/18/6/copying-files-go

	var srcInfo os.FileInfo
	if srcInfo, err = os.Stat(src); err != nil {
		return 0, err
	}

	if !srcInfo.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	var srcFile *os.File
	if srcFile, err = os.Open(src); err != nil {
		return 0, err
	}
	defer srcFile.Close()

	var destination *os.File
	if destination, err = os.Create(dst); err != nil {
		return 0, err
	}
	defer destination.Close()

	nBytes, err = io.Copy(destination, srcFile)
	return
}

// BackupFile will copy a file to the same location, with the same file name and
// suffixed with the given tag and finally a datestamp in the format of:
// YYYYMMDDHHMMSS
func BackupFile(tag, path string) (err error) {
	if IsFile(path) {
		stamp := time.Now().Format("20060102150405")
		dst := fmt.Sprintf("%v.%v.%v", path, tag, stamp)
		_, err = CopyFile(path, dst)
	}
	return
}