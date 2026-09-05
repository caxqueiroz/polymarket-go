package polymarket

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"
)

// Independent viem vectors using Polymarket's official V2 typed data and
// deposit-wallet envelope, not hashes computed by this SDK.
func TestV2SigningVectors(t *testing.T) {
	data, err := os.ReadFile("testdata/v2-signing.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Vectors []struct {
			Name    string
			NegRisk bool
			Order   struct {
				Salt, Maker, Signer, TokenID, MakerAmount, TakerAmount, Timestamp, Metadata, Builder string
				Side                                                                                 int
				SignatureType                                                                        SignatureType
			}
			Digest, SigningDigest, Signature string
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	signer, err := NewOrderSigner(testPrivateKey, Polygon)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range fixture.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			integer := func(s string) *big.Int {
				v, ok := new(big.Int).SetString(s, 10)
				if !ok {
					t.Fatalf("invalid fixture integer %q", s)
				}
				return v
			}
			side := Buy
			if v.Order.Side == 1 {
				side = Sell
			}
			o := &OrderData{
				Salt: integer(v.Order.Salt), Maker: v.Order.Maker, Signer: v.Order.Signer,
				TokenID: integer(v.Order.TokenID), MakerAmount: integer(v.Order.MakerAmount),
				TakerAmount: integer(v.Order.TakerAmount), Side: side, SignatureType: v.Order.SignatureType,
				Timestamp: integer(v.Order.Timestamp), Metadata: v.Order.Metadata, Builder: v.Order.Builder,
				Expiration: big.NewInt(0), NegRisk: v.NegRisk,
			}
			digest := orderDigest(o, Polygon)
			if got := fmt.Sprintf("0x%x", digest); got != v.Digest {
				t.Errorf("digest = %s, want %s", got, v.Digest)
			}
			signed, err := signer.SignOrder(o)
			if err != nil {
				t.Fatal(err)
			}
			// The SDK uses random ECDSA nonces, so signatures need not be byte-equal.
			// Verify against the independently generated signing digest and envelope.
			got, err := hex.DecodeString(strings.TrimPrefix(signed.Signature, "0x"))
			if err != nil || len(got) < 65 {
				t.Fatalf("invalid signature: %v", err)
			}
			want, err := hex.DecodeString(strings.TrimPrefix(v.Signature, "0x"))
			if err != nil {
				t.Fatal(err)
			}
			if hex.EncodeToString(got[65:]) != hex.EncodeToString(want[65:]) {
				t.Error("incorrect deposit-wallet envelope")
			}
			expectedDigest, err := hex.DecodeString(strings.TrimPrefix(v.SigningDigest, "0x"))
			if err != nil {
				t.Fatal(err)
			}
			address, err := ecRecover(expectedDigest, [32]byte(got[:32]), [32]byte(got[32:64]), got[64]-27)
			if err != nil || address != signer.Address() {
				t.Fatalf("signature recovered %s: %v", address, err)
			}
			// GTD expiry is HTTP-only in V2: changing it must not change the digest.
			o.Expiration = big.NewInt(1900000000)
			if got := orderDigest(o, Polygon); got != digest {
				t.Error("expiration entered the V2 signed struct")
			}
		})
	}
}

func TestV2BuildOrderDefaults(t *testing.T) {
	signer, err := NewOrderSigner(testPrivateKey, Polygon)
	if err != nil {
		t.Fatal(err)
	}
	before := time.Now().UnixMilli()
	o, err := signer.BuildOrder(CreateOrderParams{TokenID: "123", Price: .5, Size: 10, Side: Buy})
	if err != nil {
		t.Fatal(err)
	}
	if o.Timestamp == nil || o.Timestamp.Int64() < before || o.Timestamp.Int64() > time.Now().UnixMilli() {
		t.Fatalf("timestamp must default to current milliseconds: %v", o.Timestamp)
	}
	zero := "0x" + strings.Repeat("0", 64)
	if o.Metadata != zero || o.Builder != zero {
		t.Fatalf("bytes32 defaults: metadata=%s builder=%s", o.Metadata, o.Builder)
	}
	o, err = signer.BuildOrder(CreateOrderParams{TokenID: "123", Price: .5, Size: 10, Side: Buy,
		Maker: "0x1234567890abcdef1234567890abcdef12345678", SignatureType: SignatureTypePoly1271,
		Timestamp: 1780449126930, Builder: "0x" + strings.Repeat("22", 32)})
	if err != nil {
		t.Fatal(err)
	}
	if o.Signer != o.Maker || o.Timestamp.Int64() != 1780449126930 {
		t.Fatalf("deposit wallet fields: %+v", o)
	}
	if _, err := signer.SignOrder(o); err != nil {
		t.Fatal(err)
	}
}

func TestV2BuildOrderRejectsInvalidInput(t *testing.T) {
	signer, err := NewOrderSigner(testPrivateKey, Polygon)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*CreateOrderParams)
	}{
		{"nan price", func(p *CreateOrderParams) { p.Price = math.NaN() }},
		{"nan size", func(p *CreateOrderParams) { p.Size = math.NaN() }},
		{"infinite size", func(p *CreateOrderParams) { p.Size = math.Inf(1) }},
		{"amount overflow", func(p *CreateOrderParams) { p.Size = 1e100 }},
		{"rounded to zero", func(p *CreateOrderParams) { p.Size = 1e-10 }},
		{"invalid side", func(p *CreateOrderParams) { p.Side = "HOLD" }},
		{"negative token", func(p *CreateOrderParams) { p.TokenID = "-1" }},
		{"oversized token", func(p *CreateOrderParams) { p.TokenID = new(big.Int).Lsh(big.NewInt(1), 256).String() }},
		{"invalid maker", func(p *CreateOrderParams) { p.Maker = "0xwrong" }},
		{"legacy fee", func(p *CreateOrderParams) { p.FeeRateBPS = 10 }},
		{"legacy nonce", func(p *CreateOrderParams) { p.Nonce = 1 }},
		{"legacy taker", func(p *CreateOrderParams) { p.Taker = "0x1234567890abcdef1234567890abcdef12345678" }},
		{"invalid metadata", func(p *CreateOrderParams) { p.Metadata = "0x1234" }},
		{"invalid builder", func(p *CreateOrderParams) { p.Builder = "0x" + strings.Repeat("zz", 32) }},
		{"negative timestamp", func(p *CreateOrderParams) { p.Timestamp = -1 }},
		{"negative expiration", func(p *CreateOrderParams) { p.Expiration = -1 }},
		{"unknown signature type", func(p *CreateOrderParams) { p.SignatureType = 4 }},
		{"deposit wallet missing", func(p *CreateOrderParams) { p.SignatureType = SignatureTypePoly1271 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := CreateOrderParams{TokenID: "123", Price: .5, Size: 10, Side: Buy}
			tc.change(&p)
			if _, err := signer.BuildOrder(p); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
}

func TestV2SignOrderRejectsMalformedRawOrders(t *testing.T) {
	signer, err := NewOrderSigner(testPrivateKey, Polygon)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*OrderData)
	}{
		{"missing token", func(o *OrderData) { o.TokenID = nil }},
		{"oversized salt", func(o *OrderData) { o.Salt = new(big.Int).Lsh(big.NewInt(1), 256) }},
		{"negative amount", func(o *OrderData) { o.MakerAmount = big.NewInt(-1) }},
		{"invalid maker", func(o *OrderData) { o.Maker = "0x123" }},
		{"signer mismatch", func(o *OrderData) { o.Signer = "0x1234567890abcdef1234567890abcdef12345678" }},
		{"missing timestamp", func(o *OrderData) { o.Timestamp = nil }},
		{"bad metadata", func(o *OrderData) { o.Metadata = "not hex" }},
		{"legacy nonce", func(o *OrderData) { o.Nonce = big.NewInt(1) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o, err := signer.BuildOrder(CreateOrderParams{TokenID: "123", Price: .5, Size: 10, Side: Buy})
			if err != nil {
				t.Fatal(err)
			}
			tc.change(o)
			if _, err := signer.SignOrder(o); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
	if _, err := signer.SignOrder(nil); err == nil {
		t.Fatal("nil order accepted")
	}
}

func TestV2SignatureRecoversSigner(t *testing.T) {
	signer, err := NewOrderSigner(testPrivateKey, Polygon)
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signer.CreateOrder(CreateOrderParams{TokenID: "123", Price: .5, Size: 10, Side: Buy})
	if err != nil {
		t.Fatal(err)
	}
	b, err := hex.DecodeString(strings.TrimPrefix(signed.Signature, "0x"))
	if err != nil {
		t.Fatal(err)
	}
	digest := orderDigest(&signed.Order, Polygon)
	address, err := ecRecover(digest[:], [32]byte(b[:32]), [32]byte(b[32:64]), b[64]-27)
	if err != nil || address != signer.Address() {
		t.Fatalf("recovered %s: %v", address, err)
	}
}
