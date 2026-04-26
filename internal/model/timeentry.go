package model

type TimeEntry struct {
	ID          int64
	UUID        string
	PersonID    int64
	WorkItemID  *int64
	Description *string
	StartTime   int64
	EndTime     *int64
	Duration    *int64
	TZName      string
	TZOffsetMin int
	CreatedAt   int64
	UpdatedAt   *int64
}
