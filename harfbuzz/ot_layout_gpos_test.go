package harfbuzz

import (
	"testing"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/font/opentype/tables"
)

func TestApplyGPOSMarksRejectsMalformedBaseAnchor(t *testing.T) {
	for _, test := range []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"short format", []byte{0}},
		{"unknown format", []byte{0, 4, 0, 10, 0, 20, 0, 0, 0, 0}},
		{"short format 1", []byte{0, 1}},
		{"short format 2", []byte{0, 2}},
		{"short format 3", []byte{0, 3}},
		{"bad device offset", []byte{0, 3, 0, 10, 0, 20, 0, 30, 0, 0}},
	} {
		t.Run(test.name, func(t *testing.T) {
			base, _, err := tables.ParseBaseArray(append([]byte{0, 1, 0, 4}, test.data...), 1)
			if err != nil {
				t.Fatal(err)
			}
			buffer := NewBuffer()
			buffer.Info = []GlyphInfo{{Glyph: 1}, {Glyph: 2}}
			buffer.Pos = make([]GlyphPosition, 2)
			buffer.idx = 1
			c := otApplyContext{buffer: buffer, font: &Font{
				face: font.NewFace(&font.Font{}), faceUpem: 1000, XScale: 1000, YScale: 1000,
			}}
			marks := tables.MarkArray{
				MarkRecords: []tables.MarkRecord{{MarkClass: 0}},
				MarkAnchors: []tables.Anchor{tables.AnchorFormat1{XCoordinate: 10, YCoordinate: 20}},
			}
			if c.applyGPOSMarks(marks, 0, 0, base.Anchors(), 0) {
				t.Fatal("malformed anchor prevented subsequent subtables from applying")
			}
			if buffer.idx != 1 || buffer.Pos[1] != (GlyphPosition{}) {
				t.Fatal("rejected anchor changed the buffer")
			}
		})
	}
}
