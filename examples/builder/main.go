// Command builder reads the public V2 builder trade feed.
//
// Required environment variables:
//
//	POLY_BUILDER_CODE - public 0x-prefixed bytes32 builder code
//
// Usage:
//
//	POLY_BUILDER_CODE=0x... go run ./examples/builder
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	polymarket "github.com/caxqueiroz/polymarket-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	trades, err := polymarket.NewClient().Clob.GetBuilderTradesByCode(ctx, os.Getenv("POLY_BUILDER_CODE"), nil)
	if err != nil {
		slog.Error("getting builder trades", "error", err)
		os.Exit(1)
	}
	fmt.Printf("Builder trades: %d\n", len(trades))
	for _, trade := range trades {
		fmt.Printf("  %s: %s %s @ %s (%s), builder fee %s\n",
			trade.ID, trade.Side, trade.Size, trade.Price, trade.Status, trade.BuilderFee)
	}
}
