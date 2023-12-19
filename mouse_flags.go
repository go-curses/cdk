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

package cdk

// MouseFlags are options to modify the handling of mouse events.
// Actual events can be or'd together.
type MouseFlags int

const (
	MouseButtonEvents = MouseFlags(1) // Click events only
	MouseDragEvents   = MouseFlags(2) // Click-drag events (includes button events)
	MouseMotionEvents = MouseFlags(4) // All mouse events (includes click and drag events)
)
