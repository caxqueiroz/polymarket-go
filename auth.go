package polymarket

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Credentials holds API key credentials for authenticated CLOB requests.
type Credentials struct {
	Key        string // API key (UUID)
	Secret     string // Base64url-encoded HMAC secret
	Passphrase string // Passphrase
	Address    string // Wallet address (hex, e.g. "0x...")
}

// hmacSign computes the POLY_SIGNATURE for a request.
// message = timestamp + method + requestPath [+ body]
func hmacSign(secret, timestamp, method, requestPath, body string) (string, error) {
	key, err := base64.URLEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("polymarket: decoding API secret: %w", err)
	}

	message := timestamp + method + requestPath + body

	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	digest := h.Sum(nil)

	return base64.URLEncoding.EncodeToString(digest), nil
}

// signRequest injects the 5 POLY_* authentication headers onto an http.Request.
func signRequest(req *http.Request, creds *Credentials, body string) error {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	requestPath := req.URL.Path
	if req.URL.RawQuery != "" {
		requestPath += "?" + req.URL.RawQuery
	}

	sig, err := hmacSign(creds.Secret, timestamp, req.Method, requestPath, body)
	if err != nil {
		return err
	}

	req.Header.Set("POLY_API_KEY", creds.Key)
	req.Header.Set("POLY_SIGNATURE", sig)
	req.Header.Set("POLY_TIMESTAMP", timestamp)
	req.Header.Set("POLY_PASSPHRASE", creds.Passphrase)
	req.Header.Set("POLY_ADDRESS", creds.Address)

	return nil
}
