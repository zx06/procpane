package autostart

import (
	"strings"
	"testing"
)

func TestLaunchAgentPlistXML(t *testing.T) {
	const exe = "/Applications/procpane.app/Contents/MacOS/procpane-app"
	got := launchAgentPlistXML(macBundleID, exe)
	for _, want := range []string{
		"<string>" + macBundleID + "</string>",
		"<string>" + exe + "</string>",
		"<key>RunAtLoad</key>",
		"<key>LimitLoadToSessionType</key>",
		"<string>Aqua</string>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("plist missing %q:\n%s", want, got)
		}
	}
}
