package di

import (
	"testing"

	"github.com/go-text/typesetting/harfbuzz"
	tu "github.com/go-text/typesetting/testutils"
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

	tu.Assert(t, !DirectionLTR.HasVerticalOrientation())
	tu.Assert(t, !DirectionRTL.HasVerticalOrientation())
	tu.Assert(t, !DirectionTTB.HasVerticalOrientation())
	tu.Assert(t, !DirectionBTT.HasVerticalOrientation())

	for _, test := range []struct {
		sideways    bool
		progression Progression
		hb          harfbuzz.Direction
	}{
		{true, FromTopLeft, harfbuzz.TopToBottom},
		{true, TowardTopLeft, harfbuzz.BottomToTop},
		{false, FromTopLeft, harfbuzz.TopToBottom},
		{false, TowardTopLeft, harfbuzz.BottomToTop},
	} {
		d := axisVertical
		d.SetProgression(test.progression)

		tu.Assert(t, !d.HasVerticalOrientation())
		d.SetSideways(test.sideways)

		tu.Assert(t, d.HasVerticalOrientation())
		tu.Assert(t, d.IsSideways() == test.sideways)
		tu.Assert(t, d.Axis() == Vertical)
		tu.Assert(t, d.Progression() == test.progression)
		tu.Assert(t, d.Harfbuzz() == test.hb)
	}
}
