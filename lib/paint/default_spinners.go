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

package paint

type SpinnerName string

const (
	SafeSpinner     SpinnerName = "safe-spinner"
	OneDotSpinner   SpinnerName = "one-dot-spinner"
	OrbitDotSpinner SpinnerName = "orbit-dot-spinner"
	SevenDotSpinner SpinnerName = "seven-dot-spinner"
)

var (
	spinnerOverrides = map[SpinnerName]SpinnerRuneSet{}
)

func RegisterSpinners(name SpinnerName, spinner SpinnerRuneSet) {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	spinnerOverrides[name] = spinner
}

func GetSpinners(name SpinnerName) (arrow SpinnerRuneSet, ok bool) {
	pkgLock.RLock()
	defer pkgLock.RUnlock()
	if arrow, ok = spinnerOverrides[name]; !ok {
		switch name {
		case SafeSpinner:
			return SafeSpinnerRuneSet, true
		case OneDotSpinner:
			return OneDotSpinnerRuneSet, true
		case OrbitDotSpinner:
			return OrbitDotSpinnerRuneSet, true
		case SevenDotSpinner:
			return SevenDotSpinnerRuneSet, true
		}
	}
	return
}