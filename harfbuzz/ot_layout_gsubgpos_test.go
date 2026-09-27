package harfbuzz

import "testing"

func TestRecurseStopsAtNestingLimit(t *testing.T) {
	c := otApplyContext{buffer: NewBuffer(), nestingLevelLeft: maxNestingLevel}
	calls := 0
	c.recurseFunc = func(c *otApplyContext, _ uint16) bool {
		calls++
		if calls > maxNestingLevel {
			t.Fatal("lookup recursed past the nesting limit")
		}
		return c.recurse(0)
	}
	if c.recurse(0) || calls != maxNestingLevel || c.nestingLevelLeft != maxNestingLevel {
		t.Fatalf("recursion did not stop and restore its depth: calls=%d, depth=%d", calls, c.nestingLevelLeft)
	}
}

func TestRecurseWithoutCallback(t *testing.T) {
	c := otApplyContext{buffer: NewBuffer(), nestingLevelLeft: maxNestingLevel}
	if c.recurse(0) {
		t.Fatal("lookup recursed without a callback")
	}
}
