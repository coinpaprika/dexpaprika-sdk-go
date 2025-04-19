package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/coinpaprika/dexpaprika-sdk-go/dexpaprika"
)

func main() {
	// Create a client with production settings
	client := dexpaprika.NewClient(
		// Add retry with exponential backoff
		dexpaprika.WithRetryConfig(3, 1*time.Second, 5*time.Second),
		// Add rate limiting - 2 requests per second
		dexpaprika.WithRateLimit(2.0),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a cached client for better performance
	cachedClient := dexpaprika.NewCachedClient(client, nil, 5*time.Minute)

	fmt.Println("=== DexPaprika SDK Production Demo ===")

	// Example 1: Get Networks with Caching
	fmt.Println("\n1. Networks (cached):")
	networks, err := cachedClient.GetNetworks(ctx)
	if err != nil {
		handleError("Failed to get networks", err)
		os.Exit(1)
	}
	fmt.Printf("   Found %d networks\n", len(networks))

	// Example 2: Get a specific network's DEXes with pagination
	if len(networks) > 0 {
		networkID := networks[0].ID
		fmt.Printf("\n2. DEXes on %s (with pagination):\n", networkID)

		// Create a DEXes paginator
		dexesPaginator := dexpaprika.NewDexesPaginator(client, networkID, 5)
		dexCount := 0
		pageNum := 0

		// Process all pages
		for dexesPaginator.HasNextPage() {
			pageNum++
			fmt.Printf("   Page %d:\n", pageNum)

			if err := dexesPaginator.GetNextPage(ctx); err != nil {
				handleError("Failed to get DEXes page", err)
				break
			}

			dexes := dexesPaginator.GetCurrentPage()
			for _, dex := range dexes {
				fmt.Printf("   - %s (%s protocol)\n", dex.Name, dex.Protocol)
				dexCount++

				// Just show a few DEXes for the demo
				if dexCount >= 5 {
					break
				}
			}

			// For the demo, stop after one page
			if true {
				break
			}
		}
	}

	// Example 3: Get top pools with error handling
	fmt.Println("\n3. Top pools (with error handling):")
	poolsOpts := &dexpaprika.ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}

	pools, err := client.Pools.List(ctx, poolsOpts)
	if err != nil {
		var apiErr *dexpaprika.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("   API Error: %s (Status Code: %d)\n",
				apiErr.Message,
				apiErr.StatusCode)

			// Check for specific error conditions
			if errors.Is(err, dexpaprika.ErrRateLimit) {
				fmt.Println("   Rate limit exceeded, try again later")
			}
		} else {
			fmt.Printf("   Other error: %v\n", err)
		}
	} else {
		fmt.Printf("   Found %d pools\n", len(pools.Pools))

		// Display first pool
		if len(pools.Pools) > 0 {
			pool := pools.Pools[0]
			fmt.Printf("   Top pool: %s on %s (Volume: $%.2f)\n",
				pool.DexName,
				pool.Chain,
				pool.VolumeUSD)
		}
	}

	// Example 4: Get token details with caching
	fmt.Println("\n4. Token details (cached):")
	// Use a well-known token like WETH on Ethereum
	tokenChain := "ethereum"
	tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // #nosec WETH

	tokenDetails, err := cachedClient.GetTokenDetails(ctx, tokenChain, tokenAddress)
	if err != nil {
		handleError("Failed to get token details", err)
	} else {
		fmt.Printf("   Token: %s (%s)\n", tokenDetails.Name, tokenDetails.Symbol)
		fmt.Printf("   Chain: %s\n", tokenDetails.Chain)
		fmt.Printf("   Decimals: %d\n", tokenDetails.Decimals)
		if tokenDetails.Summary != nil && tokenDetails.Summary.PriceUSD > 0 {
			fmt.Printf("   Price: $%.2f\n", tokenDetails.Summary.PriceUSD)
		}
	}

	// Example 5: Search functionality
	fmt.Println("\n5. Search API:")
	searchResults, err := client.Search.Search(ctx, "ethereum")
	if err != nil {
		handleError("Failed to search", err)
	} else {
		fmt.Printf("   Found: %d tokens, %d pools, %d DEXes\n",
			len(searchResults.Tokens),
			len(searchResults.Pools),
			len(searchResults.Dexes))
	}

	// Example 6: Get global stats
	fmt.Println("\n6. Global stats:")
	stats, err := cachedClient.GetStats(ctx)
	if err != nil {
		handleError("Failed to get stats", err)
	} else {
		fmt.Printf("   Chains: %d\n", stats.Chains)
		fmt.Printf("   Factories: %d\n", stats.Factories)
		fmt.Printf("   Pools: %d\n", stats.Pools)
		fmt.Printf("   Tokens: %d\n", stats.Tokens)
	}

	// Example 7: Pool details + OHLCV
	fmt.Println("\n7. Pool details and OHLCV data:")
	if len(pools.Pools) > 0 {
		poolChain := pools.Pools[0].Chain
		poolID := pools.Pools[0].ID

		fmt.Printf("   Getting details for pool %s on %s\n", poolID, poolChain)

		poolDetails, err := client.Pools.GetDetails(ctx, poolChain, poolID, false)
		if err != nil {
			handleError("Failed to get pool details", err)
		} else {
			fmt.Printf("   Pool: %s on %s\n", poolDetails.DexName, poolDetails.Chain)
			fmt.Printf("   Fee: %.2f%%\n", poolDetails.Fee*100)
			fmt.Printf("   Last price: $%.4f\n", poolDetails.LastPriceUSD)

			// Try to get OHLCV data
			fmt.Println("   Getting OHLCV data...")
			now := time.Now()
			yesterday := now.Add(-24 * time.Hour)
			ohlcvOpts := &dexpaprika.OHLCVOptions{
				Start:    yesterday.Format("2006-01-02"),
				End:      now.Format("2006-01-02"),
				Interval: "1h",
				Limit:    3, // Just a few for demo
			}

			ohlcv, err := client.Pools.GetOHLCV(ctx, poolChain, poolID, ohlcvOpts)
			if err != nil {
				handleError("Failed to get OHLCV data", err)
			} else {
				fmt.Printf("   Found %d OHLCV records\n", len(ohlcv))

				// Display the first few records
				for i, record := range ohlcv {
					if i >= 3 {
						break
					}
					fmt.Printf("   - %s: Open=$%.4f, Close=$%.4f, Volume=%d\n",
						record.TimeOpen,
						record.Open,
						record.Close,
						record.Volume)
				}
			}
		}
	}

	fmt.Println("\nDemo completed successfully!")
}

func handleError(message string, err error) {
	var apiErr *dexpaprika.APIError
	if errors.As(err, &apiErr) {
		log.Printf("%s: API Error (%d): %s", message, apiErr.StatusCode, apiErr.Message)

		// Check for specific error types
		switch {
		case errors.Is(err, dexpaprika.ErrRateLimit):
			log.Println("Rate limit exceeded")
		case errors.Is(err, dexpaprika.ErrNotFound):
			log.Println("Resource not found")
		case errors.Is(err, dexpaprika.ErrUnauthorized):
			log.Println("Unauthorized request")
		}
	} else {
		log.Printf("%s: %v", message, err)
	}
}
