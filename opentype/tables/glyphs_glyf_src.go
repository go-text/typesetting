// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// shared with gvar, sbix, eblc
// return an error only if data is not long enough
func ParseLoca(src []byte, numGlyphs int, isLong bool) (out []uint32, err error) {
	var size int
	if isLong {
		size = (numGlyphs + 1) * 4
	} else {
		size = (numGlyphs + 1) * 2
	}
	if L := len(src); L < size {
		return nil, fmt.Errorf("reading Loca: EOF: expected length: %d, got %d", size, L)
	}
	out = make([]uint32, numGlyphs+1)
	if isLong {
		for i := range out {
			out[i] = binary.BigEndian.Uint32(src[4*i:])
		}
	} else {
		for i := range out {
			out[i] = 2 * uint32(binary.BigEndian.Uint16(src[2*i:])) // The actual local offset divided by 2 is stored.
		}
	}
	return out, nil
}

// Glyph Data
type Glyf []Glyph

// ParseGlyf parses the 'glyf' table.
// locaOffsets has length numGlyphs + 1, and is returned by ParseLoca
func ParseGlyf(src []byte, locaOffsets []uint32) (Glyf, error) {
	out := make(Glyf, len(locaOffsets)-1)
	var err error
	for i := range out {
		start, end := locaOffsets[i], locaOffsets[i+1]
		// If a glyph has no outline, then loca[n] = loca [n+1].
		if start == end {
			continue
		}
		out[i], _, err = ParseGlyph(src[start:end])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

type Glyph struct {
	numberOfContours int16     // If the number of contours is greater than or equal to zero, this is a simple glyph. If negative, this is a composite glyph — the value -1 should be used for composite glyphs.
	XMin             int16     // Minimum x for coordinate data.
	YMin             int16     // Minimum y for coordinate data.
	XMax             int16     // Maximum x for coordinate data.
	YMax             int16     // Maximum y for coordinate data.
	Data             GlyphData `isOpaque:"" subsliceStart:"AtCurrent"`
}

func (gl *Glyph) parseData(src []byte) (err error) {
	if gl.numberOfContours >= 0 { // simple glyph
		gl.Data, _, err = ParseSimpleGlyph(src, int(gl.numberOfContours))
	} else { // composite glyph
		gl.Data, _, err = ParseCompositeGlyph(src)
	}
	return err
}

type GlyphData interface {
	isGlyphData()
}

func (SimpleGlyph) isGlyphData()    {}
func (CompositeGlyph) isGlyphData() {}

type SimpleGlyph struct {
	EndPtsOfContours []uint16 // [numberOfContours] Array of point indices for the last point of each contour, in increasing numeric order.
	Instructions     []byte   `arrayCount:"FirstUint16"` // [instructionLength] Array of instruction byte code for the glyph.

	pointsData simpleGlyphData `isOpaque:"" subsliceStart:"AtCurrent"`
}

// Len return the number of contours points.
// It is the same as len(Points()), but is more efficient.
func (sg SimpleGlyph) Len() int {
	if len(sg.EndPtsOfContours) == 0 { // nothing to do
		return 0
	}

	return int(sg.EndPtsOfContours[len(sg.EndPtsOfContours)-1]) + 1
}

const (
	xShortVector                  = 0x02
	xIsSameOrPositiveXShortVector = 0x10
	yShortVector                  = 0x04
	yIsSameOrPositiveYShortVector = 0x20
)

const repeatFlag = 0x08

// to avoid costly length check at run time, we precompute the expected data size for coordinates
func (sg *SimpleGlyph) parsePointsData(src []byte, _ int) error {
	numPoints := sg.Len()

	var (
		coordinatesLengthX, coordinatesLengthY int
		cursor                                 int
		L                                      = len(src)
	)
	for i := 0; i < numPoints; i++ {
		if L <= cursor {
			return errors.New("invalid simple glyph data flags (EOF)")
		}
		flag := src[cursor]
		cursor++

		localLengthX, localLengthY := 0, 0
		if flag&xShortVector != 0 {
			localLengthX = 1
		} else if flag&xIsSameOrPositiveXShortVector == 0 {
			localLengthX = 2
		}
		if flag&yShortVector != 0 {
			localLengthY = 1
		} else if flag&yIsSameOrPositiveYShortVector == 0 {
			localLengthY = 2
		}

		if flag&repeatFlag != 0 {
			if L <= cursor {
				return errors.New("invalid simple glyph data flags (EOF)")
			}
			repeatCount := int(src[cursor])
			cursor++
			if i+repeatCount+1 > numPoints { // gracefully handle out of bounds
				repeatCount = numPoints - i - 1
			}
			i += repeatCount
			localLengthX += repeatCount * localLengthX
			localLengthY += repeatCount * localLengthY
		}

		coordinatesLengthX += localLengthX
		coordinatesLengthY += localLengthY
	}

	endData := cursor + coordinatesLengthX + coordinatesLengthY
	if L, E := len(src), endData; L < E {
		return fmt.Errorf("EOF: expected length: %d, got %d", E, L)
	}

	startDataX, startDataY := cursor, cursor+coordinatesLengthX
	sg.pointsData = simpleGlyphData{data: src[:endData], startDataX: startDataX, startDataY: startDataY}

	return nil
}

type GlyphContourPoint struct {
	Flag uint8
	X, Y int16
}

// Points decodes the encoded data, returning the coordinates and flag of the contour points.
func (sg SimpleGlyph) Points() []GlyphContourPoint {
	if len(sg.EndPtsOfContours) == 0 { // nothing to do
		return nil
	}

	numPoints := int(sg.EndPtsOfContours[len(sg.EndPtsOfContours)-1]) + 1
	points := make([]GlyphContourPoint, numPoints)

	src := sg.pointsData.data
	var cursor int
	// decompress flags
	for i := 0; i < numPoints; i++ {
		flag := src[cursor]
		points[i].Flag = flag
		cursor++

		if flag&repeatFlag != 0 {
			repeatCount := int(src[cursor])
			cursor++
			if i+repeatCount+1 > numPoints { // gracefully handle out of bounds
				repeatCount = numPoints - i - 1
			}
			subSlice := points[i+1 : i+repeatCount+1]
			for j := range subSlice {
				subSlice[j].Flag = flag
			}
			i += repeatCount
		}
	}

	// read x and y coordinates
	dataX, dataY := src[sg.pointsData.startDataX:sg.pointsData.startDataY], src[sg.pointsData.startDataY:]
	parseGlyphContourPoints(dataX, dataY, points)

	return points
}

// returns the position after the read and the relative coordinate
// the input slice has already been checked for length
func readContourPoint(flag byte, data []byte, pos int, shortFlag, sameFlag uint8) (int, int16) {
	var v int16
	if flag&shortFlag != 0 {
		val := data[pos]
		pos++
		if flag&sameFlag != 0 {
			v += int16(val)
		} else {
			v -= int16(val)
		}
	} else if flag&sameFlag == 0 {
		val := binary.BigEndian.Uint16(data[pos:])
		pos += 2
		v += int16(val)
	}
	return pos, v
}

// update the points in place
func parseGlyphContourPoints(dataX, dataY []byte, points []GlyphContourPoint) {
	var (
		posX, posY               int   // position into data
		vX, offsetX, vY, offsetY int16 // coordinates are relative to the previous
	)
	for i, p := range points {
		posX, offsetX = readContourPoint(p.Flag, dataX, posX, xShortVector, xIsSameOrPositiveXShortVector)
		vX += offsetX
		points[i].X = vX

		posY, offsetY = readContourPoint(p.Flag, dataY, posY, yShortVector, yIsSameOrPositiveYShortVector)
		vY += offsetY
		points[i].Y = vY
	}
}

type CompositeGlyph struct {
	glyphs       []byte `isOpaque:""`
	nbGlyphs     uint16 `isOpaque:""`
	Instructions []byte `isOpaque:""`
}

const (
	arg1And2AreWords = 1 << iota
	_
	_
	weHaveAScale
	_
	moreComponents
	weHaveAnXAndYScale
	weHaveATwoByTwo
	weHaveInstructions
)

// we store the condensed form, but we process it to accelerate
// runtime access.
func (cg *CompositeGlyph) parseGlyphs(src []byte) error {
	var (
		flags  uint16
		cursor int
	)
	for do := true; do; do = flags&moreComponents != 0 {
		if L := len(src[cursor:]); L < 4 {
			return fmt.Errorf("EOF: expected length: %d, got %d", 4, L)
		}
		flags = binary.BigEndian.Uint16(src[cursor:])

		if flags&arg1And2AreWords != 0 { // 16 bits
			if L, E := len(src[cursor:]), 4+4; L < E {
				return fmt.Errorf("EOF: expected length: %d, got %d", E, L)
			}
			cursor += 8
		} else {
			if L, E := len(src[cursor:]), 4+2; L < E {
				return fmt.Errorf("EOF: expected length: %d, got %d", E, L)
			}
			cursor += 6
		}

		if flags&weHaveAScale != 0 {
			if L := len(src[cursor:]); L < 2 {
				return fmt.Errorf("EOF: expected length: %d, got %d", 2, L)
			}
			cursor += 2
		} else if flags&weHaveAnXAndYScale != 0 {
			if L := len(src[cursor:]); L < 4 {
				return fmt.Errorf("EOF: expected length: %d, got %d", 4, L)
			}
			cursor += 4
		} else if flags&weHaveATwoByTwo != 0 {
			if L := len(src[cursor:]); L < 8 {
				return fmt.Errorf("EOF: expected length: %d, got %d", 8, L)
			}
			cursor += 8
		}
		cg.nbGlyphs++
	}

	cg.glyphs = src[:cursor]

	if flags&weHaveInstructions != 0 {
		if L := len(src[cursor:]); L < 2 {
			return fmt.Errorf("EOF: expected length: 2, got %d", L)
		}
		E := int(binary.BigEndian.Uint16(src[cursor:]))
		if L := len(src[cursor:]); L < E {
			return fmt.Errorf("EOF: expected length: %d, got %d", E, len(src[cursor:]))
		}
		cg.Instructions = src[cursor : cursor+E]
	}

	return nil
}

// Len returns the number of composite parts in this glyph.
// It is the same as len(Parts()) but is more efficient.
func (cg CompositeGlyph) Len() int { return int(cg.nbGlyphs) }

// Parts decodes the glyph part of the composite glyph.
func (cg *CompositeGlyph) Parts() []CompositeGlyphPart {
	out := make([]CompositeGlyphPart, cg.nbGlyphs)

	src := cg.glyphs
	for i := range out {
		part := &out[i]

		flags := binary.BigEndian.Uint16(src)
		part.Flags = flags
		part.GlyphIndex = GlyphID(binary.BigEndian.Uint16(src[2:]))

		if flags&arg1And2AreWords != 0 { // 16 bits
			part.arg1 = binary.BigEndian.Uint16(src[4:])
			part.arg2 = binary.BigEndian.Uint16(src[6:])
			src = src[8:]
		} else {
			part.arg1 = uint16(src[4])
			part.arg2 = uint16(src[5])
			src = src[6:]
		}

		part.Scale[0], part.Scale[3] = 1, 1
		if flags&weHaveAScale != 0 {
			part.Scale[0] = Float214FromUint(binary.BigEndian.Uint16(src))
			part.Scale[3] = part.Scale[0]
			src = src[2:]
		} else if flags&weHaveAnXAndYScale != 0 {
			part.Scale[0] = Float214FromUint(binary.BigEndian.Uint16(src))
			part.Scale[3] = Float214FromUint(binary.BigEndian.Uint16(src[2:]))
			src = src[4:]
		} else if flags&weHaveATwoByTwo != 0 {
			part.Scale[0] = Float214FromUint(binary.BigEndian.Uint16(src))
			part.Scale[1] = Float214FromUint(binary.BigEndian.Uint16(src[2:]))
			part.Scale[2] = Float214FromUint(binary.BigEndian.Uint16(src[4:]))
			part.Scale[3] = Float214FromUint(binary.BigEndian.Uint16(src[6:]))
			src = src[8:]
		}

	}

	return out
}

// already handled in parseGlyphs
func (cg *CompositeGlyph) parseNbGlyphs(src []byte) error     { return nil }
func (cg *CompositeGlyph) parseInstructions(src []byte) error { return nil }

type CompositeGlyphPart struct {
	Flags      uint16
	GlyphIndex GlyphID

	// raw value before interpretation:
	// arg1 and arg2 may be either :
	//	- unsigned, when used as indices into the contour point list
	//    (see ArgsAsIndices)
	//  - signed, when used as translation in the transformation matrix
	//	  (see ArgsAsTranslation)
	arg1, arg2 uint16

	// Scale is a matrix x, 01, 10, y ; default to identity
	Scale [4]float32
}

func (c *CompositeGlyphPart) HasUseMyMetrics() bool {
	const useMyMetrics = 0x0200
	return c.Flags&useMyMetrics != 0
}

// return true if arg1 and arg2 indicated an anchor point,
// not offsets
func (c *CompositeGlyphPart) IsAnchored() bool {
	const argsAreXyValues = 0x0002
	return c.Flags&argsAreXyValues == 0
}

func (c *CompositeGlyphPart) IsScaledOffsets() bool {
	const (
		scaledComponentOffset   = 0x0800
		unscaledComponentOffset = 0x1000
	)
	return c.Flags&(scaledComponentOffset|unscaledComponentOffset) == scaledComponentOffset
}

func (c *CompositeGlyphPart) ArgsAsTranslation() (int16, int16) {
	// arg1 and arg2 are interpreted as signed integers here
	// the conversion depends on the original size (8 or 16 bits)
	if c.Flags&arg1And2AreWords != 0 {
		return int16(c.arg1), int16(c.arg2)
	}
	return int16(int8(uint8(c.arg1))), int16(int8(uint8(c.arg2)))
}

func (c *CompositeGlyphPart) ArgsAsIndices() (int, int) {
	// arg1 and arg2 are interpreted as unsigned integers here
	return int(c.arg1), int(c.arg2)
}
