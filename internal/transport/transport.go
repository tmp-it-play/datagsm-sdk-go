package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New(baseURL string, client *http.Client) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
		return nil, errors.New("datagsm: invalid base URL")
	}
	loopback := u.Hostname() == "localhost" || net.ParseIP(u.Hostname()).IsLoopback()
	if u.Scheme != "https" && !(u.Scheme == "http" && loopback) {
		return nil, errors.New("datagsm: base URL requires HTTPS (HTTP is allowed on loopback)")
	}
	hc := http.Client{Timeout: 30 * time.Second}
	if client != nil {
		hc = *client
	}
	// Redirects can disclose API keys or replay token and write requests.
	hc.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), HTTP: &hc}, nil
}

func (c *Client) Request(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Request, error) {
	rawURL := c.BaseURL + path
	if len(query) > 0 {
		rawURL += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("datagsm: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) Do(req *http.Request, target any, wrapped bool) error {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("datagsm: request failed: %w", err)
	}
	defer resp.Body.Close()
	return decode(resp, target, wrapped)
}

func decode(resp *http.Response, target any, wrapped bool) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("datagsm: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var failure struct {
			Message     string `json:"message"`
			Code        string `json:"error"`
			Description string `json:"error_description"`
		}
		_ = json.Unmarshal(body, &failure)
		message := failure.Description
		if message == "" {
			message = failure.Message
		}
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
		return &datagsm.APIError{StatusCode: resp.StatusCode, Code: failure.Code, Message: message, RetryAfter: resp.Header.Get("Retry-After")}
	}
	if target == nil {
		return nil
	}
	if wrapped {
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			return fmt.Errorf("datagsm: decode envelope: %w", err)
		}
		body = envelope.Data
	}
	if len(body) == 0 || bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
		return errors.New("datagsm: response data is missing")
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("datagsm: decode response: %w", err)
	}
	return nil
}
