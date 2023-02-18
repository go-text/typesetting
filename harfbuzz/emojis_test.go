package harfbuzz

import (
	"fmt"
	"strings"
	"testing"
)

func TestEmojisSequences(t *testing.T) {
	for _, sequence := range emojisSequences {
		var runes []string
		for _, r := range sequence {
			runes = append(runes, fmt.Sprintf("U+%X", r))
		}
		clusters := strings.Repeat("|1=0", len(sequence))[1:]
		test := fmt.Sprintf("fonts/AdobeBlank2.ttf;--no-glyph-names --no-positions;%s;[%s]", strings.Join(runes, ","), clusters)

		testD := newTestData(t, ".", test)
		runShapingTest(t, testD, false)
	}
}
