package fontscan

import (
	"errors"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
)

// name values corresponding to the xxxConsts arrays
var (
	styleStrings   [len(styleConsts)]string
	weightStrings  [len(weightConsts)]string
	stretchStrings [len(stretchConsts)]string
)

func init() {
	for i, v := range styleConsts {
		styleStrings[i] = v.name
	}
	for i, v := range weightConsts {
		weightStrings[i] = v.name
	}
	for i, v := range stretchConsts {
		stretchStrings[i] = v.name
	}
}

var styleConsts = [...]struct {
	name  string
	value fonts.Style
}{
	{"italic", fonts.StyleItalic},
	{"kursiv", fonts.StyleItalic},
	{"oblique", fonts.StyleOblique},
}

var weightConsts = [...]struct {
	name  string
	value fonts.Weight
}{
	{"thin", fonts.WeightThin},
	{"extralight", fonts.WeightExtraLight},
	{"ultralight", fonts.WeightExtraLight},
	{"light", fonts.WeightLight},
	{"demilight", (fonts.WeightLight + fonts.WeightNormal) / 2},
	{"semilight", (fonts.WeightLight + fonts.WeightNormal) / 2},
	{"book", fonts.WeightNormal - 20},
	{"regular", fonts.WeightNormal},
	{"normal", fonts.WeightNormal},
	{"medium", fonts.WeightMedium},
	{"demibold", fonts.WeightSemibold},
	{"demi", fonts.WeightSemibold},
	{"semibold", fonts.WeightSemibold},
	{"extrabold", fonts.WeightExtraBold},
	{"superbold", fonts.WeightExtraBold},
	{"ultrabold", fonts.WeightExtraBold},
	{"bold", fonts.WeightBold},
	{"ultrablack", fonts.WeightBlack + 20},
	{"superblack", fonts.WeightBlack + 20},
	{"extrablack", fonts.WeightBlack + 20},
	{"black", fonts.WeightBlack},
	{"heavy", fonts.WeightBlack},
}

var stretchConsts = [...]struct {
	name  string
	value fonts.Stretch
}{
	{"ultracondensed", fonts.StretchUltraCondensed},
	{"extracondensed", fonts.StretchExtraCondensed},
	{"semicondensed", fonts.StretchSemiCondensed},
	{"condensed", fonts.StretchCondensed}, /* must be after *condensed */
	{"normal", fonts.StretchNormal},
	{"semiexpanded", fonts.StretchSemiExpanded},
	{"extraexpanded", fonts.StretchExtraExpanded},
	{"ultraexpanded", fonts.StretchUltraExpanded},
	{"expanded", fonts.StretchExpanded}, /* must be after *expanded */
	{"extended", fonts.StretchExpanded},
}

// Aspect stores the properties that specify which font in a family to use:
// style, weight, and stretchiness.
type Aspect struct {
	Style   Style
	Weight  Weight
	Stretch Stretch
}

type (
	Style   = fonts.Style
	Weight  = fonts.Weight
	Stretch = fonts.Stretch
)

// some fonts includes aspect information in a string description,
// usually called "style"
// inferFromStyle scans such a string and fills the missing fields,
// eventually defaulting to "regular" values : StyleNormal, WeightNormal, StretchNormal
func (as *Aspect) inferFromStyle(additionalStyle string) {
	additionalStyle = ignoreBlanksAndCase(additionalStyle)

	if as.Style == 0 {
		if index := stringContainsConst(additionalStyle, styleStrings[:]); index != -1 {
			as.Style = styleConsts[index].value
		} else {
			as.Style = fonts.StyleNormal
		}
	}

	if as.Weight == 0 {
		if index := stringContainsConst(additionalStyle, weightStrings[:]); index != -1 {
			as.Weight = weightConsts[index].value
		} else {
			as.Weight = fonts.WeightNormal
		}
	}

	if as.Stretch == 0 {
		if index := stringContainsConst(additionalStyle, stretchStrings[:]); index != -1 {
			as.Stretch = stretchConsts[index].value
		} else {
			as.Stretch = fonts.StretchNormal
		}
	}
}

var rp = strings.NewReplacer(" ", "", "\t", "")

func ignoreBlanksAndCase(s1 string) string { return rp.Replace(strings.ToLower(s1)) }

// returns the index in `constants` of a constant contained in `str`,
// or -1
func stringContainsConst(str string, constants []string) int {
	for i, c := range constants {
		if strings.Contains(str, c) {
			return i
		}
	}
	return -1
}

const aspectSize = 1 + 4 + 4

// serializeTo serialize the Aspect in binary format
func (as Aspect) serialize() []byte {
	var buffer [aspectSize]byte
	buffer[0] = byte(as.Style)
	serializeFloat(float32(as.Weight), buffer[1:])
	serializeFloat(float32(as.Stretch), buffer[5:])
	return buffer[:]
}

// deserializeFrom reads the binary format produced by serializeTo
// it returns the number of bytes read from `data`
func (as *Aspect) deserializeFrom(data []byte) (int, error) {
	if len(data) < aspectSize {
		return 0, errors.New("invalid Aspect (EOF)")
	}
	as.Style = Style(data[0])
	as.Weight = Weight(deserializeFloat(data[1:]))
	as.Stretch = Stretch(deserializeFloat(data[5:]))
	return aspectSize, nil
}
