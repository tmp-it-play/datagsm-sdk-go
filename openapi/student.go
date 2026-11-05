package openapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

type StudentService struct{ client *Client }

type StudentQuery struct {
	StudentID        *int64
	Name             *string
	Email            *string
	Grade            *int
	ClassNum         *int
	Number           *int
	Sex              *datagsm.Sex
	Role             *datagsm.StudentRole
	Specialty        *string
	Major            *datagsm.Major
	GitHubID         *string
	DormitoryRoom    *int
	IncludeGraduates *bool
	IncludeWithdrawn *bool
	OnlyEnrolled     *bool
	Page             int
	Size             int
	SortBy           datagsm.StudentSortBy
	SortDirection    datagsm.SortDirection
}

type StudentPage struct {
	Students      []datagsm.Student `json:"students"`
	TotalElements int64             `json:"totalElements"`
	TotalPages    int               `json:"totalPages"`
}

type CreateStudentRequest struct {
	Name                string              `json:"name"`
	Sex                 datagsm.Sex         `json:"sex"`
	Email               string              `json:"email"`
	Grade               *int                `json:"grade,omitempty"`
	ClassNum            *int                `json:"classNum,omitempty"`
	Number              *int                `json:"number,omitempty"`
	Role                datagsm.StudentRole `json:"role,omitempty"`
	DormitoryRoomNumber *int                `json:"dormitoryRoomNumber,omitempty"`
	MajorClubID         *int64              `json:"majorClubId,omitempty"`
	AutonomousClubID    *int64              `json:"autonomousClubId,omitempty"`
	Specialty           *string             `json:"specialty,omitempty"`
}

type UpdateStudentRequest struct {
	Name                string              `json:"name"`
	Sex                 datagsm.Sex         `json:"sex"`
	Email               string              `json:"email"`
	Grade               int                 `json:"grade"`
	ClassNum            int                 `json:"classNum"`
	Number              int                 `json:"number"`
	Role                datagsm.StudentRole `json:"role"`
	DormitoryRoomNumber *int                `json:"dormitoryRoomNumber,omitempty"`
	MajorClubID         *int64              `json:"majorClubId,omitempty"`
	AutonomousClubID    *int64              `json:"autonomousClubId,omitempty"`
	Specialty           *string             `json:"specialty,omitempty"`
	GitHubID            *string             `json:"githubId,omitempty"`
}

func (q StudentQuery) params() (url.Values, error) {
	p, err := pagination(q.Page, q.Size, string(q.SortBy), q.SortDirection)
	if err != nil {
		return nil, err
	}
	set(p, "studentId", q.StudentID)
	set(p, "name", q.Name)
	set(p, "email", q.Email)
	set(p, "grade", q.Grade)
	set(p, "classNum", q.ClassNum)
	set(p, "number", q.Number)
	set(p, "sex", q.Sex)
	set(p, "role", q.Role)
	set(p, "specialty", q.Specialty)
	set(p, "major", q.Major)
	set(p, "githubId", q.GitHubID)
	set(p, "dormitoryRoom", q.DormitoryRoom)
	set(p, "includeGraduates", q.IncludeGraduates)
	set(p, "includeWithdrawn", q.IncludeWithdrawn)
	set(p, "onlyEnrolled", q.OnlyEnrolled)
	return p, nil
}

func (s *StudentService) List(ctx context.Context, query StudentQuery) (*StudentPage, error) {
	params, err := query.params()
	if err != nil {
		return nil, err
	}
	var result StudentPage
	if err := s.client.do(ctx, http.MethodGet, "/v1/students", params, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *StudentService) Create(ctx context.Context, input CreateStudentRequest) (*datagsm.Student, error) {
	var result datagsm.Student
	if err := s.client.do(ctx, http.MethodPost, "/v1/students", nil, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces student information; omitted optional fields are cleared.
func (s *StudentService) Update(ctx context.Context, id int64, input UpdateStudentRequest) (*datagsm.Student, error) {
	if err := validID(id); err != nil {
		return nil, err
	}
	var result datagsm.Student
	if err := s.client.do(ctx, http.MethodPut, fmt.Sprintf("/v1/students/%d", id), nil, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
