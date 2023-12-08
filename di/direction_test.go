package di

import (
	"testing"

	"github.com/go-text/typesetting/harfbuzz"
	tu "github.com/go-text/typesetting/opentype/testutils"
)

func TestDirection(t *testing.T) {
	tu.Assert(t, DirectionLTR.Axis() == Horizontal)
	tu.Assert(t, DirectionRTL.Axis() == Horizontal)
	tu.Assert(t, DirectionTTB.Axis() == Vertical)
	tu.Assert(t, DirectionBTT.Axis() == Vertical)
	tu.Assert(t, !DirectionLTR.IsVertical())
	tu.Assert(t, !DirectionRTL.IsVertical())
	tu.Assert(t, DirectionTTB.IsVertical())
	tu.Assert(t, DirectionBTT.IsVertical())

	tu.Assert(t, DirectionLTR.Progression() == FromTopLeft)
	tu.Assert(t, DirectionRTL.Progression() == TowardTopLeft)
	tu.Assert(t, DirectionTTB.Progression() == FromTopLeft)
	tu.Assert(t, DirectionBTT.Progression() == TowardTopLeft)

	tu.Assert(t, !DirectionTTB.IsSideways())
	tu.Assert(t, !DirectionBTT.IsSideways())

	tu.Assert(t, DirectionLTR.SwitchAxis() == DirectionTTB)
	tu.Assert(t, DirectionRTL.SwitchAxis() == DirectionBTT)
	tu.Assert(t, DirectionTTB.SwitchAxis() == DirectionLTR)
	tu.Assert(t, DirectionBTT.SwitchAxis() == DirectionRTL)

	tu.Assert(t, DirectionLTR.Harfbuzz() == harfbuzz.LeftToRight)
	tu.Assert(t, DirectionRTL.Harfbuzz() == harfbuzz.RightToLeft)
	tu.Assert(t, DirectionTTB.Harfbuzz() == harfbuzz.TopToBottom)
	tu.Assert(t, DirectionBTT.Harfbuzz() == harfbuzz.BottomToTop)

	d := DirectionTTB
	d.SetSideways()
	tu.Assert(t, d.IsSideways())
	tu.Assert(t, d.Axis() == Vertical)
	tu.Assert(t, d.Progression() == FromTopLeft)
	tu.Assert(t, d.Harfbuzz() == harfbuzz.TopToBottom)
	d = DirectionBTT
	d.SetSideways()
	tu.Assert(t, d.IsSideways())
	tu.Assert(t, d.Axis() == Vertical)
	tu.Assert(t, d.Progression() == TowardTopLeft)
	tu.Assert(t, d.Harfbuzz() == harfbuzz.BottomToTop)
}
