package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/coinpaprika/dexpaprika-sdk-go/dexpaprika"
)

func main() {
	// Create a new client
	client := dexpaprika.NewClient()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test 1: Get networks
	fmt.Println("Testing Networks API...")
	networks, err := client.Networks.List(ctx)
	if err != nil {
		log.Fatalf("Failed to get networks: %v", err)
	}
	fmt.Printf("Successfully fetched %d networks\n", len(networks))

	// Print the first network if available
	if len(networks) > 0 {
		fmt.Printf("First network: %s (%s)\n", networks[0].DisplayName, networks[0].ID)
	}

	// Test 2: Get top pools
	fmt.Println("\nTesting Pools API...")
	poolsOpts := &dexpaprika.ListOptions{
		Limit: 5,
	}
	pools, err := client.Pools.List(ctx, poolsOpts)
	if err != nil {
		log.Fatalf("Failed to get pools: %v", err)
	}
	fmt.Printf("Successfully fetched %d pools\n", len(pools.Pools))

	// Print the first pool if available
	if len(pools.Pools) > 0 {
		firstPool := pools.Pools[0]
		fmt.Printf("First pool: %s on %s (Volume: $%.2f)\n",
			firstPool.DexName,
			firstPool.Chain,
			firstPool.VolumeUSD)
	}

	// Test 3: Get pools on a specific network
	networkID := "ethereum"
	fmt.Printf("\nTesting Pools on Network API (%s)...\n", networkID)
	networkPools, err := client.Pools.ListByNetwork(ctx, networkID, poolsOpts)
	if err != nil {
		log.Fatalf("Failed to get pools on network: %v", err)
	}
	fmt.Printf("Successfully fetched %d pools on %s\n", len(networkPools.Pools), networkID)

	// Test 4: Get DEXes on a network
	fmt.Printf("\nTesting DEXes on Network API (%s)...\n", networkID)
	dexes, err := client.Networks.ListDexes(ctx, networkID, 0, 5)
	if err != nil {
		log.Fatalf("Failed to get DEXes on network: %v", err)
	}
	fmt.Printf("Successfully fetched %d DEXes on %s\n", len(dexes.Dexes), networkID)

	// Test 5: If we have a pool, get its details
	if len(pools.Pools) > 0 {
		poolID := pools.Pools[0].ID
		poolChain := pools.Pools[0].Chain
		fmt.Printf("\nTesting Pool Details API (Pool %s on %s)...\n", poolID, poolChain)

		poolDetails, err := client.Pools.GetDetails(ctx, poolChain, poolID, false)
		if err != nil {
			log.Printf("Warning: Failed to get pool details: %v", err)
		} else {
			fmt.Printf("Successfully fetched pool details - Fee: %.2f%%\n", poolDetails.Fee*100)
		}
	}

	// Test 6: Search for a token
	fmt.Println("\nTesting Search API...")
	searchResults, err := client.Search.Search(ctx, "bitcoin")
	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}

	fmt.Printf("Search results - Tokens: %d, Pools: %d, Dexes: %d\n",
		len(searchResults.Tokens),
		len(searchResults.Pools),
		len(searchResults.Dexes))

	// If we didn't encounter any errors, the SDK is working properly
	fmt.Println("\n✅ The DexPaprika SDK appears to be working properly!")
	os.Exit(0)
}
