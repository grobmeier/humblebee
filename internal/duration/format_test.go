package duration

import "testing"

func TestFormatSeconds(t *testing.T) {
	if got := FormatSeconds(45); got != "45s" {
		t.Fatalf("got %q", got)
	}
	if got := FormatSeconds(75); got != "1m 15s" {
		t.Fatalf("got %q", got)
	}
	if got := FormatSeconds(2*3600 + 15*60 + 30); got != "2h 15m" {
		t.Fatalf("got %q", got)
	}
}

