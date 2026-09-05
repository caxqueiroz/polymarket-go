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
	builderCode  string
	httpClient   *http.Client
	clobBaseURL  string
	gammaBaseURL string
	dataBaseURL  string
	creds        *Credentials
}

// WithBuilderCode sets the default public code for GetBuilderTrades queries.
// For order attribution, also set CreateOrderParams.Builder before signing.
func WithBuilderCode(code string) Option {
	return func(o *options) { o.builderCode = code }
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

// WithCredentials sets API key credentials for authenticated CLOB requests.
func WithCredentials(c *Credentials) Option {
	return func(o *options) { o.creds = c }
}

// WithBuilderCredentials is retained for source compatibility and has no effect.
// Deprecated: CLOB V2 attribution uses CreateOrderParams.Builder in the signed
// order. Legacy POLY_BUILDER_* headers are no longer sent by this client.
func WithBuilderCredentials(c *BuilderCredentials) Option {
	return func(o *options) {}
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
		creds:      o.creds,
	}

	return &Client{
		Clob:  &ClobClient{base: base, baseURL: o.clobBaseURL, builderCode: o.builderCode},
		Gamma: &GammaClient{base: base, baseURL: o.gammaBaseURL},
		Data:  &DataClient{base: base, baseURL: o.dataBaseURL},
	}
}

// baseClient holds the shared HTTP logic.
type baseClient struct {
	httpClient *http.Client
	creds      *Credentials
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

	if b.creds != nil {
		if err := signRequest(req, b.creds, ""); err != nil {
			return err
		}
	}

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

	if b.creds != nil {
		if err := signRequest(req, b.creds, ""); err != nil {
			return nil, err
		}
	}

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

	if b.creds != nil {
		if err := signRequest(req, b.creds, string(data)); err != nil {
			return err
		}
	}

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

	if b.creds != nil {
		if err := signRequest(req, b.creds, string(data)); err != nil {
			return nil, err
		}
	}

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

// setL1Headers injects L1 (EIP-712) authentication headers onto an http.Request.
// These are used instead of L2 (HMAC) headers for API key management endpoints.
func setL1Headers(req *http.Request, l1 *L1AuthHeaders) {
	req.Header.Set("POLY_ADDRESS", l1.Address)
	req.Header.Set("POLY_SIGNATURE", l1.Signature)
	req.Header.Set("POLY_TIMESTAMP", l1.Timestamp)
	req.Header.Set("POLY_NONCE", l1.Nonce)
}

func (b *baseClient) getL1(ctx context.Context, baseURL, path string, params url.Values, l1 *L1AuthHeaders, out any) error {
	u := baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("polymarket: creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	setL1Headers(req, l1)

	return b.do(req, out)
}

func (b *baseClient) postL1(ctx context.Context, baseURL, path string, payload any, l1 *L1AuthHeaders, out any) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("polymarket: encoding request body: %w", err)
		}
		body = bytes.NewReader(data)
	}

	u := baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, body)
	if err != nil {
		return fmt.Errorf("polymarket: creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	setL1Headers(req, l1)

	return b.do(req, out)
}

func (b *baseClient) deleteL1(ctx context.Context, baseURL, path string, l1 *L1AuthHeaders) error {
	u := baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return fmt.Errorf("polymarket: creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	setL1Headers(req, l1)

	return b.do(req, nil)
}

func (b *baseClient) deleteRaw(ctx context.Context, baseURL, path string, payload any) ([]byte, error) {
	var data []byte
	var err error
	if payload != nil {
		data, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("polymarket: encoding request body: %w", err)
		}
	}

	u := baseURL + path
	var req *http.Request
	if len(data) > 0 {
		req, err = http.NewRequestWithContext(ctx, http.MethodDelete, u, bytes.NewReader(data))
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("polymarket: creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if len(data) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	if b.creds != nil {
		if err := signRequest(req, b.creds, string(data)); err != nil {
			return nil, err
		}
	}

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
