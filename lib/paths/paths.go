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

package paths

import (
	"fmt"
	"io/ioutil"
	"os"

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
