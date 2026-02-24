package polymarket

import (
	"encoding/base64"
	"net/http"
	"testing"
)

// testSecret is a known base64url-encoded key for deterministic tests.
var testSecret = base64.URLEncoding.EncodeToString([]byte("test-secret-key-1234"))

var testCreds = &Credentials{
	Key:        "test-api-key",
	Secret:     testSecret,
	Passphrase: "test-passphrase",
	Address:    "0xTestAddress",
}

func TestHmacSign(t *testing.T) {
	sig1, err := hmacSign(testSecret, "1700000000", "GET", "/orders", "")
	if err != nil {
		t.Fatalf("hmacSign() error: %v", err)
	}
	if sig1 == "" {
		t.Fatal("hmacSign() returned empty signature")
	}

	// Same inputs produce same signature.
	sig2, err := hmacSign(testSecret, "1700000000", "GET", "/orders", "")
	if err != nil {
		t.Fatalf("hmacSign() error: %v", err)
	}
	if sig1 != sig2 {
		t.Errorf("same inputs produced different signatures: %q vs %q", sig1, sig2)
	}

	// Different timestamp produces different signature.
	sig3, err := hmacSign(testSecret, "1700000001", "GET", "/orders", "")
	if err != nil {
		t.Fatalf("hmacSign() error: %v", err)
	}
	if sig1 == sig3 {
		t.Error("different timestamps produced same signature")
	}
}

func TestHmacSignWithBody(t *testing.T) {
	sigNoBody, err := hmacSign(testSecret, "1700000000", "POST", "/order", "")
	if err != nil {
		t.Fatalf("hmacSign() error: %v", err)
	}

	sigWithBody, err := hmacSign(testSecret, "1700000000", "POST", "/order", `{"size":"10"}`)
	if err != nil {
		t.Fatalf("hmacSign() error: %v", err)
	}

	if sigNoBody == sigWithBody {
		t.Error("body should affect signature")
	}
}

func TestHmacSignInvalidSecret(t *testing.T) {
	_, err := hmacSign("not-valid-base64!!", "1700000000", "GET", "/orders", "")
	if err == nil {
		t.Fatal("expected error for invalid base64 secret")
	}
}

func TestSignRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://clob.polymarket.com/orders?market=0xabc", nil)

	if err := signRequest(req, testCreds, ""); err != nil {
		t.Fatalf("signRequest() error: %v", err)
	}

	headers := []string{"POLY_API_KEY", "POLY_SIGNATURE", "POLY_TIMESTAMP", "POLY_PASSPHRASE", "POLY_ADDRESS"}
	for _, h := range headers {
		if req.Header.Get(h) == "" {
			t.Errorf("header %s is empty", h)
		}
	}

	if req.Header.Get("POLY_API_KEY") != "test-api-key" {
		t.Errorf("POLY_API_KEY = %q, want %q", req.Header.Get("POLY_API_KEY"), "test-api-key")
	}
	if req.Header.Get("POLY_PASSPHRASE") != "test-passphrase" {
		t.Errorf("POLY_PASSPHRASE = %q, want %q", req.Header.Get("POLY_PASSPHRASE"), "test-passphrase")
	}
	if req.Header.Get("POLY_ADDRESS") != "0xTestAddress" {
		t.Errorf("POLY_ADDRESS = %q, want %q", req.Header.Get("POLY_ADDRESS"), "0xTestAddress")
	}
}

func TestSignRequestNoBody(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://clob.polymarket.com/orders", nil)

	if err := signRequest(req, testCreds, ""); err != nil {
		t.Fatalf("signRequest() error: %v", err)
	}

	if req.Header.Get("POLY_SIGNATURE") == "" {
		t.Error("POLY_SIGNATURE should not be empty")
	}
}

func TestSignRequestWithBody(t *testing.T) {
	body := `{"order_id":"abc123"}`
	req, _ := http.NewRequest("POST", "https://clob.polymarket.com/order", nil)

	if err := signRequest(req, testCreds, body); err != nil {
		t.Fatalf("signRequest() error: %v", err)
	}

	// Signature should differ from a request without body.
	req2, _ := http.NewRequest("POST", "https://clob.polymarket.com/order", nil)
	if err := signRequest(req2, testCreds, ""); err != nil {
		t.Fatalf("signRequest() error: %v", err)
	}

	if req.Header.Get("POLY_SIGNATURE") == req2.Header.Get("POLY_SIGNATURE") {
		t.Error("signatures should differ when body differs")
	}
}

func TestSignRequestQueryStringIgnored(t *testing.T) {
	req1, _ := http.NewRequest("GET", "https://clob.polymarket.com/orders?market=0xabc", nil)
	req2, _ := http.NewRequest("GET", "https://clob.polymarket.com/orders", nil)

	signRequest(req1, testCreds, "")
	signRequest(req2, testCreds, "")

	// Query string is NOT part of the HMAC message (matching the Python/TS SDKs),
	// so both requests with the same path should produce the same signature.
	if req1.Header.Get("POLY_SIGNATURE") != req2.Header.Get("POLY_SIGNATURE") {
		t.Error("query string should not affect signature (only path is signed)")
	}
}
