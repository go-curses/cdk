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

package math

import (
	"fmt"
)

// ClampI - Returns the `value` given unless it's smaller than `min` or greater
// than `max`. If it's less than `min`, `min` is returned and if it's greater
// than `max` it returns `max`.
func ClampI(value, min, max int) int {
	if value >= min && value <= max {
		return value
	}
	if value > max {
		return max
	}
	return min
}

// ClampF - Returns the `value` given unless it's smaller than `min` or greater
// than `max`. If it's less than `min`, `min` is returned and if it's greater
// than `max` it returns `max`.
func ClampF(value, min, max float64) float64 {
	if value >= min && value <= max {
		return value
	}
	if value > max {
		return max
	}
	return min
}

// FloorI - Returns the `value` given unless it's less than `min`, in which case
// it returns `min`.
func FloorI(v, min int) int {
	if v < min {
		return min
	}
	return v
}

// SumInts - Add the given list of integers up and return the result.
func SumInts(ints []int) (sum int) {
	sum = 0
	for _, v := range ints {
		sum += v
	}
	return
}

// EqInts - Compare two integer arrays and if both are the same, returns true
// and false otherwise.
func EqInts(a, b []int) (same bool) {
	same = true
	if len(a) != len(b) {
		same = false
	} else {
		for i, av := range a {
			if av != b[i] {
				same = false
				break
			}
		}
	}
	return
}

// CeilF2I - Round the given floating point number to the nearest larger integer
// and return that as an integer.
func CeilF2I(v float64) int {
	delta := v - float64(int(v))
	if delta > 0 {
		return int(v) + 1
	}
	return int(v)
}

// DistInts - Distribute the value of max across the values of the given input
func DistInts(max int, in []int) (out []int) {
	if len(in) == 0 {
		out = make([]int, 0)
		return
	}
	out = append(out, in...)
	front := true
	first, last := 0, len(out)-1
	fw, bw := 0, last
	for SumInts(out) < max {
		if front {
			out[fw]++
			front = false
			fw++
			if fw > last {
				fw = first
			}
		} else {
			out[bw]++
			front = true
			bw--
			if bw < first {
				bw = last
			}
		}
	}
	return
}

// SolveSpaceAlloc - Resolve the issue of computing the gaps between things.
func SolveSpaceAlloc(nChildren, nSpace, minSpacing int) (increment int, gaps []int) {
	numGaps := nChildren - 1
	totalMinSpacing := minSpacing * numGaps
	availableSpace := nSpace - totalMinSpacing
	remainder := availableSpace % nChildren
	increment = availableSpace / nChildren
	extra := totalMinSpacing + remainder
	gaps = DistInts(extra, make([]int, numGaps))
	return
}

// Distribute - Another variation on distributing space between things.
func Distribute(total, available, parts, nChildren, spacing int) (values, gaps []int, err error) {
	numGaps := nChildren - 1
	if numGaps > 0 {
		gaps = make([]int, numGaps)
		for i := 0; i < numGaps; i++ {
			gaps[i] = spacing
		}
	} else {
		gaps = make([]int, 0)
	}
	available -= SumInts(gaps)
	values = make([]int, parts)
	if parts > 0 {
		values = DistInts(available, values)
	}
	totalValues := SumInts(values)
	totalGaps := SumInts(gaps)
	totalDist := totalValues + totalGaps
	if totalDist > total {
		err = fmt.Errorf("totalDist[%d] > total[%d]", totalDist, total)
	} else if totalDist < total {
		delta := total - totalDist
		values = DistInts(SumInts(values)+delta, values)
	}
	return
}
