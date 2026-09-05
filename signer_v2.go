package polymarket

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const zeroBytes32 = "0x0000000000000000000000000000000000000000000000000000000000000000"

func validAddress(address string) bool {
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return false
	}
	_, err := hex.DecodeString(address[2:])
	return err == nil
}

func normalizeBytes32(value string) (string, error) {
	if value == "" {
		return zeroBytes32, nil
	}
	if len(value) != 66 || !strings.HasPrefix(value, "0x") {
		return "", fmt.Errorf("expected 0x-prefixed bytes32")
	}
	if _, err := hex.DecodeString(value[2:]); err != nil {
		return "", fmt.Errorf("invalid bytes32: %w", err)
	}
	return strings.ToLower(value), nil
}

// Only called after validation (or with valid typed-data fixtures).
func abiEncodeBytes32(value string) []byte {
	if value == "" {
		return make([]byte, 32)
	}
	b, _ := hex.DecodeString(strings.TrimPrefix(value, "0x"))
	return b
}

func validateV2Order(o *OrderData) error {
	if o == nil {
		return fmt.Errorf("polymarket: nil order")
	}
	for _, field := range []struct {
		name     string
		value    *big.Int
		positive bool
	}{
		{"salt", o.Salt, false}, {"tokenId", o.TokenID, false},
		{"makerAmount", o.MakerAmount, true}, {"takerAmount", o.TakerAmount, true},
		{"timestamp", o.Timestamp, true},
	} {
		if field.value == nil || field.value.Sign() < 0 || field.value.BitLen() > 256 || (field.positive && field.value.Sign() == 0) {
			return fmt.Errorf("polymarket: invalid %s", field.name)
		}
	}
	if !validAddress(o.Maker) || !validAddress(o.Signer) || strings.EqualFold(o.Maker, zeroAddress) || strings.EqualFold(o.Signer, zeroAddress) {
		return fmt.Errorf("polymarket: invalid maker or signer address")
	}
	if o.Side != Buy && o.Side != Sell {
		return fmt.Errorf("polymarket: invalid order side")
	}
	if o.SignatureType < SignatureTypeEOA || o.SignatureType > SignatureTypePoly1271 {
		return fmt.Errorf("polymarket: unsupported signature type %d", o.SignatureType)
	}
	if o.SignatureType == SignatureTypePoly1271 && !strings.EqualFold(o.Maker, o.Signer) {
		return fmt.Errorf("polymarket: POLY_1271 signer must be the deposit-wallet maker")
	}
	if (o.Nonce != nil && o.Nonce.Sign() != 0) || (o.FeeRateBPS != nil && o.FeeRateBPS.Sign() != 0) || (o.Taker != "" && !strings.EqualFold(o.Taker, zeroAddress)) {
		return fmt.Errorf("polymarket: feeRateBps, nonce and taker are not supported in V2")
	}
	if o.Expiration != nil && (o.Expiration.Sign() < 0 || o.Expiration.BitLen() > 256) {
		return fmt.Errorf("polymarket: invalid expiration")
	}
	if _, err := normalizeBytes32(o.Metadata); err != nil {
		return fmt.Errorf("polymarket: metadata: %w", err)
	}
	if _, err := normalizeBytes32(o.Builder); err != nil {
		return fmt.Errorf("polymarket: builder: %w", err)
	}
	return nil
}

// depositWalletDigest implements the POLY_1271 TypedDataSign envelope used by
// Polymarket deposit wallets. The application domain is the V2 Exchange domain;
// the nested struct binds the order to the DepositWallet contract and chain.
func depositWalletDigest(o *OrderData, chain ChainConfig) [32]byte {
	typeHash := keccak256([]byte("TypedDataSign(Order contents,string name,string version,uint256 chainId,address verifyingContract,bytes32 salt)" + orderTypeString))
	contents := hashOrder(o)
	name := keccak256([]byte("DepositWallet"))
	version := keccak256([]byte("1"))
	buf := append([]byte{}, typeHash[:]...)
	buf = append(buf, contents[:]...)
	buf = append(buf, name[:]...)
	buf = append(buf, version[:]...)
	buf = append(buf, abiEncodeUint256(big.NewInt(int64(chain.ChainID)))...)
	buf = append(buf, abiEncodeAddress(o.Signer)...)
	buf = append(buf, make([]byte, 32)...)
	structHash := keccak256(buf)
	exchange := chain.Exchange
	if o.NegRisk {
		exchange = chain.NegRiskExchange
	}
	domain := computeDomainSeparator(chain.ChainID, exchange)
	digest := append([]byte{0x19, 0x01}, domain[:]...)
	digest = append(digest, structHash[:]...)
	return keccak256(digest)
}
