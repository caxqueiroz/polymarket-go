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

	// List active markets from the Gamma API.
	active := true
	limit := 5
	markets, err := client.Gamma.ListMarkets(ctx, &polymarket.GammaMarketParams{
		Active: &active,
		Limit:  &limit,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d active markets:\n\n", len(markets))
	for _, m := range markets {
		fmt.Printf("  %s\n", m.Question)
		fmt.Printf("    Slug:       %s\n", m.Slug)
		fmt.Printf("    Volume:     $%.2f\n", m.Volume)
		fmt.Printf("    Liquidity:  $%.2f\n", m.Liquidity)
		fmt.Printf("    Outcomes:   %v\n", []string(m.Outcomes))
		fmt.Printf("    Prices:     %v\n", []string(m.OutcomePrices))
		fmt.Println()
	}
}
