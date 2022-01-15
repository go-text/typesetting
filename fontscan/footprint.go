package fontscan

import "github.com/benoitkugler/textlayout/fonts"

type Footprint struct {
	Family string

	Location fonts.FaceID

	Runes RuneSet // supported runes
	Langs LangSet // supported languages

	// TODO:

}
