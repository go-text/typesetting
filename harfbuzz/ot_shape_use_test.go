package harfbuzz

import "testing"

func TestUSE(t *testing.T) {
	if !(joiningFormInit < 4 && joiningFormIsol < 4 && joiningFormMedi < 4 && joiningFormFina < 4) {
		t.Error()
	}
}
