package openapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

type ProjectService struct{ client *Client }

type ProjectQuery struct {
	ProjectID     *int64
	ProjectName   *string
	ClubID        *int64
	Status        *datagsm.ProjectStatus
	Category      *datagsm.ProjectCategory
	Page          int
	Size          int
	SortBy        datagsm.ProjectSortBy
	SortDirection datagsm.SortDirection
}

type ProjectPage struct {
	Projects      []datagsm.Project `json:"projects"`
	TotalElements int64             `json:"totalElements"`
	TotalPages    int               `json:"totalPages"`
}

type ProjectRequest struct {
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	StartYear      int                   `json:"startYear"`
	ClubID         *int64                `json:"clubId,omitempty"`
	ParticipantIDs []int64               `json:"participantIds"`
	Status         datagsm.ProjectStatus `json:"status,omitempty"`
	EndYear        *int                  `json:"endYear,omitempty"`
	// Nil preserves categories on update; an empty slice clears them.
	Categories    []datagsm.ProjectCategory `json:"categories,omitzero"`
	Repositories  []string                  `json:"repositories,omitempty"`
	TechStacks    []string                  `json:"techStacks,omitempty"`
	IconKey       *string                   `json:"iconKey,omitempty"`
	DeploymentURL *string                   `json:"deploymentUrl,omitempty"`
}

func (q ProjectQuery) params() (url.Values, error) {
	p, err := pagination(q.Page, q.Size, string(q.SortBy), q.SortDirection)
	if err != nil {
		return nil, err
	}
	set(p, "projectId", q.ProjectID)
	set(p, "projectName", q.ProjectName)
	set(p, "clubId", q.ClubID)
	set(p, "status", q.Status)
	set(p, "category", q.Category)
	return p, nil
}

func (s *ProjectService) List(ctx context.Context, query ProjectQuery) (*ProjectPage, error) {
	params, err := query.params()
	if err != nil {
		return nil, err
	}
	var result ProjectPage
	if err := s.client.do(ctx, http.MethodGet, "/v1/projects", params, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ProjectService) Create(ctx context.Context, input ProjectRequest) (*datagsm.Project, error) {
	return s.write(ctx, http.MethodPost, "/v1/projects", input)
}

// Update clears omitted repositories, tech stacks, and club; status changes use End or Reactivate.
func (s *ProjectService) Update(ctx context.Context, id int64, input ProjectRequest) (*datagsm.Project, error) {
	if err := validID(id); err != nil {
		return nil, err
	}
	if input.Status != "" || input.EndYear != nil {
		return nil, errors.New("datagsm: use End or Reactivate to change project status")
	}
	return s.write(ctx, http.MethodPut, fmt.Sprintf("/v1/projects/%d", id), input)
}

func (s *ProjectService) write(ctx context.Context, method, path string, input ProjectRequest) (*datagsm.Project, error) {
	if input.ParticipantIDs == nil {
		input.ParticipantIDs = []int64{}
	}
	var result datagsm.Project
	if err := s.client.do(ctx, method, path, nil, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ProjectService) Delete(ctx context.Context, id int64) error {
	if err := validID(id); err != nil {
		return err
	}
	return s.client.do(ctx, http.MethodDelete, fmt.Sprintf("/v1/projects/%d", id), nil, nil, nil)
}

func (s *ProjectService) End(ctx context.Context, id int64, year int) error {
	if err := validID(id); err != nil {
		return err
	}
	if year <= 0 {
		return errors.New("datagsm: end year must be positive")
	}
	return s.client.do(ctx, http.MethodPost, fmt.Sprintf("/v1/projects/%d/end", id), nil, struct {
		EndYear int `json:"endYear"`
	}{year}, nil)
}

func (s *ProjectService) Reactivate(ctx context.Context, id int64) error {
	if err := validID(id); err != nil {
		return err
	}
	return s.client.do(ctx, http.MethodPost, fmt.Sprintf("/v1/projects/%d/reactivate", id), nil, nil, nil)
}
