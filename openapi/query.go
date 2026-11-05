package openapi

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

func pagination(page, size int, sortBy string, direction datagsm.SortDirection) (url.Values, error) {
	if page < 0 || size < 0 || size > 1000 {
		return nil, errors.New("datagsm: page must be nonnegative and size must be 0 (default) or 1-1000")
	}
	q := url.Values{"page": {strconv.Itoa(page)}}
	if size != 0 {
		q.Set("size", strconv.Itoa(size))
	}
	if sortBy != "" {
		q.Set("sortBy", sortBy)
	}
	if direction != "" {
		q.Set("sortDirection", string(direction))
	}
	return q, nil
}

func set[T ~string | ~int | ~int64 | ~bool](q url.Values, name string, value *T) {
	if value != nil {
		q.Set(name, fmt.Sprint(*value))
	}
}

// DateQuery uses YYYY-MM-DD dates without a time zone.
type DateQuery struct {
	Date     string
	FromDate string
	ToDate   string
}

func (q DateQuery) params() (url.Values, error) {
	if q.Date != "" && (q.FromDate != "" || q.ToDate != "") {
		return nil, errors.New("datagsm: date and date range cannot be combined")
	}
	p := url.Values{}
	for key, value := range map[string]string{"date": q.Date, "fromDate": q.FromDate, "toDate": q.ToDate} {
		if value == "" {
			continue
		}
		if _, err := time.Parse(time.DateOnly, value); err != nil {
			return nil, fmt.Errorf("datagsm: %s must use YYYY-MM-DD", key)
		}
		p.Set(key, value)
	}
	if q.FromDate != "" && q.ToDate != "" {
		from, _ := time.Parse(time.DateOnly, q.FromDate)
		to, _ := time.Parse(time.DateOnly, q.ToDate)
		if to.Before(from) || to.Sub(from) > 365*24*time.Hour {
			return nil, errors.New("datagsm: date range must be ordered and at most 365 days")
		}
	}
	return p, nil
}

func validID(id int64) error {
	if id <= 0 {
		return errors.New("datagsm: ID must be positive")
	}
	return nil
}
