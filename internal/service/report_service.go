package service

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/timeutil"
)

type ReportService struct {
	items   *repo.WorkItemRepo
	entries *repo.TimeEntryRepo
}

func NewReportService(db *sql.DB) *ReportService {
	return &ReportService{
		items:   repo.NewWorkItemRepo(db),
		entries: repo.NewTimeEntryRepo(db),
	}
}

type ReportLine struct {
	WorkItemID int64
	Name       string
	Depth      int
	Seconds    int64
	Percent    *int
}

type MonthlyReport struct {
	Title       string
	Lines       []ReportLine
	TotalSec    int64
	WorkingDays int
	AvgPerDay   int64
}

func (s *ReportService) Monthly(personID int64, year int, month time.Month, now time.Time, loc *time.Location) (*MonthlyReport, error) {
	window := timeutil.MonthWindow(year, month, loc)
	entries, err := s.entries.ListOverlapping(personID, window.Start.UTC().Unix(), window.End.UTC().Unix())
	if err != nil {
		return nil, err
	}

	// Load work items (include archived so reports remain complete).
	workItems, err := s.items.ListAll(personID)
	if err != nil {
		return nil, err
	}

	// Map NULL workitem_id entries to the Default work item row (if present).
	var defaultID int64 = 0
	if def, err := s.items.FindByNameUnderParent(personID, nil, "Default"); err == nil && def != nil {
		defaultID = def.ID
	}

	own := map[int64]int64{}
	daySet := map[string]struct{}{}

	for _, e := range entries {
		if e.EndTime == nil {
			continue
		}
		secs := timeutil.OverlapSeconds(e.StartTime, *e.EndTime, window)
		if secs <= 0 {
			continue
		}
		id := defaultID
		if e.WorkItemID != nil {
			id = *e.WorkItemID
		}
		own[id] += secs

		perDay := timeutil.SplitByLocalDay(e.StartTime, *e.EndTime, loc)
		for day, dur := range perDay {
			if dur <= 0 {
				continue
			}
			// Only count days within the month window.
			t, err := time.ParseInLocation("2006-01-02", day, loc)
			if err != nil {
				continue
			}
			if !t.Before(window.Start) && t.Before(window.End) {
				daySet[day] = struct{}{}
			}
		}
	}

	// Build tree + compute subtree totals.
	nodes := map[int64]*TreeNode{}
	for _, wi := range workItems {
		copy := wi
		nodes[wi.ID] = &TreeNode{Item: copy}
	}
	for _, n := range nodes {
		if n.Item.ParentID == nil {
			continue
		}
		parent := nodes[*n.Item.ParentID]
		if parent != nil {
			parent.Children = append(parent.Children, n)
		}
	}
	// Ensure Default exists for report even if missing.
	if defaultID == 0 {
		nodes[0] = &TreeNode{Item: model.WorkItem{ID: 0, Name: "Default", Depth: 0, Status: model.WorkItemStatusActive}}
		defaultID = 0
	}

	var sortNode func(n *TreeNode)
	sortNode = func(n *TreeNode) {
		sort.Slice(n.Children, func(i, j int) bool {
			return n.Children[i].Item.Name < n.Children[j].Item.Name
		})
		for _, c := range n.Children {
			sortNode(c)
		}
	}

	roots := []*TreeNode{}
	for _, n := range nodes {
		if n.Item.ParentID == nil {
			roots = append(roots, n)
		}
	}
	// Ensure Default is in roots if it has time but isn't root (should be root).
	if def := nodes[defaultID]; def != nil && def.Item.ParentID == nil {
		// ok
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].Item.Name < roots[j].Item.Name })
	for _, r := range roots {
		sortNode(r)
	}

	subtree := map[int64]int64{}
	var compute func(n *TreeNode) int64
	compute = func(n *TreeNode) int64 {
		total := own[n.Item.ID]
		for _, c := range n.Children {
			total += compute(c)
		}
		subtree[n.Item.ID] = total
		return total
	}
	for _, r := range roots {
		compute(r)
	}

	var total int64
	for _, secs := range own {
		total += secs
	}

	lines := []ReportLine{}
	var addLines func(n *TreeNode, depth int, isRoot bool)
	addLines = func(n *TreeNode, depth int, isRoot bool) {
		secs := subtree[n.Item.ID]
		if secs <= 0 {
			// Skip items with no time in this window, but still traverse to find descendants with time.
			for _, c := range n.Children {
				addLines(c, depth+1, false)
			}
			return
		}
		var percent *int
		if isRoot && total > 0 {
			p := int((float64(secs) / float64(total) * 100.0) + 0.5)
			percent = &p
		}
		lines = append(lines, ReportLine{
			WorkItemID: n.Item.ID,
			Name:       n.Item.Name,
			Depth:      depth,
			Seconds:    secs,
			Percent:    percent,
		})
		for _, c := range n.Children {
			addLines(c, depth+1, false)
		}
	}
	for _, r := range roots {
		addLines(r, 0, true)
	}

	workingDays := len(daySet)
	var avg int64
	if workingDays > 0 {
		avg = total / int64(workingDays)
	}

	title := fmt.Sprintf("Time Report - %s %d", month.String(), year)
	_ = now // reserved for future options
	return &MonthlyReport{
		Title:       title,
		Lines:       lines,
		TotalSec:    total,
		WorkingDays: workingDays,
		AvgPerDay:   avg,
	}, nil
}

