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

	// Get a market to find a token ID.
	page, err := client.Clob.ListMarkets(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	if len(page.Data) == 0 {
		log.Fatal("no markets found")
	}

	market := page.Data[0]
	if len(market.Tokens) == 0 {
		log.Fatal("market has no tokens")
	}

	token := market.Tokens[0]
	fmt.Printf("Order Book for: %s (%s)\n\n", market.Question, token.Outcome)

	book, err := client.Clob.GetBook(ctx, token.TokenID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("  Tick Size: %s\n", book.TickSize)
	fmt.Printf("  Neg Risk:  %v\n\n", book.NegRisk)

	maxLevels := 10

	fmt.Println("  BIDS")
	fmt.Println("  Price      Size")
	fmt.Println("  -----      ----")
	for i, level := range book.Bids {
		if i >= maxLevels {
			fmt.Printf("  ... (%d more levels)\n", len(book.Bids)-maxLevels)
			break
		}
		fmt.Printf("  %-10s %s\n", level.Price, level.Size)
	}

	fmt.Println()
	fmt.Println("  ASKS")
	fmt.Println("  Price      Size")
	fmt.Println("  -----      ----")
	for i, level := range book.Asks {
		if i >= maxLevels {
			fmt.Printf("  ... (%d more levels)\n", len(book.Asks)-maxLevels)
			break
		}
		fmt.Printf("  %-10s %s\n", level.Price, level.Size)
	}
}
