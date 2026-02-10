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

	query := "Bitcoin"
	fmt.Printf("Searching for: %q\n\n", query)

	results, err := client.Gamma.Search(ctx, query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total results: %d (hasMore: %v)\n\n", results.Pagination.TotalResults, results.Pagination.HasMore)

	// Print events
	if len(results.Events) > 0 {
		fmt.Printf("=== Events (%d) ===\n", len(results.Events))
		for i, e := range results.Events {
			fmt.Printf("  %d. %s\n", i+1, e.Title)
			fmt.Printf("     Slug:   %s\n", e.Slug)
			fmt.Printf("     Volume: $%.2f\n", e.Volume)
			if len(e.Markets) > 0 {
				fmt.Printf("     Markets (%d):\n", len(e.Markets))
				for _, m := range e.Markets {
					fmt.Printf("       - %s (Volume: $%.2f)\n", m.Question, m.Volume)
				}
			}
			fmt.Println()
		}
	}

	// Print tags
	if len(results.Tags) > 0 {
		fmt.Printf("=== Tags (%d) ===\n", len(results.Tags))
		for _, t := range results.Tags {
			fmt.Printf("  - %s (slug: %s, events: %d)\n", t.Label, t.Slug, t.EventCount)
		}
		fmt.Println()
	}

	// Print profiles
	if len(results.Profiles) > 0 {
		fmt.Printf("=== Profiles (%d) ===\n", len(results.Profiles))
		for _, p := range results.Profiles {
			fmt.Printf("  - %s (%s)\n", p.Name, p.Pseudonym)
		}
		fmt.Println()
	}
}
