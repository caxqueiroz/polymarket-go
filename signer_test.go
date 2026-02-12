package polymarket

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
	"testing"
)

const testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewOrderSigner(t *testing.T) {
	signer, err := NewOrderSigner(testPrivateKey, Polygon)
	if err != nil {
		t.Fatal(err)
	}

	if signer.Address() != "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266" {
		t.Fatalf("address = %s, want 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", signer.Address())
	}
}

func TestCalculateOrderAmounts(t *testing.T) {
	tests := []struct {
		name      string
		price     float64
		size      float64
		side      Side
		wantMaker int64
		wantTaker int64
	}{
		{"buy 0.50 x 10", 0.50, 10.0, Buy, 5_000_000, 10_000_000},
		{"sell 0.50 x 10", 0.50, 10.0, Sell, 10_000_000, 5_000_000},
		{"buy 0.75 x 100", 0.75, 100.0, Buy, 75_000_000, 100_000_000},
		{"sell 0.25 x 50", 0.25, 50.0, Sell, 50_000_000, 12_500_000},
		{"buy 0.01 x 1", 0.01, 1.0, Buy, 10_000, 1_000_000},
		{"sell 0.99 x 1", 0.99, 1.0, Sell, 1_000_000, 990_000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maker, taker := calculateOrderAmounts(tt.price, tt.size, tt.side)
			if maker.Int64() != tt.wantMaker {
				t.Errorf("makerAmount = %d, want %d", maker.Int64(), tt.wantMaker)
			}
			if taker.Int64() != tt.wantTaker {
				t.Errorf("takerAmount = %d, want %d", taker.Int64(), tt.wantTaker)
			}
		})
	}
}

func TestBuildOrder(t *testing.T) {
	signer, _ := NewOrderSigner(testPrivateKey, Polygon)

	order, err := signer.BuildOrder(CreateOrderParams{
		TokenID: "12345",
		Price:   0.50,
		Size:    10.0,
		Side:    Buy,
	})
	if err != nil {
		t.Fatal(err)
	}

	if order.Salt.Sign() == 0 {
		t.Error("salt should not be zero")
	}
	if order.Maker != signer.Address() {
		t.Errorf("maker = %s, want %s", order.Maker, signer.Address())
	}
	if order.Signer != signer.Address() {
		t.Errorf("signer = %s, want %s", order.Signer, signer.Address())
	}
	if order.Taker != zeroAddress {
		t.Errorf("taker = %s, want %s", order.Taker, zeroAddress)
	}
	if order.TokenID.Int64() != 12345 {
		t.Errorf("tokenID = %d, want 12345", order.TokenID.Int64())
	}
	if order.MakerAmount.Int64() != 5_000_000 {
		t.Errorf("makerAmount = %d, want 5000000", order.MakerAmount.Int64())
	}
	if order.TakerAmount.Int64() != 10_000_000 {
		t.Errorf("takerAmount = %d, want 10000000", order.TakerAmount.Int64())
	}
	if order.Side != Buy {
		t.Errorf("side = %s, want BUY", order.Side)
	}
}

func TestBuildOrderWithMakerOverride(t *testing.T) {
	signer, _ := NewOrderSigner(testPrivateKey, Polygon)
	proxyWallet := "0x1234567890abcdef1234567890abcdef12345678"

	order, err := signer.BuildOrder(CreateOrderParams{
		TokenID:       "123",
		Price:         0.50,
		Size:          10.0,
		Side:          Buy,
		Maker:         proxyWallet,
		SignatureType: SignatureTypePoly,
	})
	if err != nil {
		t.Fatal(err)
	}

	if order.Maker != checksumAddress(proxyWallet) {
		t.Errorf("maker = %s, want %s", order.Maker, checksumAddress(proxyWallet))
	}
	if order.Signer != signer.Address() {
		t.Errorf("signer should still be the signer's address")
	}
	if order.SignatureType != SignatureTypePoly {
		t.Errorf("signatureType = %d, want %d", order.SignatureType, SignatureTypePoly)
	}
}

func TestBuildOrderValidation(t *testing.T) {
	signer, _ := NewOrderSigner(testPrivateKey, Polygon)

	tests := []struct {
		name   string
		params CreateOrderParams
	}{
		{"price zero", CreateOrderParams{TokenID: "123", Price: 0, Size: 10, Side: Buy}},
		{"price 1", CreateOrderParams{TokenID: "123", Price: 1, Size: 10, Side: Buy}},
		{"price negative", CreateOrderParams{TokenID: "123", Price: -0.5, Size: 10, Side: Buy}},
		{"size zero", CreateOrderParams{TokenID: "123", Price: 0.5, Size: 0, Side: Buy}},
		{"size negative", CreateOrderParams{TokenID: "123", Price: 0.5, Size: -10, Side: Buy}},
		{"invalid tokenID", CreateOrderParams{TokenID: "abc", Price: 0.5, Size: 10, Side: Buy}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := signer.BuildOrder(tt.params)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestDomainSeparatorDeterministic(t *testing.T) {
	ds1 := computeDomainSeparator(137, "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E")
	ds2 := computeDomainSeparator(137, "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E")

	if ds1 != ds2 {
		t.Fatal("domain separator not deterministic")
	}

	// Different chain ID should produce different separator
	ds3 := computeDomainSeparator(80002, "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E")
	if ds1 == ds3 {
		t.Fatal("different chain IDs should produce different domain separators")
	}

	// Different exchange address should produce different separator
	ds4 := computeDomainSeparator(137, "0xC5d563A36AE78145C45a50134d48A1215220f80a")
	if ds1 == ds4 {
		t.Fatal("different exchanges should produce different domain separators")
	}
}

func TestOrderHashDeterministic(t *testing.T) {
	order := &OrderData{
		Salt:          big.NewInt(12345),
		Maker:         "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Signer:        "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Taker:         zeroAddress,
		TokenID:       big.NewInt(67890),
		MakerAmount:   big.NewInt(5_000_000),
		TakerAmount:   big.NewInt(10_000_000),
		Expiration:    big.NewInt(0),
		Nonce:         big.NewInt(0),
		FeeRateBPS:    big.NewInt(0),
		Side:          Buy,
		SignatureType: SignatureTypeEOA,
	}

	h1 := hashOrder(order)
	h2 := hashOrder(order)

	if h1 != h2 {
		t.Fatal("order hash not deterministic")
	}

	// Different order should produce different hash
	order2 := *order
	order2.MakerAmount = big.NewInt(6_000_000)
	h3 := hashOrder(&order2)
	if h1 == h3 {
		t.Fatal("different orders should produce different hashes")
	}
}

func TestOrderDigestRegularVsNegRisk(t *testing.T) {
	order := &OrderData{
		Salt:          big.NewInt(12345),
		Maker:         "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Signer:        "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Taker:         zeroAddress,
		TokenID:       big.NewInt(67890),
		MakerAmount:   big.NewInt(5_000_000),
		TakerAmount:   big.NewInt(10_000_000),
		Expiration:    big.NewInt(0),
		Nonce:         big.NewInt(0),
		FeeRateBPS:    big.NewInt(0),
		Side:          Buy,
		SignatureType: SignatureTypeEOA,
		NegRisk:       false,
	}

	d1 := orderDigest(order, Polygon)

	order.NegRisk = true
	d2 := orderDigest(order, Polygon)

	if d1 == d2 {
		t.Fatal("regular and neg-risk orders should have different digests")
	}
}

func TestCreateOrderSignature(t *testing.T) {
	signer, _ := NewOrderSigner(testPrivateKey, Polygon)

	signed, err := signer.CreateOrder(CreateOrderParams{
		TokenID: "71321045679252212594626385532706912750332728571942532289631379312455583992563",
		Price:   0.50,
		Size:    10.0,
		Side:    Buy,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Signature format: 0x + 130 hex chars (65 bytes: r + s + v)
	if !strings.HasPrefix(signed.Signature, "0x") {
		t.Errorf("signature should start with 0x")
	}
	if len(signed.Signature) != 132 {
		t.Errorf("signature length = %d, want 132 (0x + 130 hex chars)", len(signed.Signature))
	}

	// Verify v is 27 or 28
	sigBytes, _ := hex.DecodeString(signed.Signature[2:])
	v := sigBytes[64]
	if v != 27 && v != 28 {
		t.Errorf("v = %d, want 27 or 28", v)
	}

	// Verify owner is the maker
	if signed.Owner != signed.Order.Maker {
		t.Errorf("owner = %s, want %s", signed.Owner, signed.Order.Maker)
	}

	// Recover signer address from signature to verify it's correct
	digest := orderDigest(&signed.Order, Polygon)
	var r, s [32]byte
	copy(r[:], sigBytes[:32])
	copy(s[:], sigBytes[32:64])
	recovered, err := ecRecover(digest[:], r, s, v-27)
	if err != nil {
		t.Fatalf("ecRecover error: %v", err)
	}
	if recovered != signer.Address() {
		t.Errorf("recovered = %s, want %s", recovered, signer.Address())
	}
}

func TestCreateOrderNegRisk(t *testing.T) {
	signer, _ := NewOrderSigner(testPrivateKey, Polygon)

	signed, err := signer.CreateOrder(CreateOrderParams{
		TokenID: "12345",
		Price:   0.50,
		Size:    10.0,
		Side:    Sell,
		NegRisk: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify signature against neg-risk domain
	digest := orderDigest(&signed.Order, Polygon)
	sigBytes, _ := hex.DecodeString(signed.Signature[2:])
	var r, s [32]byte
	copy(r[:], sigBytes[:32])
	copy(s[:], sigBytes[32:64])
	recovered, err := ecRecover(digest[:], r, s, sigBytes[64]-27)
	if err != nil {
		t.Fatalf("ecRecover error: %v", err)
	}
	if recovered != signer.Address() {
		t.Errorf("recovered = %s, want %s", recovered, signer.Address())
	}
}

func TestAbiEncode(t *testing.T) {
	// uint256(1) should be 31 zero bytes + 0x01
	b := abiEncodeUint256(big.NewInt(1))
	if len(b) != 32 {
		t.Fatalf("length = %d, want 32", len(b))
	}
	if b[31] != 1 {
		t.Errorf("last byte = %d, want 1", b[31])
	}
	for i := 0; i < 31; i++ {
		if b[i] != 0 {
			t.Errorf("byte[%d] = %d, want 0", i, b[i])
		}
	}

	// address should be left-padded to 32 bytes
	addr := abiEncodeAddress("0xff00000000000000000000000000000000000001")
	if len(addr) != 32 {
		t.Fatalf("address length = %d, want 32", len(addr))
	}
	if addr[12] != 0xff {
		t.Errorf("addr[12] = 0x%02x, want 0xff", addr[12])
	}
	if addr[31] != 0x01 {
		t.Errorf("addr[31] = 0x%02x, want 0x01", addr[31])
	}
}

func TestClobPostOrder(t *testing.T) {
	// PostOrder requires authentication — use newTestClobClientAuth.
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/order" {
			t.Errorf("path = %s, want /order", r.URL.Path)
		}

		var req postOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		if req.OrderType != "GTC" {
			t.Errorf("orderType = %s, want GTC", req.OrderType)
		}
		if req.Order.Side != "BUY" {
			t.Errorf("side = %s, want BUY", req.Order.Side)
		}
		if !strings.HasPrefix(req.Order.Signature, "0x") {
			t.Error("signature should start with 0x")
		}

		// Owner should be the API key, not the wallet address.
		if req.Owner != testCreds.Key {
			t.Errorf("owner = %q, want API key %q", req.Owner, testCreds.Key)
		}

		// Salt should be a valid number (json.Number marshals without quotes).
		if saltStr := string(req.Order.Salt); saltStr == "" {
			t.Error("salt should not be empty")
		}

		w.Write([]byte(`{"success":true,"orderID":"order-abc-123"}`))
	}))

	signer, _ := NewOrderSigner(testPrivateKey, Polygon)
	signed, _ := signer.CreateOrder(CreateOrderParams{
		TokenID: "12345",
		Price:   0.50,
		Size:    10.0,
		Side:    Buy,
	})

	resp, err := clob.PostOrder(context.Background(), signed, OrderGTC)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Success {
		t.Error("success = false")
	}
	if resp.OrderID != "order-abc-123" {
		t.Errorf("orderID = %s, want order-abc-123", resp.OrderID)
	}
}

func TestClobPostOrderRequiresAuth(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server without credentials")
	}))

	signer, _ := NewOrderSigner(testPrivateKey, Polygon)
	signed, _ := signer.CreateOrder(CreateOrderParams{
		TokenID: "12345",
		Price:   0.50,
		Size:    10.0,
		Side:    Buy,
	})

	_, err := clob.PostOrder(context.Background(), signed, OrderGTC)
	if err == nil {
		t.Fatal("PostOrder without credentials should fail")
	}
}

func TestEIP712TypeHashes(t *testing.T) {
	// Verify the type hash constants are valid keccak256 hashes of the type strings
	domainType := "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
	got := keccak256([]byte(domainType))
	if got != eip712DomainTypeHash {
		t.Error("domain type hash mismatch")
	}

	nameHash := keccak256([]byte("Polymarket CTF Exchange"))
	if nameHash != eip712DomainName {
		t.Error("domain name hash mismatch")
	}

	versionHash := keccak256([]byte("1"))
	if versionHash != eip712DomainVersion {
		t.Error("domain version hash mismatch")
	}

	orderType := "Order(uint256 salt,address maker,address signer,address taker," +
		"uint256 tokenId,uint256 makerAmount,uint256 takerAmount," +
		"uint256 expiration,uint256 nonce,uint256 feeRateBps," +
		"uint8 side,uint8 signatureType)"
	got = keccak256([]byte(orderType))
	if got != orderTypeHash {
		t.Error("order type hash mismatch")
	}
}

// ---- ClobAuth EIP-712 Tests ----

func TestClobAuthTypeHashes(t *testing.T) {
	// Verify ClobAuth domain type hash
	domainType := "EIP712Domain(string name,string version,uint256 chainId)"
	got := keccak256([]byte(domainType))
	if got != clobAuthDomainTypeHash {
		t.Error("ClobAuth domain type hash mismatch")
	}

	nameHash := keccak256([]byte("ClobAuthDomain"))
	if nameHash != clobAuthDomainName {
		t.Error("ClobAuth domain name hash mismatch")
	}

	authType := "ClobAuth(address address,string timestamp,uint256 nonce,string message)"
	got = keccak256([]byte(authType))
	if got != clobAuthTypeHash {
		t.Error("ClobAuth type hash mismatch")
	}
}

func TestClobAuthDomainSeparatorDeterministic(t *testing.T) {
	ds1 := computeClobAuthDomainSeparator(137)
	ds2 := computeClobAuthDomainSeparator(137)
	if ds1 != ds2 {
		t.Fatal("ClobAuth domain separator not deterministic")
	}

	// Different chain ID → different separator
	ds3 := computeClobAuthDomainSeparator(80002)
	if ds1 == ds3 {
		t.Fatal("different chain IDs should produce different ClobAuth domain separators")
	}
}

func TestClobAuthDomainDiffersFromOrderDomain(t *testing.T) {
	// ClobAuth domain has no verifyingContract; order domain does.
	clobAuth := computeClobAuthDomainSeparator(137)
	order := computeDomainSeparator(137, Polygon.Exchange)
	if clobAuth == order {
		t.Fatal("ClobAuth and order domain separators should differ")
	}
}

func TestClobAuthDigestDeterministic(t *testing.T) {
	d1 := clobAuthDigest("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", "1700000000", 0, 137)
	d2 := clobAuthDigest("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", "1700000000", 0, 137)
	if d1 != d2 {
		t.Fatal("ClobAuth digest not deterministic")
	}

	// Different timestamp → different digest
	d3 := clobAuthDigest("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", "1700000001", 0, 137)
	if d1 == d3 {
		t.Fatal("different timestamps should produce different digests")
	}

	// Different nonce → different digest
	d4 := clobAuthDigest("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", "1700000000", 1, 137)
	if d1 == d4 {
		t.Fatal("different nonces should produce different digests")
	}

	// Different address → different digest
	d5 := clobAuthDigest("0x0000000000000000000000000000000000000001", "1700000000", 0, 137)
	if d1 == d5 {
		t.Fatal("different addresses should produce different digests")
	}
}

func TestBuildL1AuthHeaders(t *testing.T) {
	signer, _ := NewOrderSigner(testPrivateKey, Polygon)

	l1, err := signer.BuildL1AuthHeaders("1700000000", 0)
	if err != nil {
		t.Fatal(err)
	}

	// Address should match signer
	if l1.Address != signer.Address() {
		t.Errorf("Address = %s, want %s", l1.Address, signer.Address())
	}

	// Timestamp should match
	if l1.Timestamp != "1700000000" {
		t.Errorf("Timestamp = %s, want 1700000000", l1.Timestamp)
	}

	// Nonce should match
	if l1.Nonce != "0" {
		t.Errorf("Nonce = %s, want 0", l1.Nonce)
	}

	// Signature format: 0x + 130 hex chars (65 bytes)
	if !strings.HasPrefix(l1.Signature, "0x") {
		t.Error("signature should start with 0x")
	}
	if len(l1.Signature) != 132 {
		t.Errorf("signature length = %d, want 132", len(l1.Signature))
	}

	// Verify v is 27 or 28
	sigBytes, _ := hex.DecodeString(l1.Signature[2:])
	v := sigBytes[64]
	if v != 27 && v != 28 {
		t.Errorf("v = %d, want 27 or 28", v)
	}

	// Verify signature can be recovered to the correct address
	digest := clobAuthDigest(signer.Address(), "1700000000", 0, 137)
	var r, s [32]byte
	copy(r[:], sigBytes[:32])
	copy(s[:], sigBytes[32:64])
	recovered, err := ecRecover(digest[:], r, s, v-27)
	if err != nil {
		t.Fatalf("ecRecover error: %v", err)
	}
	if recovered != signer.Address() {
		t.Errorf("recovered = %s, want %s", recovered, signer.Address())
	}
}

func TestBuildL1AuthHeadersDifferentNonces(t *testing.T) {
	signer, _ := NewOrderSigner(testPrivateKey, Polygon)

	l1a, _ := signer.BuildL1AuthHeaders("1700000000", 0)
	l1b, _ := signer.BuildL1AuthHeaders("1700000000", 1)

	// Different nonces → different signatures (probabilistically different due to random k,
	// but the digest is certainly different which we tested above)
	if l1a.Nonce != "0" {
		t.Errorf("Nonce = %s, want 0", l1a.Nonce)
	}
	if l1b.Nonce != "1" {
		t.Errorf("Nonce = %s, want 1", l1b.Nonce)
	}
}
