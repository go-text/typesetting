package harfbuzz

import (
	"fmt"
	"strings"
)

// Ported from src/hb-buffer-verify.cc

func (b *Buffer) verifyValidGID(font *Font) error {
	for _, glyph := range b.Info {
		_, ok := font.GlyphExtents(glyph.Glyph)
		if !ok {
			return fmt.Errorf("Unknow glyph %d in font", glyph.Glyph)
		}
	}
	return nil
}

// check that clusters are monotone.
func (b *Buffer) verifyMonotone() error {
	if b.ClusterLevel == MonotoneGraphemes || b.ClusterLevel == MonotoneCharacters {
		isForward := b.Props.Direction.isForward()

		info := b.Info

		for i := 1; i < len(info); i++ {
			if info[i-1].Cluster != info[i].Cluster && (info[i-1].Cluster < info[i].Cluster) != isForward {
				return fmt.Errorf("cluster at index %d is not monotone", i)
			}
		}
	}

	return nil
}

func (b *Buffer) showRunes() string {
	var s strings.Builder
	for _, r := range b.Info {
		fmt.Fprintf(&s, "U+%04X(at:%d),", r.codepoint, r.Cluster)
	}
	return s.String()
}

func (b *Buffer) showGIDs() string {
	var s strings.Builder
	for _, r := range b.Info {
		fmt.Fprintf(&s, "%d,", r.Glyph)
	}
	return s.String()
}

func (b *Buffer) verifyUnsafeToBreak(textBuffer *Buffer, font *Font, features []Feature) error {
	if b.ClusterLevel != MonotoneGraphemes && b.ClusterLevel != MonotoneCharacters {
		/* Cannot perform this check without monotone clusters. */
		return nil
	}

	/* Check that breaking up shaping at safe-to-break is indeed safe. */

	fragment, reconstruction := NewBuffer(), NewBuffer()
	copyBufferProperties(reconstruction, b)

	info := b.Info
	text := textBuffer.Info

	/* Chop text and shape fragments. */
	forward := b.Props.Direction.isForward()
	start := 0
	textStart := len(textBuffer.Info)
	if forward {
		textStart = 0
	}
	textEnd := textStart
	for end := 1; end < len(b.Info)+1; end++ {
		offset := 1
		if forward {
			offset = 0
		}
		if end < len(b.Info) && (info[end].Cluster == info[end-1].Cluster ||
			info[end-offset].Mask&GlyphUnsafeToBreak != 0) {
			continue
		}

		/* Shape segment corresponding to glyphs start..end. */
		if end == len(b.Info) {
			if forward {
				textEnd = len(textBuffer.Info)
			} else {
				textStart = 0
			}
		} else {
			if forward {
				cluster := info[end].Cluster
				for textEnd < len(textBuffer.Info) && text[textEnd].Cluster < cluster {
					textEnd++
				}
			} else {
				cluster := info[end-1].Cluster
				for textStart != 0 && text[textStart-1].Cluster >= cluster {
					textStart--
				}
			}
		}
		if !(textStart < textEnd) {
			return fmt.Errorf("unexpected %d >= %d", textStart, textEnd)
		}

		if debugMode {
			fmt.Println()
			fmt.Printf("VERIFY SAFE TO BREAK : start %d end %d text start %d end %d\n", start, end, textStart, textEnd)
			fmt.Println()
		}

		fragment.Clear()
		copyBufferProperties(fragment, b)

		flags := fragment.Flags
		if 0 < textStart {
			flags = (flags & ^Bot)
		}
		if textEnd < len(textBuffer.Info) {
			flags = (flags & ^Eot)
		}
		fragment.Flags = flags

		appendBuffer(fragment, textBuffer, textStart, textEnd)
		fragment.Shape(font, features)
		appendBuffer(reconstruction, fragment, 0, len(fragment.Info))

		start = end
		if forward {
			textStart = textEnd
		} else {
			textEnd = textStart
		}
	}

	diff := bufferDiff(reconstruction, b, ^GID(0), 0)
	if diff & ^bdfGlyphFlagsMismatch != 0 {
		return fmt.Errorf("unsafe-to-break test failed: %b (%s -> %s)", diff, b.showRunes(), b.showGIDs())
	}

	return nil
}
