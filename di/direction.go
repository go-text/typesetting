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

const (
	progression Direction = 1 << iota
	// BAxisVertical is the bit for the axis, 0 for horizontal, 1 for vertical
	BAxisVertical

	// If this flag is set, the orientation is chosen
	// using the [BVerticalUpright] flag.
	// Otherwise, the segmenter will resolve the orientation based
	// on unicode properties
	BVerticalOrientationSet
	// BVerticalSideways is set for 'sideways', unset for 'upright'
	BVerticalSideways
)

// IsVertical returns whether d is laid out on a vertical
// axis. If the return value is false, d is on the horizontal
// axis.
func (d Direction) IsVertical() bool { return d&BAxisVertical != 0 }

// Axis returns the layout axis for d.
func (d Direction) Axis() Axis {
	if d.IsVertical() {
		return Vertical
	}
	return Horizontal
}

// Progression returns the text layout progression for d.
func (d Direction) Progression() Progression {
	if d&progression == 0 {
		return FromTopLeft
	}
	return TowardTopLeft
}

// Axis indicates the axis of layout for a piece of text.
type Axis bool

const (
	Horizontal Axis = false
	Vertical   Axis = true
)

// Progression indicates how text is read within its Axis relative
// to the top left corner.
type Progression bool

const (
	// FromTopLeft indicates text in which a reader starts reading
	// at the top left corner of the text and moves away from it.
	// DirectionLTR and DirectionTTB are examples of FromTopLeft
	// Progression.
	FromTopLeft Progression = false
	// TowardTopLeft indicates text in which a reader starts reading
	// at the opposite end of the text's Axis from the top left corner
	// and moves towards it. DirectionRTL and DirectionBTT are examples
	// of TowardTopLeft progression.
	TowardTopLeft Progression = true
)

// Orientation describes a glyph orientation.
//
// When shaping vertical text, some glyphs are rotated
// by 90°. This flag should be used by renderers to also
// rotate the glyph when drawing.

// IsSideways returns true if the direction has a 'sideways' vertical
// orientation.
func (d Direction) IsSideways() bool { return d&BVerticalSideways != 0 }
