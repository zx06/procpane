package kill

import (
	"os"
	"testing"
)

func TestValidateTargetPID_rejectsLowPIDs(t *testing.T) {
	for _, pid := range []int32{0, 1, -1} {
		if err := ValidateTargetPID(pid); err == nil {
			t.Fatalf("pid %d should be rejected", pid)
		}
	}
}

func TestValidateTargetPID_rejectsSelf(t *testing.T) {
	self := int32(os.Getpid())
	if err := ValidateTargetPID(self); err == nil {
		t.Fatal("expected refusal to kill self")
	}
}

func TestValidateTargetPID_allowsOther(t *testing.T) {
	// Pick a pid unlikely to be us or reserved; validation only checks rules.
	if err := ValidateTargetPID(999999); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}
