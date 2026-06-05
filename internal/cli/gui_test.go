package cli

import "testing"

func TestGUILaunchCandidatesPreferConfiguredPath(t *testing.T) {
	candidates := guiLaunchCandidates("/usr/local/bin/humblebee", "/Applications/HumbleBee.app", "darwin")

	if len(candidates) == 0 {
		t.Fatal("expected launch candidates")
	}
	if candidates[0].command != "open" || len(candidates[0].args) != 1 || candidates[0].args[0] != "/Applications/HumbleBee.app" {
		t.Fatalf("expected configured macOS app to be opened first, got %#v", candidates[0])
	}
}

func TestGUILaunchCandidatesIncludeAdjacentBinary(t *testing.T) {
	candidates := guiLaunchCandidates("/usr/local/bin/humblebee", "", "linux")

	if len(candidates) == 0 {
		t.Fatal("expected launch candidates")
	}
	if candidates[0].command != "/usr/local/bin/humblebee-gui" {
		t.Fatalf("expected adjacent GUI binary first, got %#v", candidates[0])
	}
}
