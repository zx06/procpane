package kill

import (
	"fmt"
	"os"
)

// ValidateTargetPID rejects dangerous kill targets.
func ValidateTargetPID(pid int32) error {
	if pid <= 1 {
		return fmt.Errorf("refusing to kill pid %d (reserved system pid)", pid)
	}
	self := int32(os.Getpid())
	if pid == self {
		return fmt.Errorf("refusing to kill own process (pid %d)", pid)
	}
	return nil
}
