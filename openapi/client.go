// Package openapi provides DataGSM data queries and administrative writes.
package openapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tmp-it-play/datagsm-sdk-go/internal/transport"
)

const DefaultBaseURL = "https://openapi.datagsm.kr"

type Config struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

type Client struct {
	Students  *StudentService
	Clubs     *ClubService
	Projects  *ProjectService
	NEIS      *NEISService
	transport *transport.Client
	apiKey    string
}

func NewClient(config Config) (*Client, error) {
	if strings.TrimSpace(config.APIKey) == "" || strings.ContainsAny(config.APIKey, "\r\n") {
		return nil, errors.New("datagsm: API key is required and must not contain newlines")
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}
	t, err := transport.New(config.BaseURL, config.HTTPClient)
	if err != nil {
		return nil, err
	}
	c := &Client{transport: t, apiKey: config.APIKey}
	c.Students = &StudentService{c}
	c.Clubs = &ClubService{c}
	c.Projects = &ProjectService{c}
	c.NEIS = &NEISService{c}
	return c, nil
}

func (c *Client) Close() { c.transport.HTTP.CloseIdleConnections() }

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, target any) error {
	var encoded []byte
	if body != nil {
		var err error
		encoded, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("datagsm: encode request: %w", err)
		}
	}
	req, err := c.transport.Request(ctx, method, path, query, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("X-API-KEY", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.transport.Do(req, target, true)
}
