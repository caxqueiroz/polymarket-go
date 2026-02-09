package main

import (
	"context"
	"fmt"
	"log"

	polymarket "polymarket-go"
)

func main() {
	client := polymarket.NewClient()
	ctx := context.Background()

	// First, get a market from the CLOB to find a token ID.
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
		// Get price for each outcome.
		buyPrice, err := client.Clob.GetPrice(ctx, token.TokenID, polymarket.Buy)
		if err != nil {
			log.Printf("  %s: error getting buy price: %v", token.Outcome, err)
			continue
		}

		mid, err := client.Clob.GetMidpoint(ctx, token.TokenID)
		if err != nil {
			log.Printf("  %s: error getting midpoint: %v", token.Outcome, err)
			continue
		}

		spread, err := client.Clob.GetSpread(ctx, token.TokenID)
		if err != nil {
			log.Printf("  %s: error getting spread: %v", token.Outcome, err)
			continue
		}

		fmt.Printf("  %s:\n", token.Outcome)
		fmt.Printf("    Buy Price:  %s\n", buyPrice)
		fmt.Printf("    Midpoint:   %s\n", mid)
		fmt.Printf("    Spread:     %s\n", spread)
	}
}
