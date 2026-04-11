package service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/timeutil"
)

type TimerService struct {
	entries *repo.TimeEntryRepo
}

func NewTimerService(db *sql.DB) *TimerService {
	return &TimerService{entries: repo.NewTimeEntryRepo(db)}
}

func (s *TimerService) Running(personID int64) (*model.TimeEntry, error) {
	return s.entries.FindRunning(personID)
}

type StartParams struct {
	PersonID   int64
	WorkItemID *int64 // nil means Default (stored as NULL)
	Now        time.Time
}

func (s *TimerService) Start(params StartParams) (int64, error) {
	_, offsetSec := params.Now.Zone()
	offsetMin := offsetSec / 60
	e := model.TimeEntry{
		UUID:      uuid.NewString(),
		PersonID:  params.PersonID,
		WorkItemID: params.WorkItemID,
		StartTime: params.Now.UTC().Unix(),
		TZName:     params.Now.Location().String(),
		TZOffsetMin: offsetMin,
		CreatedAt: params.Now.UTC().Unix(),
	}
	return s.entries.Start(e)
}

type StopResult struct {
	StoppedEntry model.TimeEntry
	DurationSec  int64
	TodayTotal   int64
}

func (s *TimerService) Stop(personID int64, now time.Time, loc *time.Location) (*StopResult, error) {
	running, err := s.entries.FindRunning(personID)
	if err != nil {
		return nil, err
	}
	if running == nil {
		return nil, errors.New("no timer is currently running")
	}

	end := now.UTC().Unix()
	if end <= running.StartTime {
		return nil, errors.New("invalid timer end time")
	}
	duration := end - running.StartTime
	if err := s.entries.Stop(running.ID, end, duration); err != nil {
		return nil, err
	}

	// Re-load-ish: just fill stopped fields.
	stopped := *running
	stoppedEnd := end
	stoppedDur := duration
	stopped.EndTime = &stoppedEnd
	stopped.Duration = &stoppedDur

	// Today total: sum overlap of all entries with today's local window.
	w := timeutil.TodayWindow(now, loc)
	overlapping, err := s.entries.ListOverlapping(personID, w.Start.UTC().Unix(), w.End.UTC().Unix())
	if err != nil {
		return nil, err
	}
	var total int64
	for _, e := range overlapping {
		if e.EndTime == nil {
			continue
		}
		total += timeutil.OverlapSeconds(e.StartTime, *e.EndTime, w)
	}

	return &StopResult{
		StoppedEntry: stopped,
		DurationSec:  duration,
		TodayTotal:   total,
	}, nil
}
