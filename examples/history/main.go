package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	polymarket "github.com/caxqueiroz/polymarket-go"
)

func main() {
	// ── 1. Load private key ─────────────────────────────────────────────
	privKey := os.Getenv("POLY_PRIVATE_KEY")
	if privKey == "" {
		log.Fatal("set POLY_PRIVATE_KEY environment variable (hex, with or without 0x)")
	}

	// ── 2. Create signer + derive API credentials ───────────────────────
	signer, err := polymarket.NewOrderSigner(privKey, polymarket.Polygon)
	if err != nil {
		log.Fatalf("NewOrderSigner: %v", err)
	}
	fmt.Printf("Wallet: %s\n\n", signer.Address())

	client := polymarket.NewClient()
	ctx := context.Background()

	fmt.Println("Deriving API credentials...")
	apiCreds, err := client.Clob.CreateOrDeriveApiCreds(ctx, signer)
	if err != nil {
		log.Fatalf("CreateOrDeriveApiCreds: %v", err)
	}
	fmt.Printf("  API Key: %s\n\n", apiCreds.ApiKey)

	// Re-create client with HMAC credentials for authenticated endpoints.
	client = polymarket.NewClient(
		polymarket.WithCredentials(apiCreds.ToCredentials(signer.Address())),
	)

	// ── 3. Orders (CLOB, authenticated) ─────────────────────────────────
	fmt.Println("=== Orders ===")
	orders, err := client.Clob.GetOrders(ctx, nil)
	if err != nil {
		fmt.Printf("  GetOrders error: %v\n\n", err)
	} else {
		fmt.Printf("  Total: %d\n", len(orders))
		for i, o := range orders {
			if i >= 10 {
				fmt.Printf("  ... and %d more\n", len(orders)-10)
				break
			}
			fmt.Printf("  %d. [%s] %s %s @ %s  (market: %s)\n",
				i+1, o.Status, o.Side, o.OriginalSize, o.Price, o.Market)
		}
		fmt.Println()
	}

	// ── 4. Trades (CLOB, authenticated) ─────────────────────────────────
	fmt.Println("=== Trades (CLOB) ===")
	trades, err := client.Clob.GetTrades(ctx, nil)
	if err != nil {
		fmt.Printf("  GetTrades error: %v\n\n", err)
	} else {
		fmt.Printf("  Total: %d\n", len(trades))
		for i, t := range trades {
			if i >= 10 {
				fmt.Printf("  ... and %d more\n", len(trades)-10)
				break
			}
			fmt.Printf("  %d. [%s] %s %s @ %s\n", i+1, t.ID, t.Side, t.Size, t.Price)
			fmt.Printf("     Outcome:    %s\n", t.Outcome)
			fmt.Printf("     Status:     %s\n", t.Status)
			fmt.Printf("     Market:     %s\n", t.Market)
			fmt.Printf("     AssetID:    %s\n", t.AssetID)
			fmt.Printf("     Fee:        %s bps\n", t.FeeRateBPS)
			fmt.Printf("     Owner:      %s\n", t.Owner)
			fmt.Printf("     Maker:      %s\n", t.MakerAddress)
			fmt.Printf("     Trader:     %s\n", t.TradeOwner)
			fmt.Printf("     Type:       %s\n", t.Type)
			fmt.Printf("     MatchTime:  %s\n", t.MatchTime)
			fmt.Printf("     CreatedAt:  %s\n", t.CreatedAt)
			if t.TransactionHash != "" {
				fmt.Printf("     TxHash:     %s\n", t.TransactionHash)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// ── 5. Trades (Data API, public by user address) ────────────────────
	fmt.Println("=== Trades (Data API) ===")
	limit := 10
	dataTrades, err := client.Data.ListTrades(ctx, &polymarket.TradesParams{
		User:  signer.Address(),
		Limit: &limit,
	})
	if err != nil {
		fmt.Printf("  ListTrades error: %v\n\n", err)
	} else {
		fmt.Printf("  Showing up to %d trades\n", limit)
		for i, t := range dataTrades {
			ts := time.Unix(t.Timestamp, 0).Format("2006-01-02 15:04")
			fmt.Printf("  %d. [%s] %s %.4f @ $%.4f  %s — %s\n",
				i+1, ts, t.Side, t.Size, t.Price, t.Outcome, t.Title)
		}
		fmt.Println()
	}

	// ── 6. Positions (Data API, public by user address) ─────────────────
	fmt.Println("=== Positions ===")
	positions, err := client.Data.ListPositions(ctx, &polymarket.PositionsParams{
		User:  signer.Address(),
		Limit: &limit,
	})
	if err != nil {
		fmt.Printf("  ListPositions error: %v\n\n", err)
	} else {
		fmt.Printf("  Showing up to %d positions\n", len(positions))
		for i, p := range positions {
			fmt.Printf("  %d. %s — %s\n", i+1, p.Title, p.Outcome)
			fmt.Printf("     Size: %.4f  AvgPrice: $%.4f  CurPrice: $%.4f\n",
				p.Size, p.AvgPrice, p.CurPrice)
			fmt.Printf("     PnL: $%.4f (%.2f%%)\n", p.CashPnl, p.PercentPnl*100)
		}
		fmt.Println()
	}

	// ── 7. Activity (Data API, public by user address) ──────────────────
	fmt.Println("=== Activity ===")
	activity, err := client.Data.ListActivity(ctx, &polymarket.ActivityParams{
		User:  signer.Address(),
		Limit: &limit,
	})
	if err != nil {
		fmt.Printf("  ListActivity error: %v\n\n", err)
	} else {
		fmt.Printf("  Showing up to %d activities\n", len(activity))
		for i, a := range activity {
			ts := time.Unix(a.Timestamp, 0).Format("2006-01-02 15:04")
			fmt.Printf("  %d. [%s] %-8s %s %.4f @ $%.4f  %s\n",
				i+1, ts, a.Type, a.Side, a.Size, a.Price, a.Title)
		}
		fmt.Println()
	}

	// ── 8. Price history (CLOB, public) ─────────────────────────────────
	// Find an active market to show price history for.
	fmt.Println("=== Price History (sample market) ===")
	active := true
	gammaLimit := 1
	gammaMarkets, err := client.Gamma.ListMarkets(ctx, &polymarket.GammaMarketParams{
		Active: &active,
		Limit:  &gammaLimit,
	})
	if err != nil || len(gammaMarkets) == 0 {
		fmt.Printf("  Could not find an active market for price history\n")
	} else {
		gm := gammaMarkets[0]
		if len(gm.ClobTokenIDs) > 0 {
			tokenID := string(gm.ClobTokenIDs[0])
			now := time.Now().Unix()
			weekAgo := now - 7*24*3600
			history, err := client.Clob.GetPriceHistory(ctx, polymarket.PriceHistoryParams{
				Market:   tokenID,
				StartTs:  &weekAgo,
				EndTs:    &now,
				Interval: polymarket.Interval1h,
			})
			if err != nil {
				fmt.Printf("  GetPriceHistory error: %v\n", err)
			} else {
				fmt.Printf("  Market: %s\n", gm.Question)
				fmt.Printf("  Token:  %s\n", tokenID)
				fmt.Printf("  Points: %d (last 7 days, 1h interval)\n", len(history.History))
				// Show first and last few points
				for i, pt := range history.History {
					if i >= 5 && i < len(history.History)-3 {
						if i == 5 {
							fmt.Printf("  ...\n")
						}
						continue
					}
					ts := time.Unix(pt.T, 0).Format("2006-01-02 15:04")
					fmt.Printf("  [%s] $%.4f\n", ts, pt.P)
				}
			}
		}
	}
}
