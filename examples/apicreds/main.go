package main

import (
	"context"
	"fmt"
	"log"
	"os"

	polymarket "github.com/caxqueiroz/polymarket-go"
)

func main() {
	// ── 1. Load private key ─────────────────────────────────────────────
	// Only the private key is needed — API credentials will be derived from it.
	privKey := os.Getenv("POLY_PRIVATE_KEY")
	if privKey == "" {
		log.Fatal("set POLY_PRIVATE_KEY environment variable (hex, with or without 0x)")
	}

	// ── 2. Create signer from private key ───────────────────────────────
	signer, err := polymarket.NewOrderSigner(privKey, polymarket.Polygon)
	if err != nil {
		log.Fatalf("NewOrderSigner: %v", err)
	}
	fmt.Printf("Wallet address: %s\n\n", signer.Address())

	// ── 3. Create an unauthenticated client ─────────────────────────────
	client := polymarket.NewClient()
	ctx := context.Background()

	// ── 4. Derive or create API credentials (L1 auth) ───────────────────
	// This signs an EIP-712 "ClobAuth" message with your private key and
	// sends it to the CLOB API. If credentials already exist for this
	// wallet, they are returned. Otherwise, new ones are created.
	//
	// Python equivalent:
	//   client = ClobClient(host, key=private_key, chain_id=137)
	//   api_creds = client.create_or_derive_api_creds()
	fmt.Println("Deriving API credentials...")
	creds, err := client.Clob.CreateOrDeriveApiCreds(ctx, signer)
	if err != nil {
		log.Fatalf("CreateOrDeriveApiCreds: %v", err)
	}

	fmt.Println("=== API Credentials ===")
	fmt.Printf("  API Key:    %s\n", creds.ApiKey)
	fmt.Printf("  Secret:     %s\n", creds.Secret)
	fmt.Printf("  Passphrase: %s\n\n", creds.Passphrase)

	// ── 5. Create authenticated client with the derived credentials ─────
	// Convert ApiCredentials → Credentials for HMAC signing.
	authedClient := polymarket.NewClient(
		polymarket.WithCredentials(creds.ToCredentials(signer.Address())),
	)

	// ── 6. Test the credentials by making an authenticated request ──────
	fmt.Println("Testing credentials with authenticated requests...")

	// 6a. Get balance/allowance (asset_type is required: "CONDITIONAL" or "COLLATERAL")
	assetType := "COLLATERAL"
	sigType := 0 // 0=EOA, 1=POLY_PROXY, 2=POLY_GNOSIS_SAFE
	ba, err := authedClient.Clob.GetBalanceAllowance(ctx, &polymarket.BalanceParams{
		AssetType:     &assetType,
		SignatureType: &sigType,
	})
	if err != nil {
		fmt.Printf("  GetBalanceAllowance: %v\n\n", err)
	} else {
		fmt.Printf("  Balance:   %s\n", ba.Balance)
		fmt.Printf("  Allowance: %s\n\n", ba.Allowance)
	}

	// 6b. Get open orders
	orders, err := authedClient.Clob.GetOrders(ctx, nil)
	if err != nil {
		fmt.Printf("  GetOrders: %v\n\n", err)
	} else {
		fmt.Printf("  Open orders: %d\n\n", len(orders))
		for i, o := range orders {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(orders)-5)
				break
			}
			fmt.Printf("    %d. %s %s @ %s (status: %s)\n", i+1, o.Side, o.OriginalSize, o.Price, o.Status)
		}
	}

	// 6c. List existing API keys for this wallet (requires L2 auth)
	fmt.Println("Listing API keys...")
	keys, err := authedClient.Clob.GetApiKeys(ctx)
	if err != nil {
		fmt.Printf("  GetApiKeys: %v\n", err)
	} else {
		fmt.Printf("  API keys for this wallet: %d\n", len(keys))
		for i, k := range keys {
			fmt.Printf("    %d. %s\n", i+1, k)
		}
	}
}
