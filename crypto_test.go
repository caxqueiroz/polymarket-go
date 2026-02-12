package polymarket

import (
	"encoding/hex"
	"math/big"
	"testing"
)

func TestKeccak256Empty(t *testing.T) {
	result := keccak256([]byte{})
	got := hex.EncodeToString(result[:])
	want := "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	if got != want {
		t.Fatalf("keccak256(\"\") = %s, want %s", got, want)
	}
}

func TestKeccak256Hello(t *testing.T) {
	result := keccak256([]byte("hello"))
	got := hex.EncodeToString(result[:])
	want := "1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8"
	if got != want {
		t.Fatalf("keccak256(\"hello\") = %s, want %s", got, want)
	}
}

func TestKeccak256MultiPart(t *testing.T) {
	// keccak256("helloworld") computed as two parts should equal single call.
	single := keccak256([]byte("helloworld"))
	multi := keccak256([]byte("hello"), []byte("world"))
	if single != multi {
		t.Fatalf("multi-part keccak256 mismatch: %x != %x", single, multi)
	}
}

func TestKeccak256NotSHA3(t *testing.T) {
	// Ethereum's Keccak-256 differs from NIST SHA3-256.
	// SHA3-256("") = a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a
	// Keccak-256("") = c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
	result := keccak256([]byte{})
	sha3Empty := "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"
	got := hex.EncodeToString(result[:])
	if got == sha3Empty {
		t.Fatal("keccak256 should differ from SHA3-256")
	}
}

func TestPrivateKeyToAddress(t *testing.T) {
	// Hardhat / Foundry account #0
	privKey, err := parsePrivateKey("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		t.Fatal(err)
	}
	got := privateKeyToAddress(privKey)
	want := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	if got != want {
		t.Fatalf("address = %s, want %s", got, want)
	}
}

func TestPrivateKeyToAddressWithPrefix(t *testing.T) {
	privKey, err := parsePrivateKey("0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		t.Fatal(err)
	}
	got := privateKeyToAddress(privKey)
	want := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	if got != want {
		t.Fatalf("address = %s, want %s", got, want)
	}
}

func TestPrivateKeyToAddressAccount1(t *testing.T) {
	// Hardhat account #1
	privKey, err := parsePrivateKey("59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d")
	if err != nil {
		t.Fatal(err)
	}
	got := privateKeyToAddress(privKey)
	want := "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
	if got != want {
		t.Fatalf("address = %s, want %s", got, want)
	}
}

func TestParsePrivateKeyInvalid(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"not hex", "zzzz"},
		{"zero", "0000000000000000000000000000000000000000000000000000000000000000"},
		{"too large", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364142"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsePrivateKey(tt.key)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestEcdsaSignAndRecover(t *testing.T) {
	privKey, _ := parsePrivateKey("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	expectedAddr := privateKeyToAddress(privKey)

	hash := keccak256([]byte("test message for signing"))

	r, s, v, err := ecdsaSignCompact(hash[:], privKey)
	if err != nil {
		t.Fatalf("ecdsaSignCompact error: %v", err)
	}

	if r == [32]byte{} {
		t.Fatal("r is zero")
	}
	if s == [32]byte{} {
		t.Fatal("s is zero")
	}
	if v > 1 {
		t.Fatalf("v = %d, want 0 or 1", v)
	}

	// Verify low-s
	sInt := new(big.Int).SetBytes(s[:])
	if sInt.Cmp(secp256k1HalfN) > 0 {
		t.Fatal("s > n/2 (low-s normalization failed)")
	}

	// Recover address from signature
	recovered, err := ecRecover(hash[:], r, s, v)
	if err != nil {
		t.Fatalf("ecRecover error: %v", err)
	}
	if recovered != expectedAddr {
		t.Fatalf("recovered address = %s, want %s", recovered, expectedAddr)
	}
}

func TestEcdsaSignDeterministicRecovery(t *testing.T) {
	// Sign multiple times and verify recovery works each time (non-deterministic k).
	privKey, _ := parsePrivateKey("59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d")
	expectedAddr := privateKeyToAddress(privKey)

	for i := 0; i < 5; i++ {
		hash := keccak256([]byte("message"), []byte{byte(i)})
		r, s, v, err := ecdsaSignCompact(hash[:], privKey)
		if err != nil {
			t.Fatalf("iteration %d: sign error: %v", i, err)
		}
		recovered, err := ecRecover(hash[:], r, s, v)
		if err != nil {
			t.Fatalf("iteration %d: recover error: %v", i, err)
		}
		if recovered != expectedAddr {
			t.Fatalf("iteration %d: recovered %s, want %s", i, recovered, expectedAddr)
		}
	}
}

func TestEcScalarBaseMult(t *testing.T) {
	// G * 1 = G
	one := big.NewInt(1)
	p := ecBaseMult(one)
	if p.x.Cmp(secp256k1Gx) != 0 || p.y.Cmp(secp256k1Gy) != 0 {
		t.Fatal("G * 1 != G")
	}

	// G * 0 = infinity
	zero := big.NewInt(0)
	p = ecBaseMult(zero)
	if !p.isInfinity() {
		t.Fatal("G * 0 != infinity")
	}

	// G * n = infinity (n is the curve order)
	p = ecBaseMult(secp256k1N)
	if !p.isInfinity() {
		t.Fatal("G * n != infinity")
	}
}

func TestEcDouble(t *testing.T) {
	g := ecPoint{new(big.Int).Set(secp256k1Gx), new(big.Int).Set(secp256k1Gy)}
	twoG := ecDouble(g)

	// 2G should be a valid point on the curve: y² = x³ + 7 mod p
	y2 := new(big.Int).Mul(twoG.y, twoG.y)
	y2.Mod(y2, secp256k1P)
	x3 := new(big.Int).Mul(twoG.x, twoG.x)
	x3.Mul(x3, twoG.x)
	x3.Add(x3, big.NewInt(7))
	x3.Mod(x3, secp256k1P)
	if y2.Cmp(x3) != 0 {
		t.Fatal("2G is not on curve")
	}
}
