package eventexample

import (
	"errors"
	"io"
	"net/http"

	"github.com/tmp-it-play/datagsm-sdk-go/event"
)

// process must durably handle or enqueue the event, using its ID to prevent duplicate processing.
func Handler(secret string, process func(*event.Event) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4<<20))
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if len(r.Header.Values(event.SignatureHeader)) != 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		e, err := event.Parse(body, r.Header.Get(event.SignatureHeader), secret)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, event.ErrInvalidSignature) {
				status = http.StatusUnauthorized
			}
			http.Error(w, "invalid event", status)
			return
		}
		if err := process(e); err != nil {
			http.Error(w, "processing failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
