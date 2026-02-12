package service

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/repo"
)

type WorkItemService struct {
	items *repo.WorkItemRepo
}

func NewWorkItemService(db *sql.DB) *WorkItemService {
	return &WorkItemService{items: repo.NewWorkItemRepo(db)}
}

func ParseWorkItemPath(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errors.New("work item name is required")
	}
	parts := strings.Split(input, " > ")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, errors.New("invalid work item path")
		}
		if len(p) < 1 || len(p) > 200 {
			return nil, errors.New("work item name must be 1-200 characters")
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *WorkItemService) ResolveByPath(personID int64, segments []string) (*model.WorkItem, error) {
	var parentID *int64
	var current *model.WorkItem
	for _, seg := range segments {
		item, err := s.items.FindByNameUnderParent(personID, parentID, seg)
		if err != nil {
			return nil, err
		}
		if item == nil {
			return nil, fmt.Errorf("work item '%s' not found", seg)
		}
		current = item
		parentID = &item.ID
	}
	return current, nil
}

type ResolveResult struct {
	Item       *model.WorkItem
	Candidates []model.WorkItem
}

func (s *WorkItemService) ResolveByInput(personID int64, input string) (ResolveResult, error) {
	segments, err := ParseWorkItemPath(input)
	if err != nil {
		return ResolveResult{}, err
	}
	if len(segments) > 1 {
		item, err := s.ResolveByPath(personID, segments)
		if err != nil {
			return ResolveResult{}, err
		}
		return ResolveResult{Item: item}, nil
	}
	name := segments[0]
	matches, err := s.items.FindByNameAnyLevel(personID, name)
	if err != nil {
		return ResolveResult{}, err
	}
	if len(matches) == 0 {
		return ResolveResult{}, fmt.Errorf("work item '%s' not found", name)
	}
	if len(matches) == 1 {
		m := matches[0]
		return ResolveResult{Item: &m}, nil
	}
	return ResolveResult{Candidates: matches}, fmt.Errorf("work item '%s' is ambiguous; use full path with ' > '", name)
}

func (s *WorkItemService) CreateFromInput(personID int64, input string, now time.Time) (*model.WorkItem, *model.WorkItem, error) {
	segments, err := ParseWorkItemPath(input)
	if err != nil {
		return nil, nil, err
	}
	var parentID *int64
	var parent *model.WorkItem
	depth := 0

	// Resolve parent chain (all but last segment must exist).
	if len(segments) > 1 {
		for i := 0; i < len(segments)-1; i++ {
			seg := segments[i]
			item, err := s.items.FindByNameUnderParent(personID, parentID, seg)
			if err != nil {
				return nil, nil, err
			}
			if item == nil {
				return nil, nil, fmt.Errorf("parent work item '%s' not found", seg)
			}
			parent = item
			parentID = &item.ID
			depth = item.Depth + 1
		}
	}

	name := segments[len(segments)-1]
	if strings.EqualFold(name, "Default") && parentID == nil {
		return nil, nil, errors.New("work item 'Default' already exists")
	}
	existing, err := s.items.FindByNameUnderParent(personID, parentID, name)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return nil, parent, fmt.Errorf("work item '%s' already exists", name)
	}

	created, err := s.items.Create(repo.CreateWorkItemParams{
		PersonID: personID,
		UUID:     uuid.NewString(),
		Name:     name,
		ParentID: parentID,
		Depth:    depth,
		Created:  now.UTC().Unix(),
	})
	if err != nil {
		return nil, nil, err
	}
	return created, parent, nil
}

type TreeNode struct {
	Item     model.WorkItem
	Children []*TreeNode
}

func BuildTree(items []model.WorkItem) []*TreeNode {
	nodes := make(map[int64]*TreeNode, len(items))
	for _, it := range items {
		copy := it
		nodes[it.ID] = &TreeNode{Item: copy}
	}
	var roots []*TreeNode
	for _, n := range nodes {
		if n.Item.ParentID == nil {
			roots = append(roots, n)
			continue
		}
		parent := nodes[*n.Item.ParentID]
		if parent == nil {
			roots = append(roots, n)
			continue
		}
		parent.Children = append(parent.Children, n)
	}

	var sortNode func(n *TreeNode)
	sortNode = func(n *TreeNode) {
		sort.Slice(n.Children, func(i, j int) bool {
			return strings.ToLower(n.Children[i].Item.Name) < strings.ToLower(n.Children[j].Item.Name)
		})
		for _, c := range n.Children {
			sortNode(c)
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		return strings.ToLower(roots[i].Item.Name) < strings.ToLower(roots[j].Item.Name)
	})
	for _, r := range roots {
		sortNode(r)
	}
	return roots
}

