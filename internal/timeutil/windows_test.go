package timeutil

import (
	"testing"
	"time"
)

func TestOverlapSeconds(t *testing.T) {
	loc := time.FixedZone("X", 0)
	w := Window{
		Start: time.Date(2026, 2, 1, 0, 0, 0, 0, loc),
		End:   time.Date(2026, 2, 2, 0, 0, 0, 0, loc),
	}
	start := w.Start.UTC().Unix() - 10
	end := w.Start.UTC().Unix() + 10
	if got := OverlapSeconds(start, end, w); got != 10 {
		t.Fatalf("got %d", got)
	}
}

func TestSplitByLocalDay(t *testing.T) {
	loc := time.FixedZone("Local", 2*3600)
	startLocal := time.Date(2026, 2, 10, 23, 30, 0, 0, loc)
	endLocal := time.Date(2026, 2, 11, 0, 30, 0, 0, loc)
	m := SplitByLocalDay(startLocal.UTC().Unix(), endLocal.UTC().Unix(), loc)
	if m["2026-02-10"] == 0 || m["2026-02-11"] == 0 {
		t.Fatalf("expected split across two days, got %#v", m)
	}
	if m["2026-02-10"] != 30*60 || m["2026-02-11"] != 30*60 {
		t.Fatalf("unexpected split durations: %#v", m)
	}
}

func TestLocationForEntry(t *testing.T) {
	loc := LocationForEntry("America/New_York", 0, time.UTC)
	if loc == nil {
		t.Fatalf("expected location")
	}
	// If the location database is present, this should load and be stable.
	if loc.String() != "America/New_York" {
		t.Fatalf("got %q", loc.String())
	}

	fixed := LocationForEntry("", -300, time.UTC)
	if fixed == nil {
		t.Fatalf("expected fixed location")
	}
}
