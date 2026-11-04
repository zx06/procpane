package model

import "testing"

func TestUnknownConstant(t *testing.T) {
	if Unknown != "?" {
		t.Fatalf("Unknown = %q, want ?", Unknown)
	}
}
