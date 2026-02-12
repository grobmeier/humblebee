package model

type WorkItemStatus string

const (
	WorkItemStatusActive   WorkItemStatus = "ACTIVE"
	WorkItemStatusArchived WorkItemStatus = "ARCHIVED"
	WorkItemStatusDeleted  WorkItemStatus = "DELETED"
)

type WorkItem struct {
	ID          int64
	UUID        string
	PersonID    int64
	Name        string
	Description *string
	ParentID    *int64
	Path        *string
	Depth       int
	Status      WorkItemStatus
	Color       *string
	CreatedAt   int64
	UpdatedAt   *int64
}

