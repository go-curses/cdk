// Copyright 2022  The CDK Authors
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

package log

import (
	"io/ioutil"
	"os"
)

var (
	_cdk__prev_exit            = GetExitFunc()
	_cdk__fake_exited      int = -1
	_cdk__last_fake_logged     = ""
	_cdk__last_fake_exited     = -1
	_cdk__last_fake_error  error
)

func FakeExiting() {
	ResetFakeExited()
	_cdk__prev_exit = GetExitFunc()
	SetExitFunc(func(code int) {
		_cdk__fake_exited = code
	})
}

func DidFakeExit() bool {
	return _cdk__fake_exited >= 0
}

func ResetFakeExited() {
	_cdk__fake_exited = -1
}

func RestoreExiting() {
	SetExitFunc(_cdk__prev_exit)
	_cdk__prev_exit = nil
	_cdk__fake_exited = -1
}

func DoWithFakeIO(fn func() error) (string, bool, error) {
	real_stdout := os.Stdout
	real_stderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	FakeExiting()
	fn_err := fn()
	faked_it := DidFakeExit()
	RestoreExiting()
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = real_stdout // restoring the real stdout
	os.Stderr = real_stderr // restoring the real stdout
	_cdk__last_fake_error = fn_err
	_cdk__last_fake_logged = string(out)
	_cdk__last_fake_exited = _cdk__fake_exited
	if fn_err != nil {
		return "", faked_it, fn_err
	}
	return string(out), faked_it, nil
}

func GetLastFakeIO() (string, int, error) {
	return _cdk__last_fake_logged, _cdk__last_fake_exited, _cdk__last_fake_error
}
