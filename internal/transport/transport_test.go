package transport

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

func TestDecodeResponse(t *testing.T) {
	for _, tt := range []struct {
		name, body string
		wrapped    bool
		status     int
		wantError  bool
	}{
		{"wrapped", `{"data":{"id":7}}`, true, 200, false},
		{"raw", `{"id":7}`, false, 200, false},
		{"missing data", `{"id":7}`, true, 200, true},
		{"null data", `{"data":null}`, true, 200, true},
		{"bad JSON", `{`, true, 200, true},
		{"empty result", ``, false, 204, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var target struct {
				ID int64 `json:"id"`
			}
			resp := &http.Response{StatusCode: tt.status, Body: io.NopCloser(strings.NewReader(tt.body))}
			err := decode(resp, &target, tt.wrapped)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v", err)
			}
			if err == nil && target.ID != 7 {
				t.Fatalf("ID = %d", target.ID)
			}
		})
	}
	if err := decode(&http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader(""))}, nil, true); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeAPIError(t *testing.T) {
	for _, tt := range []struct{ body, code, message string }{
		{`{"message":"denied"}`, "", "denied"},
		{`{"error":"invalid_grant","error_description":"expired"}`, "invalid_grant", "expired"},
		{`<html>bad gateway</html>`, "", "Too Many Requests"},
	} {
		resp := &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {"60"}}, Body: io.NopCloser(strings.NewReader(tt.body))}
		err := decode(resp, nil, false)
		var apiErr *datagsm.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != 429 || apiErr.Code != tt.code || apiErr.Message != tt.message || apiErr.RetryAfter != "60" {
			t.Fatalf("%+v", err)
		}
	}
}

func TestClientTrustBoundary(t *testing.T) {
	for _, raw := range []string{"", "/relative", "ftp://example.com", "http://example.com", "https://user:secret@example.com", "https://example.com?key=value", "https://example.com#fragment"} {
		if _, err := New(raw, nil); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	for _, raw := range []string{"https://example.com/", "http://localhost:8080", "http://127.0.0.1", "http://[::1]"} {
		hc := &http.Client{Timeout: time.Second}
		c, err := New(raw, hc)
		if err != nil {
			t.Fatal(err)
		}
		if c.HTTP == hc || hc.CheckRedirect != nil || c.HTTP.Timeout != hc.Timeout {
			t.Fatal("caller HTTP client was mutated")
		}
		if c.HTTP.CheckRedirect(nil, nil) != http.ErrUseLastResponse {
			t.Fatal("redirect allowed")
		}
	}
}
