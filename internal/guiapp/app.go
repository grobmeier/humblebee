package guiapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/paths"
	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/service"
	"github.com/grobmeier/humblebee/internal/timeutil"
)

type App struct {
	ctx context.Context
}

func New() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

type Dashboard struct {
	Initialized bool          `json:"initialized"`
	DBPath      string        `json:"dbPath"`
	UserEmail   string        `json:"userEmail"`
	Running     *RunningTimer `json:"running"`
	TodayTotal  int64         `json:"todayTotalSeconds"`
}

type RunningTimer struct {
	WorkItemName string `json:"workItemName"`
	StartTimeUTC int64  `json:"startTimeUTC"`
}

type WorkItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ParentID *int64 `json:"parentId"`
	Depth    int    `json:"depth"`
}

type StopResult struct {
	WorkItemName     string `json:"workItemName"`
	DurationSeconds  int64  `json:"durationSeconds"`
	TodayTotalSeconds int64 `json:"todayTotalSeconds"`
}

func (a *App) GetDashboard() (*Dashboard, error) {
	database, dbPath, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()

	initialized, err := db.IsInitialized(database)
	if err != nil {
		return nil, err
	}
	out := &Dashboard{
		Initialized: initialized,
		DBPath:      dbPath,
	}
	if !initialized {
		return out, nil
	}

	personRepo := repo.NewPersonRepo(database)
	p, err := personRepo.GetDefault()
	if err != nil {
		return nil, err
	}
	if p != nil {
		out.UserEmail = p.Email
	}
	if p == nil {
		return out, nil
	}

	timerRepo := repo.NewTimeEntryRepo(database)
	running, err := timerRepo.FindRunning(p.ID)
	if err != nil {
		return nil, err
	}
	if running != nil {
		name := "Default"
		if running.WorkItemID != nil {
			itemsRepo := repo.NewWorkItemRepo(database)
			item, _ := itemsRepo.GetByID(p.ID, *running.WorkItemID)
			if item != nil {
				name = item.Name
			}
		}
		out.Running = &RunningTimer{
			WorkItemName: name,
			StartTimeUTC: running.StartTime,
		}
	}

	// Today total (current local day window).
	w := timeutil.TodayWindow(time.Now(), time.Local)
	entries, err := timerRepo.ListOverlapping(p.ID, w.Start.UTC().Unix(), w.End.UTC().Unix())
	if err != nil {
		return nil, err
	}
	var total int64
	for _, e := range entries {
		if e.EndTime == nil {
			continue
		}
		total += timeutil.OverlapSeconds(e.StartTime, *e.EndTime, w)
	}
	out.TodayTotal = total
	return out, nil
}

func (a *App) Init(email string) error {
	email = strings.TrimSpace(email)
	database, _, err := a.openDB()
	if err != nil {
		return err
	}
	defer database.Close()

	initialized, err := db.IsInitialized(database)
	if err != nil {
		return err
	}
	if initialized {
		return errors.New("already initialized")
	}
	if err := db.Migrate(database); err != nil {
		return err
	}
	initSvc := service.NewInitService(database)
	_, _, err = initSvc.Init(service.InitParams{
		Email:           email,
		InitialWorkItem: "",
		Now:             time.Now(),
	})
	return err
}

func (a *App) ListWorkItems() ([]WorkItem, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}
	itemsRepo := repo.NewWorkItemRepo(database)
	items, err := itemsRepo.ListActive(personID)
	if err != nil {
		return nil, err
	}
	out := make([]WorkItem, 0, len(items))
	for _, it := range items {
		// Expose Default row too (GUI can choose to hide it and use Default semantics).
		out = append(out, WorkItem{
			ID:       it.ID,
			Name:     it.Name,
			ParentID: it.ParentID,
			Depth:    it.Depth,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func (a *App) Start(workItemID int64) error {
	database, _, err := a.openDB()
	if err != nil {
		return err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return err
	}

	timer := service.NewTimerService(database)
	running, err := timer.Running(personID)
	if err != nil {
		return err
	}
	if running != nil {
		return errors.New("timer already running")
	}

	var idPtr *int64
	// GUI uses 0 to mean Default (NULL workitem_id).
	if workItemID != 0 {
		idPtr = &workItemID
	}
	_, err = timer.Start(service.StartParams{
		PersonID:   personID,
		WorkItemID: idPtr,
		Now:        time.Now(),
	})
	return err
}

func (a *App) Stop() (*StopResult, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}

	timer := service.NewTimerService(database)
	res, err := timer.Stop(personID, time.Now(), time.Local)
	if err != nil {
		return nil, err
	}

	name := "Default"
	if res.StoppedEntry.WorkItemID != nil {
		itemsRepo := repo.NewWorkItemRepo(database)
		item, _ := itemsRepo.GetByID(personID, *res.StoppedEntry.WorkItemID)
		if item != nil {
			name = item.Name
		}
	}
	return &StopResult{
		WorkItemName:      name,
		DurationSeconds:   res.DurationSec,
		TodayTotalSeconds: res.TodayTotal,
	}, nil
}

func (a *App) openDB() (*sql.DB, string, error) {
	dbPath, err := paths.DBPath()
	if err != nil {
		return nil, "", err
	}
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, "", err
	}
	initialized, err := db.IsInitialized(database)
	if err != nil {
		_ = database.Close()
		return nil, "", err
	}
	if initialized {
		if err := db.Migrate(database); err != nil {
			_ = database.Close()
			return nil, "", err
		}
	}
	return database, dbPath, nil
}

func (a *App) requireInitialized(database *sql.DB) error {
	ok, err := db.IsInitialized(database)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("not initialized")
	}
	return nil
}

func (a *App) defaultPersonID(database *sql.DB) (int64, error) {
	people := repo.NewPersonRepo(database)
	p, err := people.GetDefault()
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, fmt.Errorf("no default user found; run init")
	}
	return p.ID, nil
}

