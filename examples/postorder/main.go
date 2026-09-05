// Command postorder previews a V2 BUY order. Only -send submits it.
// Set POLY_PRIVATE_KEY locally; never paste a real key into source or chat.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	polymarket "github.com/caxqueiroz/polymarket-go"
)

func main() {
	if err := run(); err != nil {
		slog.Error("order example failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	tokenID := flag.String("token", "", "explicit outcome token ID (required)")
	price := flag.Float64("price", 0.01, "BUY limit price; must match the market tick size")
	size := flag.Float64("size", 5, "number of shares; check the market minimum")
	maker := flag.String("maker", "", "funding wallet address; required for deposit wallets")
	signatureType := flag.Int("signature-type", 0, "0=EOA, 1=proxy, 2=Safe, 3=deposit wallet")
	negRisk := flag.Bool("neg-risk", false, "use the negative-risk exchange")
	postOnly := flag.Bool("post-only", true, "reject an order that would immediately match")
	send := flag.Bool("send", false, "submit a real order (default: offline preview only)")
	flag.Parse()
	if *tokenID == "" {
		return fmt.Errorf("-token is required; select the market and outcome explicitly")
	}
	signer, err := polymarket.NewOrderSigner(os.Getenv("POLY_PRIVATE_KEY"), polymarket.Polygon)
	if err != nil {
		return fmt.Errorf("create signer from POLY_PRIVATE_KEY: %w", err)
	}
	signed, err := signer.CreateOrder(polymarket.CreateOrderParams{
		TokenID: *tokenID, Price: *price, Size: *size, Side: polymarket.Buy,
		Maker: *maker, SignatureType: polymarket.SignatureType(*signatureType), NegRisk: *negRisk,
		Builder: os.Getenv("POLY_BUILDER_CODE"),
	})
	if err != nil {
		return fmt.Errorf("create V2 order: %w", err)
	}
	fmt.Printf("BUY %g shares at %g; maker=%s; postOnly=%t\n", *size, *price, signed.Order.Maker, *postOnly)
	if !*send {
		fmt.Println("Offline preview only: no credentials derived and no order sent. Use -send only after checking the market, funding, approvals, and account eligibility.")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	apiCreds, err := polymarket.NewClient().Clob.CreateOrDeriveApiCreds(ctx, signer)
	if err != nil {
		return fmt.Errorf("derive API credentials: %w", err)
	}
	client := polymarket.NewClient(polymarket.WithCredentials(apiCreds.ToCredentials(signer.Address())))
	resp, err := client.Clob.PostOrderWithOptions(ctx, signed, polymarket.OrderGTC,
		polymarket.PostOrderOptions{PostOnly: *postOnly})
	if err != nil {
		return fmt.Errorf("post order: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("order rejected: %s", resp.ErrorMsg)
	}
	fmt.Printf("Order ID: %s; status: %s\n", resp.OrderID, resp.Status)
	return nil
}
