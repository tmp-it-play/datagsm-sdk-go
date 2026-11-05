package event

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestSignature(t *testing.T) {
	secret := strings.Repeat("0123456789abcdef", 4)
	body := []byte(`{"x":1}`)
	const signature = "sha256=db2d677ffec1d03eed74795cad990d906c8daf4005b70186e5d0cc7614969ac7"
	if err := VerifySignature(body, signature, secret); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		body              []byte
		signature, secret string
	}{
		{[]byte(`{ "x": 1 }`), signature, secret}, {body, signature, ""}, {body, signature, "different"},
		{body, "", secret}, {body, "sha256=00", secret}, {body, "sha256=invalid", secret}, {body, strings.TrimPrefix(signature, "sha256="), secret},
	} {
		if !errors.Is(VerifySignature(tt.body, tt.signature, tt.secret), ErrInvalidSignature) {
			t.Fatal("accepted invalid signature")
		}
	}
}

func parseSigned(t *testing.T, kind Type, old, new string) (*Event, error) {
	t.Helper()
	body := []byte(fmt.Sprintf(`{"id":"evt_test","event":%q,"timestamp":"2026-06-23T05:21:48.123456Z","data":{"old":%s,"new":%s}}`, kind, old, new))
	mac := hmac.New(sha256.New, []byte("test-secret"))
	_, _ = mac.Write(body)
	return Parse(body, "sha256="+hex.EncodeToString(mac.Sum(nil)), "test-secret")
}

func TestEventChanges(t *testing.T) {
	e, err := parseSigned(t, ClubUpdated,
		`[{"index":7,"object":{"club_id":1,"name":"old"}},{"index":3,"object":{}}]`,
		`[{"index":3,"object":{"club_id":2,"name":"new"}},{"index":7,"object":{}}]`)
	if err != nil {
		t.Fatal(err)
	}
	changes, err := Changes[Club](e)
	if err != nil || len(changes) != 2 {
		t.Fatalf("%v, %v", changes, err)
	}
	if changes[0].Index != 3 || changes[0].Old != nil || changes[0].New.ID != 2 || changes[1].Index != 7 || changes[1].Old.ID != 1 || changes[1].New != nil {
		t.Fatal("snapshots were not matched by index")
	}
	if _, err := Changes[Project](e); err == nil {
		t.Fatal("accepted wrong payload type")
	}
	e, err = parseSigned(t, StudentUpdated, `[{"index":0,"object":{"student_id":1,"grade":3,"role":"GENERAL_STUDENT"}}]`, `[{"index":0,"object":{"student_id":1,"grade":null,"major":null,"role":"FUTURE_ROLE"}}]`)
	if err != nil {
		t.Fatal(err)
	}
	students, err := Changes[Student](e)
	if err != nil || students[0].New.Grade != nil || students[0].New.Major != nil || students[0].New.Role != "FUTURE_ROLE" {
		t.Fatalf("%v, %v", students, err)
	}
	e, err = parseSigned(t, ProjectUpdated, `[{"index":0,"object":{}}]`, `[{"index":0,"object":{"project_id":3,"tech_stacks":["Go"],"deployment_url":null}}]`)
	if err != nil {
		t.Fatal(err)
	}
	projects, err := Changes[Project](e)
	if err != nil || projects[0].Old != nil || projects[0].New.TechStacks[0] != "Go" {
		t.Fatalf("%v, %v", projects, err)
	}
}

func TestEventRejectsMalformedSnapshots(t *testing.T) {
	for _, tt := range []struct{ name, old, new string }{
		{"different lengths", `[]`, `[{"index":0,"object":{"club_id":1}}]`},
		{"missing lists", `null`, `null`},
		{"missing index", `[{"object":{}}]`, `[{"index":0,"object":{"club_id":1}}]`},
		{"negative index", `[{"index":-1,"object":{}}]`, `[{"index":-1,"object":{"club_id":1}}]`},
		{"null object", `[{"index":0,"object":null}]`, `[{"index":0,"object":{"club_id":1}}]`},
		{"nonobject", `[{"index":0,"object":[]}]`, `[{"index":0,"object":{"club_id":1}}]`},
		{"both empty", `[{"index":0,"object":{}}]`, `[{"index":0,"object":{}}]`},
		{"missing ID", `[{"index":0,"object":{}}]`, `[{"index":0,"object":{"name":"x"}}]`},
		{"changed ID", `[{"index":0,"object":{"club_id":1}}]`, `[{"index":0,"object":{"club_id":2}}]`},
		{"duplicate old", `[{"index":0,"object":{}},{"index":0,"object":{}}]`, `[{"index":0,"object":{"club_id":1}},{"index":1,"object":{"club_id":2}}]`},
		{"duplicate new", `[{"index":0,"object":{}},{"index":1,"object":{}}]`, `[{"index":0,"object":{"club_id":1}},{"index":0,"object":{"club_id":2}}]`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseSigned(t, ClubUpdated, tt.old, tt.new); err == nil {
				t.Fatal("accepted malformed event")
			}
		})
	}
	if _, err := parseSigned(t, StudentUpdated, `[{"index":0,"object":{"student_id":1}}]`, `[{"index":0,"object":{}}]`); err == nil {
		t.Fatal("accepted student deletion")
	}
	if _, err := Parse([]byte(`not JSON`), "sha256=invalid", "secret"); !errors.Is(err, ErrInvalidSignature) {
		t.Fatal("parsed body before checking signature")
	}
}

func TestNullableEventParticipants(t *testing.T) {
	var club Club
	if err := json.Unmarshal([]byte(`{"club_id":1,"leader":null,"participants":[{"name":"graduate","student_number":null}]}`), &club); err != nil {
		t.Fatal(err)
	}
	if club.Leader != nil || len(club.Participants) != 1 || club.Participants[0].StudentNumber != nil {
		t.Fatal("nullable participant fields were lost")
	}
}
