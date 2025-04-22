package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/coinpaprika/dexpaprika-sdk-go/dexpaprika"
)

func main() {
	fmt.Println("=== DexPaprika SDK Comprehensive Test ===")

	// Create a client with production settings
	client := dexpaprika.NewClient(
		dexpaprika.WithRetryConfig(2, 1*time.Second, 3*time.Second),
		dexpaprika.WithRateLimit(5.0),
	)

	// Create a cached client for testing caching functionality
	cachedClient := dexpaprika.NewCachedClient(client, nil, 5*time.Minute)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Test Networks endpoints
	fmt.Println("\n1. Testing Networks Endpoints:")
	testNetworks(ctx, client, cachedClient)

	// 2. Test Pools endpoints
	fmt.Println("\n2. Testing Pools Endpoints:")
	testPools(ctx, client)

	// 3. Test Tokens endpoints
	fmt.Println("\n3. Testing Tokens Endpoints:")
	testTokens(ctx, client)

	// 4. Test Search endpoint
	fmt.Println("\n4. Testing Search Endpoint:")
	testSearch(ctx, client)

	// 5. Test Utils endpoint
	fmt.Println("\n5. Testing Utils Endpoint:")
	testUtils(ctx, client)

	// 6. Test pagination
	fmt.Println("\n6. Testing Pagination:")
	testPagination(ctx, client)

	// 7. Test error handling
	fmt.Println("\n7. Testing Error Handling:")
	testErrorHandling(ctx, client)

	fmt.Println("\n=== Comprehensive Test Completed Successfully ===")
}

func testNetworks(ctx context.Context, client *dexpaprika.Client, cachedClient *dexpaprika.CachedClient) {
	// Test getting networks
	fmt.Println("   - Testing Networks.List")
	networks, err := client.Networks.List(ctx)
	if err != nil {
		log.Fatalf("Failed to get networks: %v", err)
	}
	fmt.Printf("   ✓ Got %d networks\n", len(networks))

	// Test cached networks retrieval
	fmt.Println("   - Testing cached Networks.List")
	cachedNetworks, err := cachedClient.GetNetworks(ctx)
	if err != nil {
		log.Fatalf("Failed to get cached networks: %v", err)
	}
	fmt.Printf("   ✓ Got %d networks from cache\n", len(cachedNetworks))

	if len(networks) > 0 {
		// Test getting DEXes for the first network
		networkID := networks[0].ID
		fmt.Printf("   - Testing Networks.ListDexes for %s\n", networkID)
		dexes, err := client.Networks.ListDexes(ctx, networkID, 0, 5)
		if err != nil {
			log.Fatalf("Failed to get dexes: %v", err)
		}
		fmt.Printf("   ✓ Got %d DEXes for %s\n", len(dexes.Dexes), networkID)
	}
}

func testPools(ctx context.Context, client *dexpaprika.Client) {
	// Test getting top pools
	fmt.Println("   - Testing Pools.List")
	poolsOpts := &dexpaprika.ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}
	pools, err := client.Pools.List(ctx, poolsOpts)
	if err != nil {
		log.Fatalf("Failed to get pools: %v", err)
	}
	fmt.Printf("   ✓ Got %d top pools\n", len(pools.Pools))

	// Test network-specific pools
	fmt.Println("   - Testing Pools.ListByNetwork")
	networkPools, err := client.Pools.ListByNetwork(ctx, "ethereum", poolsOpts)
	if err != nil {
		log.Fatalf("Failed to get Ethereum pools: %v", err)
	}
	fmt.Printf("   ✓ Got %d Ethereum pools\n", len(networkPools.Pools))

	// Test DEX-specific pools
	fmt.Println("   - Testing Pools.ListByDex")
	dexPools, err := client.Pools.ListByDex(ctx, "ethereum", "uniswap_v3", poolsOpts)
	if err != nil {
		log.Fatalf("Failed to get Uniswap V3 pools: %v", err)
	}
	fmt.Printf("   ✓ Got %d Uniswap V3 pools\n", len(dexPools.Pools))

	if len(pools.Pools) > 0 {
		pool := pools.Pools[0]

		// Test pool details
		fmt.Printf("   - Testing Pools.GetDetails for %s on %s\n", pool.ID, pool.Chain)
		details, err := client.Pools.GetDetails(ctx, pool.Chain, pool.ID, false)
		if err != nil {
			log.Fatalf("Failed to get pool details: %v", err)
		}
		fmt.Printf("   ✓ Got details for %s on %s\n", details.ID, details.Chain)

		// Test OHLCV data
		fmt.Printf("   - Testing Pools.GetOHLCV for %s on %s\n", pool.ID, pool.Chain)
		yesterday := time.Now().Add(-24 * time.Hour)
		ohlcvOpts := &dexpaprika.OHLCVOptions{
			Start:    yesterday.Format("2006-01-02"),
			End:      time.Now().Format("2006-01-02"),
			Interval: "1h",
			Limit:    3,
		}
		ohlcv, err := client.Pools.GetOHLCV(ctx, pool.Chain, pool.ID, ohlcvOpts)
		if err != nil {
			log.Fatalf("Failed to get OHLCV data: %v", err)
		}
		fmt.Printf("   ✓ Got %d OHLCV records\n", len(ohlcv))

		// Test transactions
		fmt.Printf("   - Testing Pools.GetTransactions for %s on %s\n", pool.ID, pool.Chain)
		transactions, err := client.Pools.GetTransactions(ctx, pool.Chain, pool.ID, 0, 5, "")
		if err != nil {
			log.Fatalf("Failed to get transactions: %v", err)
		}
		fmt.Printf("   ✓ Got %d transactions\n", len(transactions.Transactions))
	}
}

func testTokens(ctx context.Context, client *dexpaprika.Client) {
	// Test token details for a well-known token
	tokenChain := "ethereum"
	tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH

	fmt.Printf("   - Testing Tokens.GetDetails for %s on %s\n", tokenAddress, tokenChain)
	token, err := client.Tokens.GetDetails(ctx, tokenChain, tokenAddress)
	if err != nil {
		log.Fatalf("Failed to get token details: %v", err)
	}
	fmt.Printf("   ✓ Got details for %s (%s)\n", token.Name, token.Symbol)

	// Test token pools
	fmt.Printf("   - Testing Tokens.GetPools for %s on %s\n", tokenAddress, tokenChain)
	opts := &dexpaprika.ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}
	tokenPools, err := client.Tokens.GetPools(ctx, tokenChain, tokenAddress, opts, "")
	if err != nil {
		log.Fatalf("Failed to get token pools: %v", err)
	}
	fmt.Printf("   ✓ Got %d pools for %s\n", len(tokenPools.Pools), token.Symbol)

	// Test token pair pools
	pairTokenAddress := "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" // USDC
	fmt.Printf("   - Testing Tokens.GetPools for pair %s and %s\n", tokenAddress, pairTokenAddress)
	pairPools, err := client.Tokens.GetPools(ctx, tokenChain, tokenAddress, opts, pairTokenAddress)
	if err != nil {
		log.Fatalf("Failed to get token pair pools: %v", err)
	}
	fmt.Printf("   ✓ Got %d pools for the token pair\n", len(pairPools.Pools))
}

func testSearch(ctx context.Context, client *dexpaprika.Client) {
	// Test search
	query := "ethereum"
	fmt.Printf("   - Testing Search.Search for '%s'\n", query)
	results, err := client.Search.Search(ctx, query)
	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}
	fmt.Printf("   ✓ Search returned %d tokens, %d pools, %d DEXes\n",
		len(results.Tokens), len(results.Pools), len(results.Dexes))
}

func testUtils(ctx context.Context, client *dexpaprika.Client) {
	// Test getting stats
	fmt.Println("   - Testing Utils.GetStats")
	stats, err := client.Utils.GetStats(ctx)
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}
	fmt.Printf("   ✓ Got stats: %d chains, %d factories, %d pools, %d tokens\n",
		stats.Chains, stats.Factories, stats.Pools, stats.Tokens)
}

func testPagination(ctx context.Context, client *dexpaprika.Client) {
	// Test pools pagination
	fmt.Println("   - Testing PoolsPaginator")
	paginator := dexpaprika.NewPoolsPaginator(client, &dexpaprika.ListOptions{
		Limit:   10,
		OrderBy: "volume_usd",
		Sort:    "desc",
	})
	paginator.ForNetwork("ethereum")

	// Get first page
	fmt.Println("   - Getting first page of pools")
	if err := paginator.GetNextPage(ctx); err != nil {
		log.Fatalf("Failed to get first page: %v", err)
	}
	firstPage := paginator.GetCurrentPage()
	fmt.Printf("   ✓ First page has %d pools\n", len(firstPage))

	// Get second page if available
	if paginator.HasNextPage() {
		fmt.Println("   - Getting second page of pools")
		if err := paginator.GetNextPage(ctx); err != nil {
			log.Fatalf("Failed to get second page: %v", err)
		}
		secondPage := paginator.GetCurrentPage()
		fmt.Printf("   ✓ Second page has %d pools\n", len(secondPage))
	} else {
		fmt.Println("   ✓ No second page available (expected for small result sets)")
	}

	// Test DEX pagination
	fmt.Println("   - Testing DexesPaginator")
	dexPaginator := dexpaprika.NewDexesPaginator(client, "ethereum", 5)

	// Get first page
	if err := dexPaginator.GetNextPage(ctx); err != nil {
		log.Fatalf("Failed to get first page of DEXes: %v", err)
	}
	fmt.Printf("   ✓ First page has %d DEXes\n", len(dexPaginator.GetCurrentPage()))
}

func testErrorHandling(ctx context.Context, client *dexpaprika.Client) {
	// Test 404 error (resource not found)
	fmt.Println("   - Testing error handling with invalid resource")
	_, err := client.Tokens.GetDetails(ctx, "ethereum", "0xinvalid_address")
	if err != nil {
		var apiErr *dexpaprika.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("   ✓ Received API error as expected: %s (Status: %d)\n",
				apiErr.Message, apiErr.StatusCode)

			if errors.Is(err, dexpaprika.ErrNotFound) {
				fmt.Println("   ✓ Error correctly identified as 'not found'")
			} else {
				fmt.Printf("   ✗ Error not correctly identified: %v\n", err)
			}
		} else {
			fmt.Printf("   ✗ Unexpected error type: %T - %v\n", err, err)
		}
	} else {
		fmt.Println("   ✗ Expected to get an error but got success")
	}

	// Test with invalid parameters
	fmt.Println("   - Testing error handling with invalid parameters")
	_, err = client.Pools.GetOHLCV(ctx, "ethereum", "0xvalid_pool", &dexpaprika.OHLCVOptions{
		Interval: "invalid",
	})
	if err != nil {
		fmt.Printf("   ✓ Received error as expected: %v\n", err)
	} else {
		fmt.Println("   ✗ Expected to get an error but got success")
	}
}
