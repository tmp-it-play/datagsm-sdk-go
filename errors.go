package datagsm

import "fmt"

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	RetryAfter string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("datagsm: HTTP %d %s: %s", e.StatusCode, e.Code, e.Message)
}
