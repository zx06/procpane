package scan

import (
	"context"
	"testing"
)

func TestListRawListeners_all(t *testing.T) {
	raw, err := ListRawListeners(context.Background(), "all")
	if err != nil {
		t.Fatalf("ListRawListeners: %v", err)
	}
	t.Logf("found %d listeners", len(raw))
}
