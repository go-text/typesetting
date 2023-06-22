package harfbuzz

import (
	"testing"

	"github.com/go-text/typesetting/opentype/loader"
)

func TestOTFeature(t *testing.T) {
	face := openFontFile(t, "fonts/cv01.otf")

	cv01 := loader.NewTag('c', 'v', '0', '1')

	featureIndex := findFeatureForLang(&face.GSUB.Layout, 0, DefaultLanguageIndex, cv01)
	if featureIndex == NoFeatureIndex {
		t.Fatal("failed to find feature index")
	}
}
