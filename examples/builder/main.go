// Command builder demonstrates order attribution using Builder API credentials.
//
// Required environment variables:
//
//	POLY_PRIVATE_KEY          - hex-encoded Ethereum private key
//	POLY_BUILDER_API_KEY      - Builder API key (UUID)
//	POLY_BUILDER_API_SECRET   - Base64url-encoded HMAC secret
//	POLY_BUILDER_PASSPHRASE   - Builder passphrase
//
// Usage:
//
//	POLY_PRIVATE_KEY=0x... POLY_BUILDER_API_KEY=... POLY_BUILDER_API_SECRET=... POLY_BUILDER_PASSPHRASE=... go run ./examples/builder
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	polymarket "polymarket-go"
)

func main() {
	privKey := os.Getenv("POLY_PRIVATE_KEY")
	if privKey == "" {
		log.Fatal("POLY_PRIVATE_KEY is required")
	}

	builderKey := os.Getenv("POLY_BUILDER_API_KEY")
	builderSecret := os.Getenv("POLY_BUILDER_API_SECRET")
	builderPassphrase := os.Getenv("POLY_BUILDER_PASSPHRASE")

	signer, err := polymarket.NewOrderSigner(privKey, polymarket.Polygon)
	if err != nil {
		log.Fatalf("creating signer: %v", err)
	}

	// Derive API credentials.
	ctx := context.Background()
	tmpClient := polymarket.NewClient()
	apiCreds, err := tmpClient.Clob.CreateOrDeriveApiCreds(ctx, signer)
	if err != nil {
		log.Fatalf("deriving API credentials: %v", err)
	}

	// Build client options.
	opts := []polymarket.Option{
		polymarket.WithCredentials(apiCreds.ToCredentials(signer.Address())),
	}

	// Add builder credentials if configured.
	if builderKey != "" {
		fmt.Println("Builder credentials configured — orders will include attribution headers")
		opts = append(opts, polymarket.WithBuilderCredentials(&polymarket.BuilderCredentials{
			Key:        builderKey,
			Secret:     builderSecret,
			Passphrase: builderPassphrase,
		}))
	} else {
		fmt.Println("No builder credentials — orders will be anonymous")
	}

	client := polymarket.NewClient(opts...)

	// Query builder trades (if builder credentials are configured).
	if builderKey != "" {
		trades, err := client.Clob.GetBuilderTrades(ctx, nil)
		if err != nil {
			log.Fatalf("getting builder trades: %v", err)
		}
		fmt.Printf("Builder trades: %d\n", len(trades))
		for _, t := range trades {
			fmt.Printf("  %s: %s %s @ %s (%s)\n", t.ID, t.Side, t.Size, t.Price, t.Status)
		}
	}
}
