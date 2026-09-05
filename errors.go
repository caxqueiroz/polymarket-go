package polymarket

import (
	"fmt"
	"net/http"
)

// CancelError reports rejected or unacknowledged cancellations, including HTTP
// 200 responses with not_canceled entries. Batch methods also return the IDs
// that were successfully canceled.
type CancelError struct {
	NotCanceled map[string]string
}

func (e *CancelError) Error() string {
	return fmt.Sprintf("polymarket: %d order cancellations not confirmed", len(e.NotCanceled))
}

// APIError represents an error response from the Polymarket API.
type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("polymarket: %s: %s", e.Status, e.Body)
	}
	return fmt.Sprintf("polymarket: %s", e.Status)
}

// IsNotFound reports whether the error is a 404 Not Found.
func IsNotFound(err error) bool {
	if e, ok := err.(*APIError); ok {
		return e.StatusCode == http.StatusNotFound
	}
	return false
}

// IsRateLimited reports whether the error is a 429 Too Many Requests.
func IsRateLimited(err error) bool {
	if e, ok := err.(*APIError); ok {
		return e.StatusCode == http.StatusTooManyRequests
	}
	return false
}
