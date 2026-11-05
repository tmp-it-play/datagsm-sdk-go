package openapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type NEISService struct{ client *Client }

type TimetableQuery struct {
	Grade    int
	ClassNum int
	Date     string
	FromDate string
	ToDate   string
}

func (q TimetableQuery) params() (url.Values, error) {
	if q.Grade < 1 || q.Grade > 3 || q.ClassNum < 1 || q.ClassNum > 4 {
		return nil, errors.New("datagsm: grade must be 1-3 and classNum must be 1-4")
	}
	if q.Date == "" && q.FromDate == "" && q.ToDate == "" {
		return nil, errors.New("datagsm: timetable requires a date or range")
	}
	p, err := (DateQuery{q.Date, q.FromDate, q.ToDate}).params()
	if err != nil {
		return nil, err
	}
	p.Set("grade", strconv.Itoa(q.Grade))
	p.Set("classNum", strconv.Itoa(q.ClassNum))
	return p, nil
}

func (s *NEISService) Meals(ctx context.Context, query DateQuery) ([]Meal, error) {
	params, err := query.params()
	if err != nil {
		return nil, err
	}
	var result struct {
		Meals []Meal `json:"meals"`
	}
	if err := s.client.do(ctx, http.MethodGet, "/v1/neis/meals", params, nil, &result); err != nil {
		return nil, err
	}
	return result.Meals, nil
}

func (s *NEISService) Schedules(ctx context.Context, query DateQuery) ([]Schedule, error) {
	params, err := query.params()
	if err != nil {
		return nil, err
	}
	var result struct {
		Schedules []Schedule `json:"schedules"`
	}
	if err := s.client.do(ctx, http.MethodGet, "/v1/neis/schedules", params, nil, &result); err != nil {
		return nil, err
	}
	return result.Schedules, nil
}

func (s *NEISService) Timetables(ctx context.Context, query TimetableQuery) ([]Timetable, error) {
	params, err := query.params()
	if err != nil {
		return nil, err
	}
	var result struct {
		Timetables []Timetable `json:"timetables"`
	}
	if err := s.client.do(ctx, http.MethodGet, "/v1/neis/timetables", params, nil, &result); err != nil {
		return nil, err
	}
	return result.Timetables, nil
}
