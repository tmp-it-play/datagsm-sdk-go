package openapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

type ClubService struct{ client *Client }

type ClubQuery struct {
	ClubID                      *int64
	ClubName                    *string
	ClubType                    *datagsm.ClubType
	ClubStatus                  *datagsm.ClubStatus
	FoundedYear                 *int
	IncludeLeaderInParticipants *bool
	Page                        int
	Size                        int
	SortBy                      datagsm.ClubSortBy
	SortDirection               datagsm.SortDirection
}

type ClubPage struct {
	Clubs         []datagsm.Club `json:"clubs"`
	TotalElements int64          `json:"totalElements"`
	TotalPages    int            `json:"totalPages"`
}

type ClubRequest struct {
	Name           string             `json:"name"`
	Type           datagsm.ClubType   `json:"type"`
	LeaderID       *int64             `json:"leaderId,omitempty"`
	ParticipantIDs []int64            `json:"participantIds"`
	FoundedYear    int                `json:"foundedYear"`
	Status         datagsm.ClubStatus `json:"status"`
	AbolishedYear  *int               `json:"abolishedYear,omitempty"`
}

func (q ClubQuery) params() (url.Values, error) {
	p, err := pagination(q.Page, q.Size, string(q.SortBy), q.SortDirection)
	if err != nil {
		return nil, err
	}
	set(p, "clubId", q.ClubID)
	set(p, "clubName", q.ClubName)
	set(p, "clubType", q.ClubType)
	set(p, "clubStatus", q.ClubStatus)
	set(p, "foundedYear", q.FoundedYear)
	set(p, "includeLeaderInParticipants", q.IncludeLeaderInParticipants)
	return p, nil
}

func (s *ClubService) List(ctx context.Context, query ClubQuery) (*ClubPage, error) {
	params, err := query.params()
	if err != nil {
		return nil, err
	}
	var result ClubPage
	if err := s.client.do(ctx, http.MethodGet, "/v1/clubs", params, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ClubService) Create(ctx context.Context, input ClubRequest) (*datagsm.Club, error) {
	return s.write(ctx, http.MethodPost, "/v1/clubs", input)
}

// Update replaces the membership list as well as club information.
func (s *ClubService) Update(ctx context.Context, id int64, input ClubRequest) (*datagsm.Club, error) {
	if err := validID(id); err != nil {
		return nil, err
	}
	return s.write(ctx, http.MethodPut, fmt.Sprintf("/v1/clubs/%d", id), input)
}

func (s *ClubService) write(ctx context.Context, method, path string, input ClubRequest) (*datagsm.Club, error) {
	if input.ParticipantIDs == nil {
		input.ParticipantIDs = []int64{}
	}
	var result datagsm.Club
	if err := s.client.do(ctx, method, path, nil, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ClubService) Delete(ctx context.Context, id int64) error {
	if err := validID(id); err != nil {
		return err
	}
	return s.client.do(ctx, http.MethodDelete, fmt.Sprintf("/v1/clubs/%d", id), nil, nil, nil)
}
