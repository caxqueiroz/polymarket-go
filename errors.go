package polymarket

import (
	"fmt"
	"net/http"
)

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
