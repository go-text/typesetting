package di

// Direction indicates the layout direction of a piece of text.
type Direction uint8

const (
	// DirectionLTR is for Left-to-Right text.
	DirectionLTR Direction = iota
	// DirectionRTL is for Right-to-Left text.
	DirectionRTL
	// DirectionTTB is for Top-to-Bottom text.
	DirectionTTB
	// DirectionBTT is for Bottom-to-Top text.
	DirectionBTT
)

// IsVertical returns whether d is laid out on a vertical
// axis. If the return value is false, d is on the horizontal
// axis.
func (d Direction) IsVertical() bool {
	return d == DirectionBTT || d == DirectionTTB
}

// Axis returns the layout axis for d.
func (d Direction) Axis() Axis {
	switch d {
	case DirectionBTT, DirectionTTB:
		return Vertical
	default:
		return Horizontal
	}
}

// Axis indicates the axis of layout for a piece of text.
type Axis bool

const (
	Horizontal Axis = false
	Vertical   Axis = true
)
