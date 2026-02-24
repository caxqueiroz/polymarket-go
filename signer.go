package polymarket

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"
)

// ChainConfig holds chain-specific contract addresses for order signing.
type ChainConfig struct {
	ChainID         int    // e.g. 137 for Polygon
	Exchange        string // CTF Exchange address (regular markets)
	NegRiskExchange string // CTF Exchange address (neg-risk markets)
}

// Polygon is the chain configuration for Polygon mainnet.
var Polygon = ChainConfig{
	ChainID:         137,
	Exchange:        "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E",
	NegRiskExchange: "0xC5d563A36AE78145C45a50134d48A1215220f80a",
}

// SignatureType identifies how the order signature was created.
type SignatureType int

const (
	// SignatureTypeEOA is for externally owned account (direct wallet) signatures.
	SignatureTypeEOA SignatureType = 0
	// SignatureTypePoly is for Polymarket proxy wallet signatures.
	SignatureTypePoly SignatureType = 1
	// SignatureTypeGnosisSafe is for Gnosis Safe multisig signatures.
	SignatureTypeGnosisSafe SignatureType = 2
)

// OrderType specifies the time-in-force behavior of an order.
type OrderType string

const (
	OrderGTC OrderType = "GTC" // Good Till Cancelled
	OrderGTD OrderType = "GTD" // Good Till Date
	OrderFOK OrderType = "FOK" // Fill Or Kill
	OrderIOC OrderType = "IOC" // Immediate Or Cancel
)

// ---- EIP-712 typed data hashing ----
//
// Polymarket's CTF Exchange uses EIP-712 typed structured data signing.
// Domain: { name: "Polymarket CTF Exchange", version: "1", chainId, verifyingContract }
// Type:   Order(uint256 salt, address maker, address signer, address taker,
//               uint256 tokenId, uint256 makerAmount, uint256 takerAmount,
//               uint256 expiration, uint256 nonce, uint256 feeRateBps,
//               uint8 side, uint8 signatureType)

var (
	eip712DomainTypeHash = keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))
	eip712DomainName     = keccak256([]byte("Polymarket CTF Exchange"))
	eip712DomainVersion  = keccak256([]byte("1"))
	orderTypeHash        = keccak256([]byte(
		"Order(uint256 salt,address maker,address signer,address taker," +
			"uint256 tokenId,uint256 makerAmount,uint256 takerAmount," +
			"uint256 expiration,uint256 nonce,uint256 feeRateBps," +
			"uint8 side,uint8 signatureType)"))
)

// computeDomainSeparator computes the EIP-712 domain separator.
func computeDomainSeparator(chainID int, exchange string) [32]byte {
	var buf []byte
	buf = append(buf, eip712DomainTypeHash[:]...)
	buf = append(buf, eip712DomainName[:]...)
	buf = append(buf, eip712DomainVersion[:]...)
	buf = append(buf, abiEncodeUint256(big.NewInt(int64(chainID)))...)
	buf = append(buf, abiEncodeAddress(exchange)...)
	return keccak256(buf)
}

// abiEncodeUint256 encodes a big.Int as a 32-byte big-endian uint256.
func abiEncodeUint256(v *big.Int) []byte {
	b := make([]byte, 32)
	if v != nil && v.Sign() > 0 {
		vBytes := v.Bytes()
		copy(b[32-len(vBytes):], vBytes)
	}
	return b
}

// abiEncodeAddress encodes an Ethereum address (hex string) as a 32-byte ABI-encoded value.
func abiEncodeAddress(addr string) []byte {
	addr = strings.TrimPrefix(strings.ToLower(addr), "0x")
	addrBytes, _ := hex.DecodeString(addr)
	b := make([]byte, 32)
	copy(b[32-len(addrBytes):], addrBytes)
	return b
}

// abiEncodeUint8 encodes a uint8 as a 32-byte ABI-encoded value.
func abiEncodeUint8(v uint8) []byte {
	b := make([]byte, 32)
	b[31] = v
	return b
}

// ---- Order Data ----

// OrderData represents the EIP-712 order struct that gets hashed and signed.
type OrderData struct {
	Salt          *big.Int
	Maker         string // Ethereum address (0x...)
	Signer        string // Ethereum address (0x...)
	Taker         string // Ethereum address (0x...)
	TokenID       *big.Int
	MakerAmount   *big.Int
	TakerAmount   *big.Int
	Expiration    *big.Int
	Nonce         *big.Int
	FeeRateBPS    *big.Int
	Side          Side
	SignatureType SignatureType
	NegRisk       bool // determines which exchange domain to use
}

// hashOrder computes the EIP-712 struct hash of an order.
func hashOrder(o *OrderData) [32]byte {
	sideVal := uint8(0)
	if o.Side == Sell {
		sideVal = 1
	}

	var buf []byte
	buf = append(buf, orderTypeHash[:]...)
	buf = append(buf, abiEncodeUint256(o.Salt)...)
	buf = append(buf, abiEncodeAddress(o.Maker)...)
	buf = append(buf, abiEncodeAddress(o.Signer)...)
	buf = append(buf, abiEncodeAddress(o.Taker)...)
	buf = append(buf, abiEncodeUint256(o.TokenID)...)
	buf = append(buf, abiEncodeUint256(o.MakerAmount)...)
	buf = append(buf, abiEncodeUint256(o.TakerAmount)...)
	buf = append(buf, abiEncodeUint256(o.Expiration)...)
	buf = append(buf, abiEncodeUint256(o.Nonce)...)
	buf = append(buf, abiEncodeUint256(o.FeeRateBPS)...)
	buf = append(buf, abiEncodeUint8(sideVal)...)
	buf = append(buf, abiEncodeUint8(uint8(o.SignatureType))...)
	return keccak256(buf)
}

// orderDigest computes the EIP-712 digest (the 32-byte hash to sign).
func orderDigest(o *OrderData, chain ChainConfig) [32]byte {
	exchange := chain.Exchange
	if o.NegRisk {
		exchange = chain.NegRiskExchange
	}
	domainSep := computeDomainSeparator(chain.ChainID, exchange)
	structHash := hashOrder(o)

	var buf []byte
	buf = append(buf, 0x19, 0x01) // EIP-712 prefix
	buf = append(buf, domainSep[:]...)
	buf = append(buf, structHash[:]...)
	return keccak256(buf)
}

const zeroAddress = "0x0000000000000000000000000000000000000000"

// ---- Signed Order ----

// SignedOrder is an order with its EIP-712 signature, ready to be posted to the CLOB.
type SignedOrder struct {
	Order     OrderData
	Signature string // 0x-prefixed hex encoded signature (65 bytes: r‖s‖v)
	Owner     string // maker address
}

// ---- Order Signer ----

// OrderSigner builds and signs orders for the Polymarket CLOB using EIP-712 typed data.
type OrderSigner struct {
	privKey *big.Int
	address string
	chain   ChainConfig
}

// NewOrderSigner creates a new OrderSigner from a hex-encoded private key.
// The chain parameter determines which contract addresses to use for the EIP-712 domain.
func NewOrderSigner(privateKeyHex string, chain ChainConfig) (*OrderSigner, error) {
	privKey, err := parsePrivateKey(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return &OrderSigner{
		privKey: privKey,
		address: privateKeyToAddress(privKey),
		chain:   chain,
	}, nil
}

// Address returns the Ethereum address derived from the signer's private key.
func (s *OrderSigner) Address() string {
	return s.address
}

// CreateOrderParams holds the user-facing parameters for creating a new order.
type CreateOrderParams struct {
	TokenID       string        // CLOB token ID (large integer as string)
	Price         float64       // Price per share (0 < price < 1)
	Size          float64       // Number of shares (must be > 0)
	Side          Side          // Buy or Sell
	FeeRateBPS    int           // Fee rate in basis points (default 0)
	Nonce         int           // Nonce (default 0)
	Expiration    int64         // Unix timestamp, 0 = no expiration
	Taker         string        // Taker address, empty = any taker
	NegRisk       bool          // True if this is a neg-risk market
	SignatureType SignatureType // Default: SignatureTypeEOA
	Maker         string        // Maker address override (for POLY_PROXY; defaults to signer address)
}

// CreateOrder builds, signs, and returns a SignedOrder ready for posting.
// This is the primary method for creating orders.
func (s *OrderSigner) CreateOrder(p CreateOrderParams) (*SignedOrder, error) {
	order, err := s.BuildOrder(p)
	if err != nil {
		return nil, err
	}
	return s.SignOrder(order)
}

// BuildOrder creates an unsigned OrderData from the given parameters.
// Use SignOrder to sign it afterwards, or use CreateOrder for both steps.
func (s *OrderSigner) BuildOrder(p CreateOrderParams) (*OrderData, error) {
	if p.Price <= 0 || p.Price >= 1 {
		return nil, fmt.Errorf("polymarket: price must be between 0 and 1 (exclusive), got %f", p.Price)
	}
	if p.Size <= 0 {
		return nil, fmt.Errorf("polymarket: size must be positive, got %f", p.Size)
	}

	tokenID := new(big.Int)
	if _, ok := tokenID.SetString(p.TokenID, 10); !ok {
		return nil, fmt.Errorf("polymarket: invalid tokenID: %s", p.TokenID)
	}

	// Generate random salt – match Python client: round(timestamp * random_float).
	// Uses a small value (< 2^53) so the server's JSON parser doesn't lose precision.
	salt, err := generateSalt()
	if err != nil {
		return nil, fmt.Errorf("polymarket: generating salt: %w", err)
	}

	makerAmount, takerAmount := calculateOrderAmounts(p.Price, p.Size, p.Side)

	maker := p.Maker
	if maker == "" {
		maker = s.address
	}

	taker := p.Taker
	if taker == "" {
		taker = zeroAddress
	}

	return &OrderData{
		Salt:          salt,
		Maker:         checksumAddress(maker),
		Signer:        s.address,
		Taker:         checksumAddress(taker),
		TokenID:       tokenID,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		Expiration:    big.NewInt(p.Expiration),
		Nonce:         big.NewInt(int64(p.Nonce)),
		FeeRateBPS:    big.NewInt(int64(p.FeeRateBPS)),
		Side:          p.Side,
		SignatureType: p.SignatureType,
		NegRisk:       p.NegRisk,
	}, nil
}

// SignOrder signs an OrderData and returns a SignedOrder.
func (s *OrderSigner) SignOrder(order *OrderData) (*SignedOrder, error) {
	digest := orderDigest(order, s.chain)

	r, sig_s, v, err := ecdsaSignCompact(digest[:], s.privKey)
	if err != nil {
		return nil, fmt.Errorf("polymarket: signing order: %w", err)
	}

	// Pack signature: r(32) + s(32) + v(1) where v = 27 or 28
	sigBytes := make([]byte, 65)
	copy(sigBytes[:32], r[:])
	copy(sigBytes[32:64], sig_s[:])
	sigBytes[64] = v + 27

	return &SignedOrder{
		Order:     *order,
		Signature: "0x" + hex.EncodeToString(sigBytes),
		Owner:     order.Maker,
	}, nil
}

// ---- CLOB L1 Authentication (EIP-712) ----
//
// The CLOB API uses EIP-712 signatures for L1 authentication. This is used to
// create or derive API credentials (L2 keys) from a wallet's private key.
//
// Domain: { name: "ClobAuthDomain", version: "1", chainId }
// Type:   ClobAuth(address address, string timestamp, uint256 nonce, string message)
// Message: "This message attests that I control the given wallet"
//
// Reference: https://github.com/Polymarket/py-clob-client
//            https://docs.polymarket.com/developers/CLOB/authentication

var (
	clobAuthDomainTypeHash = keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId)"))
	clobAuthDomainName     = keccak256([]byte("ClobAuthDomain"))
	clobAuthDomainVersion  = keccak256([]byte("1"))
	clobAuthTypeHash       = keccak256([]byte("ClobAuth(address address,string timestamp,uint256 nonce,string message)"))
	clobAuthMessageText    = "This message attests that I control the given wallet"
)

// computeClobAuthDomainSeparator computes the EIP-712 domain separator for CLOB auth.
// Unlike order signing, the ClobAuth domain does NOT include a verifyingContract.
func computeClobAuthDomainSeparator(chainID int) [32]byte {
	var buf []byte
	buf = append(buf, clobAuthDomainTypeHash[:]...)
	buf = append(buf, clobAuthDomainName[:]...)
	buf = append(buf, clobAuthDomainVersion[:]...)
	buf = append(buf, abiEncodeUint256(big.NewInt(int64(chainID)))...)
	return keccak256(buf)
}

// clobAuthDigest computes the EIP-712 digest for a CLOB auth message.
func clobAuthDigest(address string, timestamp string, nonce int, chainID int) [32]byte {
	// Struct hash: keccak256(typeHash ‖ address ‖ keccak256(timestamp) ‖ nonce ‖ keccak256(message))
	tsHash := keccak256([]byte(timestamp))
	msgHash := keccak256([]byte(clobAuthMessageText))

	var buf []byte
	buf = append(buf, clobAuthTypeHash[:]...)
	buf = append(buf, abiEncodeAddress(address)...)
	buf = append(buf, tsHash[:]...)
	buf = append(buf, abiEncodeUint256(big.NewInt(int64(nonce)))...)
	buf = append(buf, msgHash[:]...)
	structHash := keccak256(buf)

	domainSep := computeClobAuthDomainSeparator(chainID)

	var dbuf []byte
	dbuf = append(dbuf, 0x19, 0x01) // EIP-712 prefix
	dbuf = append(dbuf, domainSep[:]...)
	dbuf = append(dbuf, structHash[:]...)
	return keccak256(dbuf)
}

// L1AuthHeaders contains the L1 authentication headers for CLOB API calls.
// These are used instead of L2 (HMAC) authentication for API key management.
type L1AuthHeaders struct {
	Address   string // POLY_ADDRESS: Ethereum address (0x-prefixed)
	Signature string // POLY_SIGNATURE: EIP-712 signature (0x-prefixed)
	Timestamp string // POLY_TIMESTAMP: Unix timestamp
	Nonce     string // POLY_NONCE: Nonce (default "0")
}

// BuildL1AuthHeaders creates L1 (EIP-712) authentication headers for the CLOB API.
// These headers are required for API key management endpoints (create, derive, delete).
func (s *OrderSigner) BuildL1AuthHeaders(timestamp string, nonce int) (*L1AuthHeaders, error) {
	digest := clobAuthDigest(s.address, timestamp, nonce, s.chain.ChainID)

	r, sig_s, v, err := ecdsaSignCompact(digest[:], s.privKey)
	if err != nil {
		return nil, fmt.Errorf("polymarket: signing clob auth: %w", err)
	}

	// Pack signature: r(32) + s(32) + v(1) where v = 27 or 28
	sigBytes := make([]byte, 65)
	copy(sigBytes[:32], r[:])
	copy(sigBytes[32:64], sig_s[:])
	sigBytes[64] = v + 27

	return &L1AuthHeaders{
		Address:   s.address,
		Signature: "0x" + hex.EncodeToString(sigBytes),
		Timestamp: timestamp,
		Nonce:     fmt.Sprintf("%d", nonce),
	}, nil
}

// calculateOrderAmounts computes the maker and taker amounts for an order.
//
// For BUY:  makerAmount = size·price (USDC paid), takerAmount = size (CT tokens received)
// For SELL: makerAmount = size (CT tokens sold), takerAmount = size·price (USDC received)
//
// Amounts are in base units with 6 decimal places (matching USDC and CT token precision).
// generateSalt creates a random salt for an order, matching the Python client's
// approach: round(unix_timestamp * random_float). This keeps the salt small enough
// (< 2^53) to be safely parsed by JavaScript JSON parsers on the server side.
func generateSalt() (*big.Int, error) {
	// Generate a random uint32 and multiply by a fractional value, similar
	// to Python's time.time() * random.random(). We use crypto/rand for entropy.
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	// Use the lower 53 bits to stay within JavaScript's safe integer range.
	v := new(big.Int).SetBytes(b)
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 53), big.NewInt(1))
	v.And(v, mask)
	return v, nil
}

func calculateOrderAmounts(price, size float64, side Side) (makerAmt, takerAmt *big.Int) {
	const scale = 1e6

	sizeScaled := int64(math.Round(size * scale))
	costScaled := int64(math.Round(size * price * scale))

	sizeInt := new(big.Int).SetInt64(sizeScaled)
	costInt := new(big.Int).SetInt64(costScaled)

	if side == Buy {
		return costInt, sizeInt
	}
	return sizeInt, costInt
}
