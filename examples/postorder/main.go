package main

import (
	"context"
	"fmt"
	"log"
	"os"

	polymarket "polymarket-go"
)

func main() {
	// ── 1. Load private key ─────────────────────────────────────────────
	privKey := os.Getenv("POLY_PRIVATE_KEY")
	if privKey == "" {
		log.Fatal("set POLY_PRIVATE_KEY environment variable (hex, with or without 0x)")
	}

	// ── 2. Create signer + unauthenticated client ───────────────────────
	signer, err := polymarket.NewOrderSigner(privKey, polymarket.Polygon)
	if err != nil {
		log.Fatalf("NewOrderSigner: %v", err)
	}
	fmt.Printf("Signer address: %s\n\n", signer.Address())

	client := polymarket.NewClient()
	ctx := context.Background()

	// ── 3. Derive API credentials (L1 auth) ─────────────────────────────
	fmt.Println("Deriving API credentials...")
	apiCreds, err := client.Clob.CreateOrDeriveApiCreds(ctx, signer)
	if err != nil {
		log.Fatalf("CreateOrDeriveApiCreds: %v", err)
	}
	fmt.Printf("  API Key: %s\n\n", apiCreds.ApiKey)

	// ── 4. Re-create client with HMAC credentials ───────────────────────
	client = polymarket.NewClient(
		polymarket.WithCredentials(apiCreds.ToCredentials(signer.Address())),
	)

	// ── 5. Find an active market via Gamma API ──────────────────────────
	// The Gamma API supports filtering by active/closed, which is much
	// faster than paginating through the CLOB's /markets endpoint.
	fmt.Println("Looking for an active market via Gamma API...")
	active := true
	closed := false
	limit := 10
	gammaMarkets, err := client.Gamma.ListMarkets(ctx, &polymarket.GammaMarketParams{
		Active: &active,
		Closed: &closed,
		Limit:  &limit,
	})
	if err != nil {
		log.Fatalf("Gamma.ListMarkets: %v", err)
	}

	var tokenID string
	var question string
	var negRisk bool
	found := false
	for _, gm := range gammaMarkets {
		if gm.Active && !gm.Closed && len(gm.ClobTokenIDs) > 0 {
			tokenID = string(gm.ClobTokenIDs[0])
			question = gm.Question
			negRisk = gm.NegRisk
			found = true
			break
		}
	}
	if !found {
		log.Fatal("no active market accepting orders found via Gamma API")
	}

	fmt.Printf("Market:  %s\n", question)
	fmt.Printf("Token:   %s\n", tokenID)
	fmt.Printf("NegRisk: %v\n\n", negRisk)

	// ── 6. Get fee rate for this token ──────────────────────────────────
	feeRate, err := client.Clob.GetFeeRateBPS(ctx, tokenID)
	if err != nil {
		log.Fatalf("GetFeeRateBPS: %v", err)
	}
	fmt.Printf("Fee rate: %.2f bps\n\n", feeRate)

	// ── 7. Build and sign order ─────────────────────────────────────────
	// Place a small BUY order at a very low price so it sits on the book
	// and is unlikely to fill immediately. Adjust as needed.
	params := polymarket.CreateOrderParams{
		TokenID:    tokenID,
		Price:      0.01, // $0.01 per share (very unlikely to fill)
		Size:       5,    // 5 shares
		Side:       polymarket.Buy,
		FeeRateBPS: int(feeRate),
		NegRisk:    negRisk,
	}

	signed, err := signer.CreateOrder(params)
	if err != nil {
		log.Fatalf("CreateOrder: %v", err)
	}

	fmt.Println("=== Signed Order ===")
	fmt.Printf("  Salt:        %s\n", signed.Order.Salt)
	fmt.Printf("  Maker:       %s\n", signed.Order.Maker)
	fmt.Printf("  Signer:      %s\n", signed.Order.Signer)
	fmt.Printf("  TokenID:     %s\n", signed.Order.TokenID)
	fmt.Printf("  MakerAmount: %s\n", signed.Order.MakerAmount)
	fmt.Printf("  TakerAmount: %s\n", signed.Order.TakerAmount)
	fmt.Printf("  Side:        %s\n", signed.Order.Side)
	fmt.Printf("  Signature:   %s\n\n", signed.Signature)

	// ── 8. Post the order ───────────────────────────────────────────────
	resp, err := client.Clob.PostOrder(ctx, signed, polymarket.OrderGTC)
	if err != nil {
		log.Fatalf("PostOrder: %v", err)
	}

	fmt.Println("=== Post Order Response ===")
	fmt.Printf("  Success:  %v\n", resp.Success)
	fmt.Printf("  OrderID:  %s\n", resp.OrderID)
	if resp.ErrorMsg != "" {
		fmt.Printf("  Error:    %s\n", resp.ErrorMsg)
	}
	if len(resp.TransactionHashes) > 0 {
		fmt.Printf("  Tx Hashes: %v\n", resp.TransactionHashes)
	}
}
