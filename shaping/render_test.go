package shaping

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/font/opentype"
)

// this file implements a very primitive "rasterizer", which can
// be used to visually inspect shaper outputs.

func drawVLine(img *image.RGBA, start image.Point, height int, c color.RGBA) {
	for y := start.Y; y <= start.Y+height; y++ {
		img.SetRGBA(start.X, y, c)
	}
}

func drawHLine(img *image.RGBA, start image.Point, width int, c color.RGBA) {
	for x := start.X; x <= start.X+width; x++ {
		img.SetRGBA(x, start.Y, c)
	}
}

func drawRect(img *image.RGBA, min, max image.Point, c color.RGBA) {
	for x := min.X; x <= max.X; x++ {
		for y := min.Y; y <= max.Y; y++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func drawPoint(img *image.RGBA, pt image.Point, c color.RGBA) {
	drawRect(img, pt.Add(image.Pt(-1, -1)), pt.Add(image.Pt(1, 1)), c)
}

// dot includes the offset
func drawGlyph(out *Output, img *image.RGBA, dot image.Point, outlines font.GlyphOutline, c color.RGBA) {
	var current font.SegmentPoint
	for _, seg := range outlines.Segments {
		points := seg.ArgsSlice()
		for _, point := range points {
			x, y := out.FromFontUnit(point.X).Round(), -out.FromFontUnit(point.Y).Round()
			drawPoint(img, dot.Add(image.Pt(x, y)), c)
		}

		last := points[len(points)-1]

		if seg.Op == opentype.SegmentOpLineTo {
			for t := float32(0); t < 1; t += 0.2 {
				middleX := out.FromFontUnit(t*current.X + (1-t)*last.X).Round()
				middleY := -out.FromFontUnit(t*current.Y + (1-t)*last.Y).Round()
				drawPoint(img, dot.Add(image.Pt(middleX, middleY)), c)
			}
		}

		current = last
	}
}

var (
	red   = color.RGBA{R: 0xFF, A: 0xFF}
	green = color.RGBA{R: 0xCF, G: 0xFF, B: 0xCF, A: 0xFF}
	black = color.RGBA{A: 0xFF}
)

func imageDims(line []Output) (width, height, baseline int) {
	firstRun := line[0]
	if firstRun.Direction.IsVertical() {
		ascent, descent := 0, 0
		for _, run := range line {
			if a := run.GlyphBounds.Ascent.Round(); a > ascent {
				ascent = a
			}
			if d := run.GlyphBounds.Descent.Round(); d < descent {
				descent = d
			}
			height += -run.Advance.Round()
		}
		baseline = -descent
		width = ascent - descent
	} else {
		ascent, descent := 0, 0
		for _, run := range line {
			if a := run.GlyphBounds.Ascent.Round(); a > ascent {
				ascent = a
			}
			if d := run.GlyphBounds.Descent.Round(); d < descent {
				descent = d
			}
			width += run.Advance.Round()
		}
		baseline = ascent
		height = ascent - descent
	}
	return
}

func drawTextLine(runs []Output, file string) error {
	width, height, baseline := imageDims(runs)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// white background
	draw.Draw(img, img.Rect, image.NewUniform(color.White), image.Point{}, draw.Src)

	if runs[0].Direction.IsVertical() {
		// draw the baseline
		drawVLine(img, image.Pt(baseline, 0), height, black)

		dot := image.Pt(baseline, 0)
		for _, run := range runs {
			dot = drawVRun(run, img, dot)
		}
	} else {
		// draw the baseline
		drawHLine(img, image.Pt(0, baseline), width, black)

		dot := image.Pt(0, baseline)
		for _, run := range runs {
			dot = drawHRun(run, img, dot)
		}
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	if err = png.Encode(f, img); err != nil {
		return err
	}
	err = f.Close()
	return err
}

// assume horizontal direction
func drawHRun(out Output, img *image.RGBA, dot image.Point) image.Point {
	for _, g := range out.Glyphs {
		// image has Y axis pointing down
		dotWithOffset := dot.Add(image.Pt(g.XOffset.Round(), -g.YOffset.Round()))

		minX := dotWithOffset.X + g.XBearing.Round()
		maxX := minX + g.Width.Round()
		minY := dotWithOffset.Y - g.YBearing.Round()
		maxY := minY - g.Height.Round()

		drawRect(img, image.Pt(minX, minY), image.Pt(maxX, maxY), green)

		// draw the dot
		drawPoint(img, dot, black)

		// draw a sketch of the glyphs
		glyphData := out.Face.GlyphData(g.GlyphID).(font.GlyphOutline)
		drawGlyph(&out, img, dotWithOffset, glyphData, black)

		dot.X += g.XAdvance.Round()
		// draw the advance
		drawVLine(img, image.Pt(dot.X, 0), img.Bounds().Dy(), red)
	}

	return dot
}

// assume vertical direction
func drawVRun(out Output, img *image.RGBA, dot image.Point) image.Point {
	for _, g := range out.Glyphs {
		// image has Y axis pointing down
		dotWithOffset := dot.Add(image.Pt(g.XOffset.Round(), -g.YOffset.Round()))

		minX := dotWithOffset.X + g.XBearing.Round()
		maxX := minX + g.Width.Round()
		minY := dotWithOffset.Y - g.YBearing.Round()
		maxY := minY - g.Height.Round()

		drawRect(img, image.Pt(minX, minY), image.Pt(maxX, maxY), green)

		// draw the dot
		drawPoint(img, dot, black)

		// draw a sketch of the glyphs
		glyphData := out.Face.GlyphData(g.GlyphID).(font.GlyphOutline)
		if out.Direction.IsSideways() {
			glyphData.Sideways(out.ToFontUnit(-g.YOffset))
		}
		drawGlyph(&out, img, dotWithOffset, glyphData, black)

		dot.Y += -g.YAdvance.Round()

		// draw the advance
		drawHLine(img, image.Pt(0, dot.Y), img.Bounds().Dx(), red)
	}

	return dot
}
