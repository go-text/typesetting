package font

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

type Font = *truetype.Font
type Resource = fonts.Resource

func ParseTTF(file Resource, loadAllTables bool) (Font, error){
	return truetype.Parse(file, loadAllTables)
}
