package bidi

import (
	"fmt"

	ucd "github.com/go-text/typesetting/internal/unicodedata"
)

// Analysis holds immutable bidi facts for one paragraph.
type Analysis struct {
	types  []ucd.BidiClass
	levels []Level
	base   Level
}

// Len returns the number of analyzed runes, including the paragraph separator.
func (a Analysis) Len() int { return len(a.levels) }

// BaseDirection returns the resolved paragraph direction.
func (a Analysis) BaseDirection() Direction {
	if a.base == 1 {
		return RightToLeft
	}
	return LeftToRight
}

// Analyze resolves the first paragraph using the requested base direction.
// Neutral selects the direction from the first strong character outside isolates,
// defaulting to LeftToRight. The other directions set the paragraph level explicitly.
// It includes the first paragraph separator and ignores subsequent text.
// Len reports the number of consumed runes; nonempty input always makes progress.
// The result does not retain text and is safe for concurrent reads.
func Analyze(text []rune, direction Direction) (Analysis, error) {
	level := implicitLevel
	switch direction {
	case Neutral:
	case LeftToRight:
		level = 0
	case RightToLeft:
		level = 1
	default:
		return Analysis{}, fmt.Errorf("bidi: invalid direction %d", direction)
	}
	p := Paragraph{text: text[:firstParagraphLen(text)], embeddingLevel: level}
	p.prepareInput()
	if len(p.initialTypes) != 0 {
		p.run()
	} else if p.embeddingLevel == implicitLevel {
		p.embeddingLevel = 0
	}
	return Analysis{types: p.initialTypes, levels: p.resultLevels, base: p.embeddingLevel}, nil
}

// LogicalRuns returns detached ranges in logical order with constant embedding
// levels, before line-specific L1 resets. Coordinates are rune indices in the
// analyzed text. The zero Analysis describes an empty left-to-right paragraph.
func (a Analysis) LogicalRuns() []Run {
	if len(a.levels) == 0 {
		return nil
	}
	p := Paragraph{resultLevels: a.levels}
	runs := p.buildRuns()
	out := make([]Run, runs.NumRuns())
	for i := range out {
		out[i] = runs.Run(i)
	}
	return out
}

// Line applies UAX #9 L1 and L2 to the rune range [start, end) and returns
// detached runs in visual order. Run coordinates remain logical indices in
// the analyzed text. The caller determines the actual line boundaries.
// Empty ranges return no runs; reversed or out-of-bounds ranges return an error.
// Shaping, combining marks (L3), and mirroring (L4) remain the caller's concern.
func (a Analysis) Line(start, end int) ([]Run, error) {
	if start < 0 || end < start || end > len(a.levels) {
		return nil, fmt.Errorf("bidi: invalid line range [%d,%d) for %d runes", start, end, len(a.levels))
	}
	if start == end {
		return nil, nil
	}
	p := Paragraph{
		initialTypes:   a.types[start:end],
		resultLevels:   append([]Level(nil), a.levels[start:end]...),
		embeddingLevel: a.base,
	}
	p.computeLevels()
	segmented := p.buildRuns()
	runs := make([]Run, segmented.NumRuns())
	var highest Level
	lowestOdd := Level(127)
	for i := range runs {
		run := segmented.Run(i)
		run.Start += start
		run.End += start
		runs[i] = run
		highest = max(highest, run.Level)
		if run.Level%2 == 1 && run.Level < lowestOdd {
			lowestOdd = run.Level
		}
	}
	// L2 reverses each contiguous sequence at or above the current level,
	// from the highest level down to the lowest odd level.
	for level := highest; level >= lowestOdd; level-- {
		for i := 0; i < len(runs); {
			if runs[i].Level < level {
				i++
				continue
			}
			first := i
			for i < len(runs) && runs[i].Level >= level {
				i++
			}
			for left, right := first, i-1; left < right; left, right = left+1, right-1 {
				runs[left], runs[right] = runs[right], runs[left]
			}
		}
	}
	return runs, nil
}

// Bound scratch storage to the consumed paragraph before prepareInput allocates.
func firstParagraphLen(text []rune) int {
	for i, r := range text {
		class, _ := ucd.LookupBidiClass(r)
		if class == ucd.BD_B {
			return i + 1
		}
	}
	return len(text)
}
