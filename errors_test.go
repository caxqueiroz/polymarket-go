package polymarket

import (
	"errors"
	"net/http"
	"testing"
)

func TestAPIErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  APIError
		want string
	}{
		{
			name: "with body",
			err:  APIError{StatusCode: 404, Status: "404 Not Found", Body: "market not found"},
			want: "polymarket: 404 Not Found: market not found",
		},
		{
			name: "without body",
			err:  APIError{StatusCode: 500, Status: "500 Internal Server Error"},
			want: "polymarket: 500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "404 error",
			err:  &APIError{StatusCode: http.StatusNotFound},
			want: true,
		},
		{
			name: "500 error",
			err:  &APIError{StatusCode: http.StatusInternalServerError},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("something"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "429 error",
			err:  &APIError{StatusCode: http.StatusTooManyRequests},
			want: true,
		},
		{
			name: "500 error",
			err:  &APIError{StatusCode: http.StatusInternalServerError},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("something"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimited(tt.err); got != tt.want {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.want)
			}
		})
	}
}
