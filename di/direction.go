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

// IsVertical returns the layout axis for d. The return value can be
// used as a value of type axis or as a boolean representing
// whether the text is vertical.
//
//     if d.IsVertical() {
//     } else {
//         // Must be Horizontal.
//     }
//
//     switch d.IsVertical() {
//         case Vertical:
//         case Horzontal:
//     }
//
func (d Direction) IsVertical() Axis {
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
