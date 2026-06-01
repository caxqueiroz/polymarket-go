# Builder API Order Attribution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Polymarket Builder API order attribution so all orders submitted via the SDK include `POLY_BUILDER_*` headers for volume tracking and rewards.

**Architecture:** The builder signing is identical to existing HMAC auth (`hmacSign`) — same algorithm (HMAC-SHA256), same message format (`timestamp + method + path + body`), same base64url encoding. The only differences are: (1) four headers instead of five (no Address), and (2) different header prefix (`POLY_BUILDER_*`). We add `BuilderCredentials` to `auth.go`, a new option to `client.go`, inject builder headers alongside existing auth headers in `postJSON`/`postJSONRaw`/`deleteRaw`, and add `GetBuilderTrades` to `clob.go`.

**Tech Stack:** Go 1.25, standard library only (crypto/hmac, crypto/sha256, encoding/base64)

**Reference:** Python SDK source at `py_builder_signing_sdk/signing/hmac.py` confirms the HMAC is identical to the existing `hmacSign` function.

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `auth.go` | Modify (lines 1-56) | Add `BuilderCredentials` struct + `signBuilderRequest()` |
| `auth_test.go` | Modify (lines 1-138) | Add builder signing tests |
| `client.go` | Modify (lines 29-90) | Add `builderCreds` to options/baseClient, new option func, inject in HTTP methods |
| `clob.go` | Modify | Add `GetBuilderTrades()` method |
| `clob_test.go` | Modify | Add builder header assertion helper + `GetBuilderTrades` test + `PostOrder` with builder test |

---

## Task 1: Add `BuilderCredentials` struct and `signBuilderRequest` to `auth.go`

**Files:**
- Modify: `auth.go:1-56`
- Test: `auth_test.go`

- [ ] **Step 1: Write failing tests for builder signing**

Add to `auth_test.go`:

```go
var testBuilderSecret = base64.URLEncoding.EncodeToString([]byte("test-builder-secret-1234"))

var testBuilderCreds = &BuilderCredentials{
	Key:        "test-builder-key",
	Secret:     testBuilderSecret,
	Passphrase: "test-builder-passphrase",
}

func TestSignBuilderRequest(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://clob.polymarket.com/order", nil)

	if err := signBuilderRequest(req, testBuilderCreds, `{"size":"10"}`); err != nil {
		t.Fatalf("signBuilderRequest() error: %v", err)
	}

	headers := []string{"POLY_BUILDER_API_KEY", "POLY_BUILDER_SIGNATURE", "POLY_BUILDER_TIMESTAMP", "POLY_BUILDER_PASSPHRASE"}
	for _, h := range headers {
		if req.Header.Get(h) == "" {
			t.Errorf("header %s is empty", h)
		}
	}

	if req.Header.Get("POLY_BUILDER_API_KEY") != "test-builder-key" {
		t.Errorf("POLY_BUILDER_API_KEY = %q, want %q", req.Header.Get("POLY_BUILDER_API_KEY"), "test-builder-key")
	}
	if req.Header.Get("POLY_BUILDER_PASSPHRASE") != "test-builder-passphrase" {
		t.Errorf("POLY_BUILDER_PASSPHRASE = %q, want %q", req.Header.Get("POLY_BUILDER_PASSPHRASE"), "test-builder-passphrase")
	}
}

func TestSignBuilderRequestNoAddress(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://clob.polymarket.com/order", nil)

	if err := signBuilderRequest(req, testBuilderCreds, ""); err != nil {
		t.Fatalf("signBuilderRequest() error: %v", err)
	}

	// Builder headers should NOT include POLY_ADDRESS or POLY_BUILDER_ADDRESS.
	if req.Header.Get("POLY_ADDRESS") != "" {
		t.Error("builder signing should not set POLY_ADDRESS")
	}
}

func TestSignBuilderRequestSignatureMatchesHmac(t *testing.T) {
	// Builder signing uses the same HMAC algorithm as regular signing.
	req, _ := http.NewRequest("POST", "https://clob.polymarket.com/order", nil)
	body := `{"token_id":"123"}`

	if err := signBuilderRequest(req, testBuilderCreds, body); err != nil {
		t.Fatalf("signBuilderRequest() error: %v", err)
	}

	ts := req.Header.Get("POLY_BUILDER_TIMESTAMP")
	expectedSig, err := hmacSign(testBuilderSecret, ts, "POST", "/order", body)
	if err != nil {
		t.Fatalf("hmacSign() error: %v", err)
	}

	if got := req.Header.Get("POLY_BUILDER_SIGNATURE"); got != expectedSig {
		t.Errorf("POLY_BUILDER_SIGNATURE = %q, want %q", got, expectedSig)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestSignBuilder" ./...`
Expected: FAIL — `BuilderCredentials` and `signBuilderRequest` undefined

- [ ] **Step 3: Implement `BuilderCredentials` and `signBuilderRequest`**

Add to `auth.go` after the `Credentials` struct (after line 19):

```go
// BuilderCredentials holds builder API key credentials for order attribution.
// These are separate from user API credentials — builder credentials attribute
// trading volume to the builder account for tracking and rewards.
type BuilderCredentials struct {
	Key        string // Builder API key (UUID)
	Secret     string // Base64url-encoded HMAC secret
	Passphrase string // Builder passphrase
}

// signBuilderRequest injects the 4 POLY_BUILDER_* attribution headers onto an http.Request.
// The HMAC signature uses the same algorithm as signRequest: HMAC-SHA256 over
// timestamp + method + path + body, with base64url encoding.
func signBuilderRequest(req *http.Request, creds *BuilderCredentials, body string) error {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	sig, err := hmacSign(creds.Secret, timestamp, req.Method, req.URL.Path, body)
	if err != nil {
		return err
	}

	req.Header.Set("POLY_BUILDER_API_KEY", creds.Key)
	req.Header.Set("POLY_BUILDER_SIGNATURE", sig)
	req.Header.Set("POLY_BUILDER_TIMESTAMP", timestamp)
	req.Header.Set("POLY_BUILDER_PASSPHRASE", creds.Passphrase)

	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestSignBuilder" ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add auth.go auth_test.go
git commit -m "feat(auth): add BuilderCredentials and signBuilderRequest for order attribution"
```

---

## Task 2: Add `WithBuilderCredentials` option and propagate to `baseClient`

**Files:**
- Modify: `client.go:29-90`
- Test: `auth_test.go` (or inline in `client.go` tests)

- [ ] **Step 1: Write failing test for the option**

Add to `auth_test.go`:

```go
func TestWithBuilderCredentials(t *testing.T) {
	c := NewClient(
		WithBuilderCredentials(&BuilderCredentials{
			Key:        "builder-key",
			Secret:     testBuilderSecret,
			Passphrase: "builder-pass",
		}),
	)

	if c.Clob.base.builderCreds == nil {
		t.Fatal("builderCreds not propagated to ClobClient.base")
	}
	if c.Clob.base.builderCreds.Key != "builder-key" {
		t.Errorf("builderCreds.Key = %q, want %q", c.Clob.base.builderCreds.Key, "builder-key")
	}
}

func TestWithoutBuilderCredentials(t *testing.T) {
	c := NewClient()

	if c.Clob.base.builderCreds != nil {
		t.Error("builderCreds should be nil when not configured")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestWith.*BuilderCredentials" ./...`
Expected: FAIL — `WithBuilderCredentials` undefined, `builderCreds` field missing

- [ ] **Step 3: Implement option and field**

In `client.go`:

Add `builderCreds` to the `options` struct (after line 34):
```go
type options struct {
	httpClient   *http.Client
	clobBaseURL  string
	gammaBaseURL string
	dataBaseURL  string
	creds        *Credentials
	builderCreds *BuilderCredentials
}
```

Add option function (after `WithCredentials`, after line 60):
```go
// WithBuilderCredentials sets builder API key credentials for order attribution.
// Builder credentials are separate from user credentials — they attribute trading
// volume to the builder account. When set, POLY_BUILDER_* headers are automatically
// added to order submission requests.
func WithBuilderCredentials(c *BuilderCredentials) Option {
	return func(o *options) { o.builderCreds = c }
}
```

Add `builderCreds` to `baseClient` struct (modify line 87-90):
```go
type baseClient struct {
	httpClient   *http.Client
	creds        *Credentials
	builderCreds *BuilderCredentials
}
```

Propagate in `NewClient` (modify line 74-77):
```go
base := &baseClient{
	httpClient:   o.httpClient,
	creds:        o.creds,
	builderCreds: o.builderCreds,
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestWith.*BuilderCredentials" ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add client.go auth_test.go
git commit -m "feat(client): add WithBuilderCredentials option and propagate to baseClient"
```

---

## Task 3: Inject builder headers in HTTP methods (`postJSON`, `postJSONRaw`, `deleteRaw`)

**Files:**
- Modify: `client.go:153-174` (postJSON), `client.go:176-216` (postJSONRaw), `client.go:308-359` (deleteRaw)
- Test: `clob_test.go`

- [ ] **Step 1: Write failing test — postJSON injects builder headers**

Add to `clob_test.go`:

```go
func newTestClobClientAuthWithBuilder(handler http.Handler) *ClobClient {
	srv := httptest.NewServer(handler)
	c := NewClient(
		WithHTTPClient(srv.Client()),
		WithClobBaseURL(srv.URL),
		WithCredentials(testCreds),
		WithBuilderCredentials(testBuilderCreds),
	)
	return c.Clob
}

func requireBuilderHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	for _, h := range []string{"POLY_BUILDER_API_KEY", "POLY_BUILDER_SIGNATURE", "POLY_BUILDER_TIMESTAMP", "POLY_BUILDER_PASSPHRASE"} {
		if r.Header.Get(h) == "" {
			t.Errorf("missing builder header %s", h)
		}
	}
}

func requireNoBuilderHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	for _, h := range []string{"POLY_BUILDER_API_KEY", "POLY_BUILDER_SIGNATURE", "POLY_BUILDER_TIMESTAMP", "POLY_BUILDER_PASSPHRASE"} {
		if r.Header.Get(h) != "" {
			t.Errorf("unexpected builder header %s = %q", h, r.Header.Get(h))
		}
	}
}

func TestPostOrderWithBuilderHeaders(t *testing.T) {
	clob := newTestClobClientAuthWithBuilder(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		requireBuilderHeaders(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true,"orderID":"order-123"}`))
	}))

	signed := &SignedOrder{
		Order:     makeTestOrder(),
		Signature: "0xdeadbeef",
	}
	resp, err := clob.PostOrder(context.Background(), signed, OrderGTC)
	if err != nil {
		t.Fatalf("PostOrder() error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestPostOrderWithoutBuilderHeaders(t *testing.T) {
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		requireNoBuilderHeaders(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true,"orderID":"order-456"}`))
	}))

	signed := &SignedOrder{
		Order:     makeTestOrder(),
		Signature: "0xdeadbeef",
	}
	_, err := clob.PostOrder(context.Background(), signed, OrderGTC)
	if err != nil {
		t.Fatalf("PostOrder() error: %v", err)
	}
}
```

Note: `makeTestOrder()` needs to return a valid `OrderData`. The existing `TestClobPostOrder` lives in `signer_test.go` and uses `signer.CreateOrder()`. Since all test files share the `polymarket` package, you can access `testPrivateKey` from `signer_test.go`. However, for these header-focused tests a minimal `OrderData` is sufficient — we only need the server to receive the request, not validate order fields:

```go
func makeTestOrder() OrderData {
	return OrderData{
		Salt:        big.NewInt(123),
		Maker:       "0xMaker",
		Signer:      "0xSigner",
		Taker:       "0x0000000000000000000000000000000000000000",
		TokenID:     big.NewInt(456),
		MakerAmount: big.NewInt(100),
		TakerAmount: big.NewInt(50),
		Expiration:  big.NewInt(0),
		Nonce:       big.NewInt(0),
		FeeRateBPS:  big.NewInt(0),
		Side:        Buy,
		SignatureType: SignatureTypeEOA,
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestPostOrder(With|Without)BuilderHeaders" ./...`
Expected: FAIL — builder headers not present (postJSON doesn't inject them yet)

- [ ] **Step 3: Inject builder headers in `postJSON`, `postJSONRaw`, and `deleteRaw`**

In `client.go`, modify `postJSON` (after the `signRequest` block, around line 171):

```go
if b.creds != nil {
	if err := signRequest(req, b.creds, string(data)); err != nil {
		return err
	}
}

// Builder attribution headers (order attribution, separate from auth).
if b.builderCreds != nil {
	if err := signBuilderRequest(req, b.builderCreds, string(data)); err != nil {
		return err
	}
}
```

Apply the same pattern to `postJSONRaw` (after line 194) and `deleteRaw` (after line 337).

Do NOT add builder headers to `get`, `getRaw`, `getL1`, `postL1`, or `deleteL1` — builder attribution only applies to order-mutating endpoints which use HMAC auth.

> **Note:** This approach sends builder headers on ALL `postJSON`/`postJSONRaw`/`deleteRaw` calls, including non-order endpoints like `GetPrices`, `GetMidpoints`, etc. This is harmless — the API ignores unknown headers — and simpler than a more surgical approach that would require refactoring individual methods. If this becomes a concern, a future refactor can restrict header injection to specific endpoints.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestPostOrder(With|Without)BuilderHeaders" ./...`
Expected: PASS

- [ ] **Step 5: Run full test suite to check for regressions**

Run: `go test ./...`
Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
git add client.go clob_test.go
git commit -m "feat(client): inject POLY_BUILDER_* headers in postJSON, postJSONRaw, deleteRaw"
```

---

## Task 4: Add `GetBuilderTrades` endpoint to `clob.go`

**Files:**
- Modify: `clob.go`
- Test: `clob_test.go`

- [ ] **Step 1: Write failing test**

Add to `clob_test.go`:

```go
func TestClobGetBuilderTrades(t *testing.T) {
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/builder-trades" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		requireAuthHeaders(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"trade-1","market":"0xabc","side":"BUY","size":"10","price":"0.55","status":"MATCHED"}]`))
	}))

	trades, err := clob.GetBuilderTrades(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetBuilderTrades() error: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].ID != "trade-1" {
		t.Errorf("trade ID = %q, want %q", trades[0].ID, "trade-1")
	}
}

func TestClobGetBuilderTradesWithMarket(t *testing.T) {
	market := "0xabc"
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("market"); got != market {
			t.Errorf("market param = %q, want %q", got, market)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))

	trades, err := clob.GetBuilderTrades(context.Background(), &market)
	if err != nil {
		t.Fatalf("GetBuilderTrades() error: %v", err)
	}
	if len(trades) != 0 {
		t.Errorf("expected 0 trades, got %d", len(trades))
	}
}

func TestClobGetBuilderTradesRequiresAuth(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	}))

	_, err := clob.GetBuilderTrades(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestClobGetBuilderTrades" ./...`
Expected: FAIL — `GetBuilderTrades` undefined

- [ ] **Step 3: Implement `GetBuilderTrades`**

Add to `clob.go` (after `PostOrder`, around line 619):

```go
// GetBuilderTrades returns trades attributed to the builder account.
// Requires authentication. Optionally filter by market condition ID.
func (c *ClobClient) GetBuilderTrades(ctx context.Context, market *string) ([]UserTrade, error) {
	if c.base.creds == nil {
		return nil, fmt.Errorf("polymarket: GetBuilderTrades requires authentication (use WithCredentials)")
	}

	params := url.Values{}
	if market != nil {
		params.Set("market", *market)
	}

	var trades []UserTrade
	if err := c.base.get(ctx, c.baseURL, "/builder-trades", params, &trades); err != nil {
		return nil, err
	}
	return trades, nil
}
```

Note: `UserTrade` already exists in `clob_models.go` — reuse it. Verify by checking that `UserTrade` has `ID`, `Market`, `Side`, `Size`, `Price`, `Status` fields. If the response shape differs, add a `BuilderTrade` type instead.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestClobGetBuilderTrades" ./...`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `go test ./...`
Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
git add clob.go clob_test.go
git commit -m "feat(clob): add GetBuilderTrades endpoint for builder trade attribution"
```

---

## Task 5: Add builder attribution example

**Files:**
- Create: `examples/builder/main.go`

- [ ] **Step 1: Write the example**

```go
// Command builder demonstrates order attribution using Builder API credentials.
//
// Required environment variables:
//
//	POLY_PRIVATE_KEY          - hex-encoded Ethereum private key
//	POLY_BUILDER_API_KEY      - Builder API key (UUID)
//	POLY_BUILDER_API_SECRET   - Base64url-encoded HMAC secret
//	POLY_BUILDER_PASSPHRASE   - Builder passphrase
//
// Usage:
//
//	POLY_PRIVATE_KEY=0x... POLY_BUILDER_API_KEY=... POLY_BUILDER_API_SECRET=... POLY_BUILDER_PASSPHRASE=... go run ./examples/builder
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	polymarket "github.com/caxqueiroz/polymarket-go"
)

func main() {
	privKey := os.Getenv("POLY_PRIVATE_KEY")
	if privKey == "" {
		log.Fatal("POLY_PRIVATE_KEY is required")
	}

	builderKey := os.Getenv("POLY_BUILDER_API_KEY")
	builderSecret := os.Getenv("POLY_BUILDER_API_SECRET")
	builderPassphrase := os.Getenv("POLY_BUILDER_PASSPHRASE")

	signer, err := polymarket.NewOrderSigner(privKey)
	if err != nil {
		log.Fatalf("creating signer: %v", err)
	}

	// Derive API credentials.
	ctx := context.Background()
	tmpClient := polymarket.NewClient()
	apiCreds, err := tmpClient.Clob.CreateOrDeriveApiCreds(ctx, signer)
	if err != nil {
		log.Fatalf("deriving API credentials: %v", err)
	}

	// Build client options.
	opts := []polymarket.Option{
		polymarket.WithCredentials(apiCreds.ToCredentials(signer.Address())),
	}

	// Add builder credentials if configured.
	if builderKey != "" {
		fmt.Println("Builder credentials configured — orders will include attribution headers")
		opts = append(opts, polymarket.WithBuilderCredentials(&polymarket.BuilderCredentials{
			Key:        builderKey,
			Secret:     builderSecret,
			Passphrase: builderPassphrase,
		}))
	} else {
		fmt.Println("No builder credentials — orders will be anonymous")
	}

	client := polymarket.NewClient(opts...)

	// Query builder trades (if builder credentials are configured).
	if builderKey != "" {
		trades, err := client.Clob.GetBuilderTrades(ctx, nil)
		if err != nil {
			log.Fatalf("getting builder trades: %v", err)
		}
		fmt.Printf("Builder trades: %d\n", len(trades))
		for _, t := range trades {
			fmt.Printf("  %s: %s %s @ %s (%s)\n", t.ID, t.Side, t.Size, t.Price, t.Status)
		}
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./examples/builder/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add examples/builder/main.go
git commit -m "docs(examples): add builder attribution example"
```

---

## Task 6: Final validation

- [ ] **Step 1: Run full test suite**

Run: `go test -v -count=1 ./...`
Expected: All tests PASS

- [ ] **Step 2: Run linter**

Run: `golangci-lint run ./...`
Expected: No errors

- [ ] **Step 3: Review all changes**

Run: `git diff main --stat` and `git diff main` to review all changes are correct.

- [ ] **Step 4: Final commit (if any linter fixes needed)**

```bash
git add -A
git commit -m "chore: address linter findings"
```
