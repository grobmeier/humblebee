package timeutil

import "time"

type Window struct {
	Start time.Time
	End   time.Time
}

func TodayWindow(now time.Time, loc *time.Location) Window {
	local := now.In(loc)
	startLocal := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	return Window{
		Start: startLocal,
		End:   startLocal.AddDate(0, 0, 1),
	}
}

func MonthWindow(year int, month time.Month, loc *time.Location) Window {
	startLocal := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	return Window{
		Start: startLocal,
		End:   startLocal.AddDate(0, 1, 0),
	}
}

func OverlapSeconds(startUTC, endUTC int64, window Window) int64 {
	ws := window.Start.UTC().Unix()
	we := window.End.UTC().Unix()

	if endUTC <= ws || startUTC >= we {
		return 0
	}
	start := startUTC
	if start < ws {
		start = ws
	}
	end := endUTC
	if end > we {
		end = we
	}
	if end <= start {
		return 0
	}
	return end - start
}

// SplitByLocalDay returns overlap duration per local-day (YYYY-MM-DD) for the given entry.
func SplitByLocalDay(entryStartUTC, entryEndUTC int64, loc *time.Location) map[string]int64 {
	out := map[string]int64{}
	if entryEndUTC <= entryStartUTC {
		return out
	}
	start := time.Unix(entryStartUTC, 0).In(loc)
	end := time.Unix(entryEndUTC, 0).In(loc)

	// Iterate local midnights.
	curDayStart := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	for curDayStart.Before(end) {
		next := curDayStart.AddDate(0, 0, 1)
		w := Window{Start: curDayStart, End: next}
		secs := OverlapSeconds(entryStartUTC, entryEndUTC, w)
		if secs > 0 {
			out[curDayStart.Format("2006-01-02")] = secs
		}
		curDayStart = next
	}
	return out
}

func LocationForEntry(tzName string, tzOffsetMin int, fallback *time.Location) *time.Location {
	if tzName != "" && tzName != "Local" {
		if loc, err := time.LoadLocation(tzName); err == nil && loc != nil {
			return loc
		}
	}
	if tzOffsetMin != 0 {
		return time.FixedZone("entry", tzOffsetMin*60)
	}
	if fallback != nil {
		return fallback
	}
	return time.Local
}
