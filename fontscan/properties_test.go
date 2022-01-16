package fontscan

import (
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func TestAspect_inferFromStyle(t *testing.T) {
	styn, wn, sten := fonts.StyleNormal, fonts.WeightNormal, fonts.StretchNormal
	tests := []struct {
		args   string
		fields Aspect
		want   Aspect
	}{
		{
			"", Aspect{styn, wn, sten}, Aspect{styn, wn, sten}, // no op
		},
		{
			"", Aspect{0, 0, 0}, Aspect{styn, wn, sten}, // default values
		},
		{
			"Black", Aspect{0, 0, 0}, Aspect{styn, fonts.WeightBlack, sten},
		},
		{
			"conDensed", Aspect{0, 0, 0}, Aspect{styn, wn, fonts.StretchCondensed},
		},
		{
			"ITALIC", Aspect{0, 0, 0}, Aspect{fonts.StyleItalic, wn, sten},
		},
		{
			"black", Aspect{0, fonts.WeightNormal, 0}, Aspect{styn, fonts.WeightNormal, sten}, // respect initial value
		},
		{
			"black oblique", Aspect{0, 0, 0}, Aspect{fonts.StyleOblique, fonts.WeightBlack, sten},
		},
	}
	for _, tt := range tests {
		as := tt.fields
		as.inferFromStyle(tt.args)
		if as != tt.want {
			t.Fatalf("unexpected aspect %v", as)
		}
	}
}
