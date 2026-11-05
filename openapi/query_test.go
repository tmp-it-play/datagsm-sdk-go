package openapi

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

func TestQueryEncoding(t *testing.T) {
	tests := []struct {
		name   string
		params func() (url.Values, error)
		want   url.Values
	}{
		{"student", (StudentQuery{StudentID: datagsm.Ptr(int64(9)), Name: datagsm.Ptr("a&b"), Grade: datagsm.Ptr(2), Specialty: datagsm.Ptr("Go"), GitHubID: datagsm.Ptr("octocat"), DormitoryRoom: datagsm.Ptr(301), IncludeGraduates: datagsm.Ptr(true), OnlyEnrolled: datagsm.Ptr(false), SortDirection: datagsm.SortDirDESC}).params,
			url.Values{"studentId": {"9"}, "name": {"a&b"}, "grade": {"2"}, "specialty": {"Go"}, "githubId": {"octocat"}, "dormitoryRoom": {"301"}, "includeGraduates": {"true"}, "onlyEnrolled": {"false"}, "sortDirection": {"DESC"}, "page": {"0"}}},
		{"club", (ClubQuery{ClubStatus: datagsm.Ptr(datagsm.ClubStatusAbolished), IncludeLeaderInParticipants: datagsm.Ptr(false), Size: 10}).params,
			url.Values{"clubStatus": {"ABOLISHED"}, "includeLeaderInParticipants": {"false"}, "page": {"0"}, "size": {"10"}}},
		{"project", (ProjectQuery{Status: datagsm.Ptr(datagsm.ProjectStatusEnded), Category: datagsm.Ptr(datagsm.ProjectCategoryTeam)}).params,
			url.Values{"status": {"ENDED"}, "category": {"TEAM"}, "page": {"0"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.params()
			if err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("params = %v, %v; want %v", got, err, tt.want)
			}
			decoded, err := url.ParseQuery(got.Encode())
			if err != nil || !reflect.DeepEqual(decoded, tt.want) {
				t.Fatalf("query escaping failed: %v, %v", decoded, err)
			}
		})
	}
}

func TestPaginationBounds(t *testing.T) {
	for _, pair := range [][2]int{{-1, 0}, {0, -1}, {0, 1001}} {
		if _, err := pagination(pair[0], pair[1], "", ""); err == nil {
			t.Fatalf("accepted %v", pair)
		}
	}
	for _, size := range []int{0, 1, 1000} {
		if _, err := pagination(0, size, "", ""); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDateQueries(t *testing.T) {
	for _, q := range []DateQuery{
		{Date: "2026-11-05", FromDate: "2026-11-01"}, {Date: "2026-02-29"},
		{FromDate: "2026-11-06", ToDate: "2026-11-05"}, {FromDate: "2025-01-01", ToDate: "2026-01-02"},
	} {
		if _, err := q.params(); err == nil {
			t.Fatalf("accepted invalid date query %+v", q)
		}
	}
	for _, q := range []DateQuery{{}, {Date: "2024-02-29"}, {FromDate: "2026-11-05"}, {ToDate: "2026-11-05"}, {FromDate: "2025-01-01", ToDate: "2026-01-01"}} {
		if _, err := q.params(); err != nil {
			t.Fatal(err)
		}
	}
	for _, q := range []TimetableQuery{{Grade: 1, ClassNum: 1}, {Grade: 0, ClassNum: 1, Date: "2026-11-05"}, {Grade: 1, ClassNum: 5, Date: "2026-11-05"}} {
		if _, err := q.params(); err == nil {
			t.Fatalf("accepted invalid timetable query %+v", q)
		}
	}
	p, err := (TimetableQuery{Grade: 2, ClassNum: 3, Date: "2026-11-05"}).params()
	if err != nil || p.Get("date") != "2026-11-05" || p.Get("grade") != "2" || p.Get("classNum") != "3" {
		t.Fatalf("%v, %v", p, err)
	}
}

func TestProjectUpdateFieldPresence(t *testing.T) {
	for _, tt := range []struct {
		name                         string
		input                        ProjectRequest
		categories, icon, deployment string
	}{
		{"preserve", ProjectRequest{}, "", "", ""},
		{"clear", ProjectRequest{Categories: []datagsm.ProjectCategory{}, IconKey: datagsm.Ptr(""), DeploymentURL: datagsm.Ptr("")}, "[]", `""`, `""`},
		{"set", ProjectRequest{Categories: []datagsm.ProjectCategory{datagsm.ProjectCategoryTeam}, IconKey: datagsm.Ptr("project-icons/id.png")}, `["TEAM"]`, `"project-icons/id.png"`, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(body, &fields); err != nil {
				t.Fatal(err)
			}
			if string(fields["categories"]) != tt.categories || string(fields["iconKey"]) != tt.icon || string(fields["deploymentUrl"]) != tt.deployment {
				t.Fatalf("unexpected field presence: %s", body)
			}
		})
	}
}
