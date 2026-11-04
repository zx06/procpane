package ssh

import (
	"os"
	"strings"
	"testing"
)

func TestReadMergedConfigUsesHomeConfig(t *testing.T) {
	installUserSSHConfig(t, testdataPath(t, "cocotyty-sshconf", "basic", "simple.conf"))
	merged, err := readMergedConfig()
	if err != nil {
		t.Fatal(err)
	}
	aliases := parseHostAliases(merged)
	if len(aliases) < 3 {
		t.Fatalf("expected simple.conf hosts, got %v", aliases)
	}
}

func TestParseHostAliasesSkipsGlobAndNegation(t *testing.T) {
	data, err := os.ReadFile(testdataPath(t, "cocotyty-sshconf", "basic", "wildcards.conf"))
	if err != nil {
		t.Fatal(err)
	}
	got := parseHostAliases(string(data))
	for _, a := range got {
		if a == "*" || strings.HasPrefix(a, "!") {
			t.Fatalf("parseHostAliases should skip %q", a)
		}
	}
	if _, ok := sliceContains(got, "prod.server.example.com"); !ok {
		t.Fatalf("missing prod host in %v", got)
	}
}

func sliceContains(ss []string, v string) (string, bool) {
	for _, s := range ss {
		if s == v {
			return s, true
		}
	}
	return "", false
}
