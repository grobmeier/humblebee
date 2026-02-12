package duration

import (
	"fmt"
	"time"
)

func FormatSeconds(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	d := time.Duration(seconds) * time.Second
	if d < time.Minute {
		return fmt.Sprintf("%ds", int64(d.Seconds()))
	}
	if d < time.Hour {
		m := int64(d.Minutes())
		s := int64(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", m, s)
	}
	h := int64(d.Hours())
	m := int64(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}

