package openapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNEISWireFormat(t *testing.T) {
	c, err := NewClient(Config{APIKey: "test", HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" || req.Header.Get("X-API-KEY") != "test" {
			t.Fatal("incorrect request authentication or method")
		}
		var body string
		switch req.URL.Path {
		case "/v1/neis/meals":
			body = `{"data":{"meals":[{"mealDate":"2026-11-05","mealMenu":["rice"],"mealAllergyInfo":null}]}}`
		case "/v1/neis/schedules":
			body = `{"data":{"schedules":[{"scheduleDate":"2026-11-05","academicYear":"2026","targetGrades":[1]}]}}`
		case "/v1/neis/timetables":
			body = `{"data":{"timetables":[{"timetableDate":"2026-11-05","academicYear":"2026","semester":null,"subject":"공강"}]}}`
		default:
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	meals, err := c.NEIS.Meals(context.Background(), DateQuery{})
	if err != nil || len(meals) != 1 || meals[0].MealDate != "2026-11-05" {
		t.Fatalf("%v, %v", meals, err)
	}
	schedules, err := c.NEIS.Schedules(context.Background(), DateQuery{})
	if err != nil || len(schedules) != 1 || schedules[0].AcademicYear != "2026" {
		t.Fatalf("%v, %v", schedules, err)
	}
	timetables, err := c.NEIS.Timetables(context.Background(), TimetableQuery{Grade: 1, ClassNum: 1, Date: "2026-11-05"})
	if err != nil || len(timetables) != 1 || timetables[0].Semester != nil || *timetables[0].Subject != "공강" {
		t.Fatalf("%v, %v", timetables, err)
	}
}

func TestWriteRequestEncoding(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	c, err := NewClient(Config{APIKey: "test", HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotMethod, gotPath = req.Method, req.URL.Path
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(body)
		if len(body) > 0 && req.Header.Get("Content-Type") != "application/json" {
			t.Fatal("missing JSON content type")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"data":{"id":1}}`))}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, tt := range []struct {
		name, method, path string
		call               func() error
	}{
		{"create student", "POST", "/v1/students", func() error { _, err := c.Students.Create(ctx, CreateStudentRequest{Name: "test"}); return err }},
		{"update student", "PUT", "/v1/students/1", func() error { _, err := c.Students.Update(ctx, 1, UpdateStudentRequest{Name: "test"}); return err }},
		{"create club", "POST", "/v1/clubs", func() error { _, err := c.Clubs.Create(ctx, ClubRequest{}); return err }},
		{"update club", "PUT", "/v1/clubs/1", func() error { _, err := c.Clubs.Update(ctx, 1, ClubRequest{}); return err }},
		{"delete club", "DELETE", "/v1/clubs/1", func() error { return c.Clubs.Delete(ctx, 1) }},
		{"create project", "POST", "/v1/projects", func() error { _, err := c.Projects.Create(ctx, ProjectRequest{}); return err }},
		{"update project", "PUT", "/v1/projects/1", func() error {
			_, err := c.Projects.Update(ctx, 1, ProjectRequest{Categories: []datagsm.ProjectCategory{}})
			return err
		}},
		{"delete project", "DELETE", "/v1/projects/1", func() error { return c.Projects.Delete(ctx, 1) }},
		{"end project", "POST", "/v1/projects/1/end", func() error { return c.Projects.End(ctx, 1, 2026) }},
		{"reactivate project", "POST", "/v1/projects/1/reactivate", func() error { return c.Projects.Reactivate(ctx, 1) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err != nil {
				t.Fatal(err)
			}
			if gotMethod != tt.method || gotPath != tt.path {
				t.Fatalf("%s %s", gotMethod, gotPath)
			}
			if tt.name == "create club" || tt.name == "create project" {
				var fields map[string]json.RawMessage
				if err := json.Unmarshal([]byte(gotBody), &fields); err != nil {
					t.Fatal(err)
				}
				if string(fields["participantIds"]) != "[]" {
					t.Fatalf("nil participants not encoded as array: %s", gotBody)
				}
			}
			if tt.name == "end project" && gotBody != `{"endYear":2026}` {
				t.Fatal(gotBody)
			}
			if tt.name == "reactivate project" && gotBody != "" {
				t.Fatal(gotBody)
			}
		})
	}
}
