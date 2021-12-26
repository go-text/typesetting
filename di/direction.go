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

// Progression returns the text layout progression for d.
func (d Direction) Progression() Progression {
	switch d {
	case DirectionTTB, DirectionLTR:
		return FromUpperLeft
	default:
		return TowardUpperLeft
	}
}

// Axis indicates the axis of layout for a piece of text.
type Axis bool

const (
	Horizontal Axis = false
	Vertical   Axis = true
)

// Progression indicates how text is read within its Axis relative
// to the origin (where the origin is the upper-left corner of the
// text).
type Progression bool

const (
	// FromUpperLeft indicates text in which a reader starts reading
	// at the upper-left corner of the text and moves away from it.
	// DirectionLTR and DirectionTTB are examples of FromUpperLeft
	// Progression.
	FromUpperLeft Progression = false
	// TowardUpperLeft indicates text in which a reader starts reading
	// at the opposite end of the text's Axis from the upper left corner
	// and moves towards it. DirectionRTL and DirectionBTT are examples
	// of TowardUpperLeft progression.
	TowardUpperLeft Progression = true
)
