package cdk

// MouseFlags are options to modify the handling of mouse events.
// Actual events can be or'd together.
type MouseFlags int

const (
	MouseButtonEvents = MouseFlags(1) // Click events only
	MouseDragEvents   = MouseFlags(2) // Click-drag events (includes button events)
	MouseMotionEvents = MouseFlags(4) // All mouse events (includes click and drag events)
)
