// Code generated by "bitmasker -type ConnectFlags"; DO NOT EDIT.

package enums

import "strconv"

// Has returns TRUE if the given flag is present in the bitmask
func (i ConnectFlags) Has(m ConnectFlags) bool {
	return i&m != 0
}

// Set returns the bitmask with the given flag set
func (i ConnectFlags) Set(m ConnectFlags) ConnectFlags {
	return i | m
}

// Clear returns the bitmask with the given flag removed
func (i ConnectFlags) Clear(m ConnectFlags) ConnectFlags {
	return i &^ m
}

// Toggle returns the bitmask with the given flag toggled
func (i ConnectFlags) Toggle(m ConnectFlags) ConnectFlags {
	return i ^ m
}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[CONNECT_AFTER-1]
	_ = x[CONNECT_SWAPPED-2]
}

const _ConnectFlags_name = "CONNECT_AFTERCONNECT_SWAPPED"

var _ConnectFlags_index = [...]uint8{0, 13, 28}

func (i ConnectFlags) String() string {
	i -= 1
	if i >= ConnectFlags(len(_ConnectFlags_index)-1) {
		return "ConnectFlags(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _ConnectFlags_name[_ConnectFlags_index[i]:_ConnectFlags_index[i+1]]
}
