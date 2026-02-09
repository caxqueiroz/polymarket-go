package polymarket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	DefaultClobBaseURL  = "https://clob.polymarket.com"
	DefaultGammaBaseURL = "https://gamma-api.polymarket.com"
	DefaultDataBaseURL  = "https://data-api.polymarket.com"
)

// Client is the top-level Polymarket client that provides access to all APIs.
type Client struct {
	Clob  *ClobClient
	Gamma *GammaClient
	Data  *DataClient
}

// Option configures the Client.
type Option func(*options)

type options struct {
	httpClient   *http.Client
	clobBaseURL  string
	gammaBaseURL string
	dataBaseURL  string
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) { o.httpClient = c }
}

// WithClobBaseURL overrides the CLOB API base URL.
func WithClobBaseURL(u string) Option {
	return func(o *options) { o.clobBaseURL = u }
}

// WithGammaBaseURL overrides the Gamma API base URL.
func WithGammaBaseURL(u string) Option {
	return func(o *options) { o.gammaBaseURL = u }
}

// WithDataBaseURL overrides the Data API base URL.
func WithDataBaseURL(u string) Option {
	return func(o *options) { o.dataBaseURL = u }
}

// NewClient creates a new Polymarket client.
func NewClient(opts ...Option) *Client {
	o := &options{
		httpClient:   http.DefaultClient,
		clobBaseURL:  DefaultClobBaseURL,
		gammaBaseURL: DefaultGammaBaseURL,
		dataBaseURL:  DefaultDataBaseURL,
	}
	for _, fn := range opts {
		fn(o)
	}

	base := &baseClient{
		httpClient: o.httpClient,
	}

	return &Client{
		Clob:  &ClobClient{base: base, baseURL: o.clobBaseURL},
		Gamma: &GammaClient{base: base, baseURL: o.gammaBaseURL},
		Data:  &DataClient{base: base, baseURL: o.dataBaseURL},
	}
}

// baseClient holds the shared HTTP logic.
type baseClient struct {
	httpClient *http.Client
}

func (b *baseClient) get(ctx context.Context, baseURL, path string, params url.Values, out any) error {
	u := baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("polymarket: creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	return b.do(req, out)
}

func (b *baseClient) getRaw(ctx context.Context, baseURL, path string, params url.Values) ([]byte, error) {
	u := baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("polymarket: creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("polymarket: sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("polymarket: reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
		}
	}

	return body, nil
}

func (b *baseClient) postJSON(ctx context.Context, baseURL, path string, payload any, out any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("polymarket: encoding request body: %w", err)
	}

	u := baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("polymarket: creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return b.do(req, out)
}

func (b *baseClient) postJSONRaw(ctx context.Context, baseURL, path string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("polymarket: encoding request body: %w", err)
	}

	u := baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("polymarket: creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("polymarket: sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("polymarket: reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
		}
	}

	return body, nil
}

func (b *baseClient) do(req *http.Request, out any) error {
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("polymarket: sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("polymarket: reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
		}
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("polymarket: decoding response: %w", err)
		}
	}

	return nil
}
