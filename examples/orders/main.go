package main

import (
	"context"
	"fmt"
	"log"
	"os"

	polymarket "polymarket-go"
)

func main() {
	// Credentials from environment variables.
	creds := &polymarket.Credentials{
		Key:        os.Getenv("POLY_API_KEY"),
		Secret:     os.Getenv("POLY_API_SECRET"),
		Passphrase: os.Getenv("POLY_PASSPHRASE"),
		Address:    os.Getenv("POLY_ADDRESS"),
	}
	if creds.Key == "" || creds.Secret == "" {
		log.Fatal("set POLY_API_KEY, POLY_API_SECRET, POLY_PASSPHRASE, and POLY_ADDRESS")
	}

	client := polymarket.NewClient(
		polymarket.WithCredentials(creds),
	)
	ctx := context.Background()

	// List open orders.
	orders, err := client.Clob.GetOrders(ctx, nil)
	if err != nil {
		log.Fatalf("GetOrders: %v", err)
	}
	fmt.Printf("Open orders: %d\n\n", len(orders))
	for _, o := range orders {
		fmt.Printf("  %s  %s %s @ %s  (status: %s)\n",
			o.ID, o.Side, o.OriginalSize, o.Price, o.Status)
	}

	// Check balance and allowance.
	ba, err := client.Clob.GetBalanceAllowance(ctx, nil)
	if err != nil {
		log.Fatalf("GetBalanceAllowance: %v", err)
	}
	fmt.Printf("\nBalance:   %s\n", ba.Balance)
	fmt.Printf("Allowance: %s\n", ba.Allowance)

	// List recent trades.
	trades, err := client.Clob.GetTrades(ctx, nil)
	if err != nil {
		log.Fatalf("GetTrades: %v", err)
	}
	fmt.Printf("\nRecent trades: %d\n", len(trades))
	for _, t := range trades {
		fmt.Printf("  %s  %s %s @ %s\n", t.ID, t.Side, t.Size, t.Price)
	}
}
