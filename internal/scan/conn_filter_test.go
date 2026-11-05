package scan

import (
	"context"
	"testing"
)

func TestListRawListeners_unchangedSkipsEstablished(t *testing.T) {
	// Regression guard: LISTEN path must not start returning ESTABLISHED rows.
	// We only assert the helper boundary here; integration uses the live OS table.
	if isEstablished("LISTEN") || isEstablished("NONE") {
		t.Fatal("isListen statuses must not match isEstablished")
	}
	if !isListen("LISTEN") || !isListen("NONE") {
		t.Fatal("LISTEN/NONE must remain listen states")
	}
	_, err := ListRawListeners(context.Background(), "tcp")
	if err != nil {
		t.Fatalf("ListRawListeners: %v", err)
	}
}
