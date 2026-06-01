package main

import (
	"context"
	"fmt"
	"log"

	polymarket "github.com/caxqueiroz/polymarket-go"
)

func main() {
	client := polymarket.NewClient()
	ctx := context.Background()

	// Get a market from the CLOB to find token IDs.
	page, err := client.Clob.ListMarkets(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	if len(page.Data) == 0 {
		log.Fatal("no markets found")
	}

	market := page.Data[0]
	fmt.Printf("Market: %s\n\n", market.Question)

	for _, token := range market.Tokens {
		feeRate, err := client.Clob.GetFeeRateBPS(ctx, token.TokenID)
		if err != nil {
			log.Printf("  %s: error getting fee rate: %v", token.Outcome, err)
			continue
		}
		fmt.Printf("  %s (token: %s)\n", token.Outcome, token.TokenID)
		fmt.Printf("    Fee Rate: %.2f bps\n\n", feeRate)
	}
}
