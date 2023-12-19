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

package bits

// Set returns `flag` with the given `mask` added.
func Set(flag, mask uint64) uint64 { return flag | mask }

// Clear returns `flag` with the given `mask` removed.
func Clear(b, flag uint64) uint64 { return b &^ flag }

// Toggle returns `flag` with the state of `mask` inverted. If it was present, its
// removed, if it was not present its added.
func Toggle(b, flag uint64) uint64 { return b ^ flag }

// Has returns TRUE if `flag` has the given `mask` present.
func Has(b, flag uint64) bool { return b&flag != 0 }
