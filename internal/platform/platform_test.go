package platform

import "testing"

func TestExtractMAC(t *testing.T) {
	cases := map[string]string{
		"aa:bb:cc:dd:ee:ff Keyboard":   "aa:bb:cc:dd:ee:ff",
		"AA-BB-CC-DD-EE-FF, Mouse":     "AA:BB:CC:DD:EE:FF",
		"Device B0:1F:2C:3D:4E:5F Foo": "B0:1F:2C:3D:4E:5F",
		"no address here":              "",
		"short aa:bb:cc":               "",
	}
	for in, want := range cases {
		if got := ExtractMAC(in); got != want {
			t.Errorf("ExtractMAC(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExpand(t *testing.T) {
	got := expand("amixer set Master {0}% and {1}", []string{"80", "x"})
	want := "amixer set Master 80% and x"
	if got != want {
		t.Fatalf("expand = %q, want %q", got, want)
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"plain":       "'plain'",
		"with space":  "'with space'",
		"it's":        `'it'\''s'`,
		`$(rm -rf /)`: `'$(rm -rf /)'`,
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestOverrideCommand(t *testing.T) {
	SetOverrides(map[string]string{"volume_up": "custom {0}", "  ": "ignored"})
	defer SetOverrides(nil)

	cmd, ok := overrideCommand(VolumeUp, []string{"5"})
	if !ok || cmd != "custom 5" {
		t.Fatalf("overrideCommand = %q, %v", cmd, ok)
	}
	if _, ok := overrideCommand(VolumeDown, nil); ok {
		t.Fatal("unexpected override for volume_down")
	}
}
