package event

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const SignatureHeader = "X-DataGSM-Signature"

var ErrInvalidSignature = errors.New("datagsm: invalid event signature")

type Event struct {
	ID        string    `json:"id"`
	Type      Type      `json:"event"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		Old []Snapshot `json:"old"`
		New []Snapshot `json:"new"`
	} `json:"data"`
}

type Snapshot struct {
	Index  *int            `json:"index"`
	Object json.RawMessage `json:"object"`
}

type Payload interface{ Student | Club | Project }

// Change uses nil for the empty object on creation or deletion.
type Change[T Payload] struct {
	Index int
	Old   *T
	New   *T
}

func VerifySignature(body []byte, signature, secret string) error {
	if secret == "" || !strings.HasPrefix(signature, "sha256=") {
		return ErrInvalidSignature
	}
	actual, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil || len(actual) != sha256.Size {
		return ErrInvalidSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	if !hmac.Equal(actual, mac.Sum(nil)) {
		return ErrInvalidSignature
	}
	return nil
}

func Parse(body []byte, signature, secret string) (*Event, error) {
	if err := VerifySignature(body, signature, secret); err != nil {
		return nil, err
	}
	var result Event
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("datagsm: decode event: %w", err)
	}
	if result.ID == "" || result.Timestamp.IsZero() {
		return nil, errors.New("datagsm: event ID and timestamp are required")
	}
	var err error
	switch result.Type {
	case StudentUpdated:
		_, err = Changes[Student](&result)
	case ClubUpdated:
		_, err = Changes[Club](&result)
	case ProjectUpdated:
		_, err = Changes[Project](&result)
	default:
		err = errors.New("datagsm: unsupported event type")
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func Changes[T Payload](e *Event) ([]Change[T], error) {
	var zero T
	var expected Type
	switch any(zero).(type) {
	case Student:
		expected = StudentUpdated
	case Club:
		expected = ClubUpdated
	case Project:
		expected = ProjectUpdated
	}
	if e == nil || e.Type != expected {
		return nil, errors.New("datagsm: event type does not match payload type")
	}
	if e.Data.Old == nil || e.Data.New == nil || len(e.Data.Old) != len(e.Data.New) {
		return nil, errors.New("datagsm: event snapshots must have equal lengths")
	}
	old := make(map[int]*T, len(e.Data.Old))
	for _, snapshot := range e.Data.Old {
		if snapshot.Index == nil || *snapshot.Index < 0 {
			return nil, errors.New("datagsm: invalid event index")
		}
		if _, exists := old[*snapshot.Index]; exists {
			return nil, errors.New("datagsm: duplicate event index")
		}
		value, err := decodeObject[T](snapshot.Object)
		if err != nil {
			return nil, err
		}
		old[*snapshot.Index] = value
	}
	changes := make([]Change[T], 0, len(e.Data.New))
	for _, snapshot := range e.Data.New {
		if snapshot.Index == nil {
			return nil, errors.New("datagsm: missing event index")
		}
		before, exists := old[*snapshot.Index]
		if !exists {
			return nil, errors.New("datagsm: unmatched or duplicate event index")
		}
		after, err := decodeObject[T](snapshot.Object)
		if err != nil {
			return nil, err
		}
		if before == nil && after == nil || expected == StudentUpdated && after == nil {
			return nil, errors.New("datagsm: invalid empty event object")
		}
		if before != nil && after != nil && payloadID(*before) != payloadID(*after) {
			return nil, errors.New("datagsm: event object IDs do not match")
		}
		changes = append(changes, Change[T]{Index: *snapshot.Index, Old: before, New: after})
		delete(old, *snapshot.Index)
	}
	return changes, nil
}

func decodeObject[T Payload](raw json.RawMessage) (*T, error) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || object == nil {
		return nil, errors.New("datagsm: event snapshot must be an object")
	}
	if len(object) == 0 {
		return nil, nil
	}
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("datagsm: decode event object: %w", err)
	}
	if payloadID(value) <= 0 {
		return nil, errors.New("datagsm: event object ID must be positive")
	}
	return &value, nil
}

func payloadID[T Payload](value T) int64 {
	switch v := any(value).(type) {
	case Student:
		return v.ID
	case Club:
		return v.ID
	case Project:
		return v.ID
	}
	return 0
}
