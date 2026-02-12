package service

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/validator"
)

type InitService struct {
	people *repo.PersonRepo
	items  *repo.WorkItemRepo
}

func NewInitService(db *sql.DB) *InitService {
	return &InitService{
		people: repo.NewPersonRepo(db),
		items:  repo.NewWorkItemRepo(db),
	}
}

type InitParams struct {
	Email           string
	InitialWorkItem string
	Now             time.Time
}

func (s *InitService) Init(params InitParams) (*model.Person, []model.WorkItem, error) {
	email := strings.TrimSpace(params.Email)
	if err := validator.ValidateEmail(email); err != nil {
		return nil, nil, err
	}
	username := email
	if at := strings.Index(email, "@"); at > 0 {
		username = email[:at]
	}

	person := model.Person{
		UUID:      uuid.NewString(),
		Email:     email,
		Username:  username,
		CreatedAt: params.Now.UTC().Unix(),
		IsActive:  true,
		IsDefault: true,
	}
	personID, err := s.people.CreateDefault(person)
	if err != nil {
		return nil, nil, err
	}
	person.ID = personID

	var created []model.WorkItem
	defaultItem, err := s.items.Create(repoCreateWorkItem(personID, "Default", nil, 0, params.Now))
	if err != nil {
		return nil, nil, err
	}
	created = append(created, *defaultItem)

	if strings.TrimSpace(params.InitialWorkItem) != "" {
		name := strings.TrimSpace(params.InitialWorkItem)
		if len(name) < 1 || len(name) > 200 {
			return nil, nil, errors.New("work item name must be 1-200 characters")
		}
		item, err := s.items.Create(repoCreateWorkItem(personID, name, nil, 0, params.Now))
		if err != nil {
			return nil, nil, err
		}
		created = append(created, *item)
	}

	return &person, created, nil
}

func repoCreateWorkItem(personID int64, name string, parentID *int64, depth int, now time.Time) repo.CreateWorkItemParams {
	return repo.CreateWorkItemParams{
		PersonID: personID,
		UUID:     uuid.NewString(),
		Name:     name,
		ParentID: parentID,
		Depth:    depth,
		Created:  now.UTC().Unix(),
	}
}
