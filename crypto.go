package polymarket

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/bits"
	"strings"
)

// ---- Keccak-256 (pre-NIST, as used by Ethereum) ----
//
// Ethereum uses the original Keccak-256 (NOT NIST SHA-3-256). The only
// difference is the domain separation byte: Keccak uses 0x01 while SHA-3
// uses 0x06. Go's crypto/sha3 only provides the NIST variant, so we
// implement Keccak-256 here to maintain zero external dependencies.
//
// Parameters: Keccak-f[1600], rate=1088 bits (136 bytes), capacity=512 bits,
// output=256 bits.

const keccakRate = 136 // bytes

// Round constants for Keccak-f[1600].
var keccakRC = [24]uint64{
	0x0000000000000001, 0x0000000000008082, 0x800000000000808A, 0x8000000080008000,
	0x000000000000808B, 0x0000000080000001, 0x8000000080008081, 0x8000000000008009,
	0x000000000000008A, 0x0000000000000088, 0x0000000080008009, 0x000000008000000A,
	0x000000008000808B, 0x800000000000008B, 0x8000000000008089, 0x8000000000008003,
	0x8000000000008002, 0x8000000000000080, 0x000000000000800A, 0x800000008000000A,
	0x8000000080008081, 0x8000000000008080, 0x0000000080000001, 0x8000000080008008,
}

// Rotation offsets for ρ step, indexed as [x + 5*y].
var keccakRot = [25]uint{
	0, 1, 62, 28, 27, // y=0
	36, 44, 6, 55, 20, // y=1
	3, 10, 43, 25, 39, // y=2
	41, 45, 15, 21, 8, // y=3
	18, 2, 61, 56, 14, // y=4
}

// keccakF1600 applies the Keccak-f[1600] permutation (24 rounds).
func keccakF1600(state *[25]uint64) {
	for round := 0; round < 24; round++ {
		// θ (theta)
		var c [5]uint64
		for x := 0; x < 5; x++ {
			c[x] = state[x] ^ state[x+5] ^ state[x+10] ^ state[x+15] ^ state[x+20]
		}
		for x := 0; x < 5; x++ {
			d := c[(x+4)%5] ^ bits.RotateLeft64(c[(x+1)%5], 1)
			for y := 0; y < 25; y += 5 {
				state[y+x] ^= d
			}
		}

		// ρ (rho) and π (pi) combined
		var tmp [25]uint64
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				src := x + 5*y
				dst := y + 5*((2*x+3*y)%5)
				tmp[dst] = bits.RotateLeft64(state[src], int(keccakRot[src]))
			}
		}

		// χ (chi)
		for y := 0; y < 25; y += 5 {
			for x := 0; x < 5; x++ {
				state[y+x] = tmp[y+x] ^ (^tmp[y+(x+1)%5] & tmp[y+(x+2)%5])
			}
		}

		// ι (iota)
		state[0] ^= keccakRC[round]
	}
}

// keccak256 computes the Keccak-256 hash (as used by Ethereum).
func keccak256(data ...[]byte) [32]byte {
	var state [25]uint64
	var block [keccakRate]byte
	offset := 0

	for _, d := range data {
		for _, b := range d {
			block[offset] = b
			offset++
			if offset == keccakRate {
				for i := 0; i < keccakRate/8; i++ {
					state[i] ^= binary.LittleEndian.Uint64(block[i*8:])
				}
				keccakF1600(&state)
				offset = 0
			}
		}
	}

	// Keccak padding: 0x01 ... 0x80 (NOT SHA-3's 0x06 ... 0x80)
	for i := offset; i < keccakRate; i++ {
		block[i] = 0
	}
	block[offset] = 0x01
	block[keccakRate-1] |= 0x80

	for i := 0; i < keccakRate/8; i++ {
		state[i] ^= binary.LittleEndian.Uint64(block[i*8:])
	}
	keccakF1600(&state)

	var out [32]byte
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(out[i*8:], state[i])
	}
	return out
}

// ---- secp256k1 elliptic curve ----
//
// secp256k1 is the elliptic curve used by Ethereum for ECDSA signatures.
// Curve equation: y² = x³ + 7 over F_p.

var (
	secp256k1P, _  = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	secp256k1N, _  = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
	secp256k1HalfN = new(big.Int).Rsh(secp256k1N, 1)
	secp256k1Gx, _ = new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	secp256k1Gy, _ = new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)
)

// ecPoint represents a point on secp256k1 in affine coordinates.
// A nil x indicates the point at infinity.
type ecPoint struct {
	x, y *big.Int
}

func (p ecPoint) isInfinity() bool { return p.x == nil }

var ecInfinity = ecPoint{}

// ecAdd returns p1 + p2 on secp256k1.
func ecAdd(p1, p2 ecPoint) ecPoint {
	if p1.isInfinity() {
		return p2
	}
	if p2.isInfinity() {
		return p1
	}
	if p1.x.Cmp(p2.x) == 0 {
		if p1.y.Cmp(p2.y) == 0 {
			return ecDouble(p1)
		}
		return ecInfinity // p1 + (-p1) = O
	}

	// λ = (y₂ - y₁) · (x₂ - x₁)⁻¹ mod p
	dy := new(big.Int).Sub(p2.y, p1.y)
	dy.Mod(dy, secp256k1P)
	dx := new(big.Int).Sub(p2.x, p1.x)
	dx.Mod(dx, secp256k1P)
	lam := new(big.Int).Mul(dy, new(big.Int).ModInverse(dx, secp256k1P))
	lam.Mod(lam, secp256k1P)

	// x₃ = λ² - x₁ - x₂ mod p
	x3 := new(big.Int).Mul(lam, lam)
	x3.Sub(x3, p1.x)
	x3.Sub(x3, p2.x)
	x3.Mod(x3, secp256k1P)

	// y₃ = λ·(x₁ - x₃) - y₁ mod p
	y3 := new(big.Int).Sub(p1.x, x3)
	y3.Mul(lam, y3)
	y3.Sub(y3, p1.y)
	y3.Mod(y3, secp256k1P)

	return ecPoint{x3, y3}
}

// ecDouble returns 2·p on secp256k1.
func ecDouble(p ecPoint) ecPoint {
	if p.isInfinity() || p.y.Sign() == 0 {
		return ecInfinity
	}

	// λ = 3x² / (2y) mod p  (a = 0 for secp256k1)
	num := new(big.Int).Mul(p.x, p.x)
	num.Mod(num, secp256k1P)
	num.Mul(num, big.NewInt(3))
	num.Mod(num, secp256k1P)

	den := new(big.Int).Add(p.y, p.y)
	den.Mod(den, secp256k1P)
	lam := new(big.Int).Mul(num, new(big.Int).ModInverse(den, secp256k1P))
	lam.Mod(lam, secp256k1P)

	x3 := new(big.Int).Mul(lam, lam)
	x3.Sub(x3, new(big.Int).Add(p.x, p.x))
	x3.Mod(x3, secp256k1P)

	y3 := new(big.Int).Sub(p.x, x3)
	y3.Mul(lam, y3)
	y3.Sub(y3, p.y)
	y3.Mod(y3, secp256k1P)

	return ecPoint{x3, y3}
}

// ecScalarMult returns k·p using the double-and-add algorithm.
func ecScalarMult(p ecPoint, k *big.Int) ecPoint {
	if k.Sign() == 0 {
		return ecInfinity
	}
	result := ecInfinity
	for i := k.BitLen() - 1; i >= 0; i-- {
		result = ecDouble(result)
		if k.Bit(i) == 1 {
			result = ecAdd(result, p)
		}
	}
	return result
}

// ecBaseMult returns k·G where G is the secp256k1 generator.
func ecBaseMult(k *big.Int) ecPoint {
	g := ecPoint{new(big.Int).Set(secp256k1Gx), new(big.Int).Set(secp256k1Gy)}
	return ecScalarMult(g, k)
}

// ecdsaSignCompact signs a 32-byte hash with a secp256k1 private key.
// Returns r (32 bytes), s (32 bytes), and v (0 or 1 recovery parameter).
func ecdsaSignCompact(hash []byte, privKey *big.Int) (r, s [32]byte, v byte, err error) {
	for {
		// Random k ∈ [1, n-1]
		k, randErr := rand.Int(rand.Reader, new(big.Int).Sub(secp256k1N, big.NewInt(1)))
		if randErr != nil {
			return r, s, 0, fmt.Errorf("polymarket: generating k: %w", randErr)
		}
		k.Add(k, big.NewInt(1))

		// R = k·G
		R := ecBaseMult(k)

		// r = R.x mod n
		rInt := new(big.Int).Mod(R.x, secp256k1N)
		if rInt.Sign() == 0 {
			continue
		}

		// Recovery parameter: parity of R.y
		v = byte(R.y.Bit(0))

		// s = k⁻¹·(hash + r·privKey) mod n
		e := new(big.Int).SetBytes(hash)
		sInt := new(big.Int).Mul(rInt, privKey)
		sInt.Add(sInt, e)
		sInt.Mod(sInt, secp256k1N)
		kInv := new(big.Int).ModInverse(k, secp256k1N)
		sInt.Mul(sInt, kInv)
		sInt.Mod(sInt, secp256k1N)
		if sInt.Sign() == 0 {
			continue
		}

		// Low-s normalization (EIP-2)
		if sInt.Cmp(secp256k1HalfN) > 0 {
			sInt.Sub(secp256k1N, sInt)
			v ^= 1
		}

		rB := rInt.Bytes()
		sB := sInt.Bytes()
		copy(r[32-len(rB):], rB)
		copy(s[32-len(sB):], sB)

		return r, s, v, nil
	}
}

// ---- Address derivation ----

// pubKeyToAddress derives an Ethereum address from a public key point.
func pubKeyToAddress(pub ecPoint) string {
	xB := pub.x.Bytes()
	yB := pub.y.Bytes()
	pubBytes := make([]byte, 64)
	copy(pubBytes[32-len(xB):32], xB)
	copy(pubBytes[64-len(yB):64], yB)
	hash := keccak256(pubBytes)
	return toChecksumAddress(hex.EncodeToString(hash[12:]))
}

// toChecksumAddress returns an EIP-55 checksummed Ethereum address.
// Input: 40-char lowercase hex string (no 0x prefix).
// The EIP-55 algorithm: keccak256(lowercase_hex), then for each char,
// if the corresponding nibble in the hash >= 8, uppercase it.
func toChecksumAddress(addr string) string {
	hash := keccak256([]byte(addr))
	result := make([]byte, 42)
	result[0] = '0'
	result[1] = 'x'
	for i := 0; i < 40; i++ {
		c := addr[i]
		// hash nibble: byte = hash[i/2], nibble = high if i even, low if i odd
		hashByte := hash[i/2]
		var nibble byte
		if i%2 == 0 {
			nibble = hashByte >> 4
		} else {
			nibble = hashByte & 0x0f
		}
		if nibble >= 8 && c >= 'a' && c <= 'f' {
			result[i+2] = c - 32 // uppercase
		} else {
			result[i+2] = c
		}
	}
	return string(result)
}

// checksumAddress normalizes any Ethereum address string to EIP-55 checksum format.
func checksumAddress(addr string) string {
	addr = strings.TrimPrefix(addr, "0x")
	addr = strings.TrimPrefix(addr, "0X")
	addr = strings.ToLower(addr)
	return toChecksumAddress(addr)
}

// privateKeyToAddress derives an Ethereum address from a private key.
func privateKeyToAddress(privKey *big.Int) string {
	return pubKeyToAddress(ecBaseMult(privKey))
}

// parsePrivateKey parses a hex-encoded secp256k1 private key (with or without 0x prefix).
func parsePrivateKey(hexKey string) (*big.Int, error) {
	if len(hexKey) >= 2 && (hexKey[:2] == "0x" || hexKey[:2] == "0X") {
		hexKey = hexKey[2:]
	}
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("polymarket: invalid private key hex: %w", err)
	}
	key := new(big.Int).SetBytes(b)
	if key.Sign() == 0 || key.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("polymarket: private key out of valid range")
	}
	return key, nil
}

// ecRecover recovers the Ethereum address from an ECDSA signature.
func ecRecover(hash []byte, r, s [32]byte, v byte) (string, error) {
	rInt := new(big.Int).SetBytes(r[:])
	sInt := new(big.Int).SetBytes(s[:])

	// Compute R.y from R.x = r: y² = r³ + 7 mod p
	rCubed := new(big.Int).Mul(rInt, rInt)
	rCubed.Mod(rCubed, secp256k1P)
	rCubed.Mul(rCubed, rInt)
	rCubed.Mod(rCubed, secp256k1P)
	rhs := new(big.Int).Add(rCubed, big.NewInt(7))
	rhs.Mod(rhs, secp256k1P)

	// y = rhs^((p+1)/4) mod p  (valid because p ≡ 3 mod 4)
	exp := new(big.Int).Add(secp256k1P, big.NewInt(1))
	exp.Rsh(exp, 2)
	y := new(big.Int).Exp(rhs, exp, secp256k1P)

	// Verify y² = rhs
	y2 := new(big.Int).Mul(y, y)
	y2.Mod(y2, secp256k1P)
	if y2.Cmp(rhs) != 0 {
		return "", fmt.Errorf("polymarket: point not on curve")
	}

	// Choose parity matching v
	if y.Bit(0) != uint(v) {
		y.Sub(secp256k1P, y)
	}

	R := ecPoint{new(big.Int).Set(rInt), y}

	// Q = r⁻¹ · (s·R - e·G)
	e := new(big.Int).SetBytes(hash)
	rInv := new(big.Int).ModInverse(rInt, secp256k1N)

	sR := ecScalarMult(R, sInt)
	eG := ecBaseMult(e)
	negEG := ecPoint{new(big.Int).Set(eG.x), new(big.Int).Sub(secp256k1P, eG.y)}

	sum := ecAdd(sR, negEG)
	Q := ecScalarMult(sum, rInv)

	return pubKeyToAddress(Q), nil
}
