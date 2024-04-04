package harfbuzz

import (
	"testing"

	ot "github.com/go-text/typesetting/font/opentype"
)

func TestOTFeature(t *testing.T) {
	face := openFontFile(t, "fonts/cv01.otf")

	cv01 := ot.NewTag('c', 'v', '0', '1')

	featureIndex := findFeatureForLang(&face.GSUB.Layout, 0, DefaultLanguageIndex, cv01)
	if featureIndex == NoFeatureIndex {
		t.Fatal("failed to find feature index")
	}
}
