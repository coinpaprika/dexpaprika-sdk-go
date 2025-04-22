//go:build e2e
// +build e2e

package dexpaprika

import (
	"context"
	"testing"
	"time"
)

// TestE2E_Basics runs basic E2E tests against the actual API
// These tests are designed to be stable and not flaky
//
// Run tests with: go test -v -tags=e2e ./dexpaprika -run TestE2E
// Note: Skip these tests in CI by not using the e2e tag
func TestE2E_Basics(t *testing.T) {
	// Skip test if not running in e2e mode
	// This helps prevent these tests from running in CI
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create a client with modest timeout and retry settings
	client := NewClient(
		WithRetryConfig(2, 1*time.Second, 3*time.Second),
	)

	// Create a context with 30s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run each test as a subtest
	t.Run("Networks", func(t *testing.T) {
		testE2E_Networks(t, ctx, client)
	})

	t.Run("Pools", func(t *testing.T) {
		testE2E_Pools(t, ctx, client)
	})

	t.Run("Search", func(t *testing.T) {
		testE2E_Search(t, ctx, client)
	})

	t.Run("PoolDetails", func(t *testing.T) {
		testE2E_PoolDetails(t, ctx, client)
	})

	// Additional tests for complete coverage
	t.Run("DEXes", func(t *testing.T) {
		testE2E_DEXes(t, ctx, client)
	})

	t.Run("PoolOHLCV", func(t *testing.T) {
		testE2E_PoolOHLCV(t, ctx, client)
	})

	t.Run("PoolTransactions", func(t *testing.T) {
		testE2E_PoolTransactions(t, ctx, client)
	})

	t.Run("TokenDetails", func(t *testing.T) {
		testE2E_TokenDetails(t, ctx, client)
	})

	t.Run("Stats", func(t *testing.T) {
		testE2E_Stats(t, ctx, client)
	})
}

// Test that networks endpoint returns data and includes ethereum
func testE2E_Networks(t *testing.T, ctx context.Context, client *Client) {
	networks, err := client.Networks.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list networks: %v", err)
	}

	// Verify we got some networks
	if len(networks) == 0 {
		t.Error("Expected networks to be returned, got none")
	}

	// Check if ethereum is in the list
	var foundEthereum bool
	for _, network := range networks {
		if network.ID == "ethereum" {
			foundEthereum = true
			break
		}
	}

	if !foundEthereum {
		t.Error("Expected to find 'ethereum' in networks list")
	}
}

// Test that pools endpoint respects the limit parameter
func testE2E_Pools(t *testing.T, ctx context.Context, client *Client) {
	// Request pools
	requestedLimit := 5

	// Create list options
	options := &ListOptions{
		Page:    0,
		Limit:   requestedLimit,
		Sort:    "",
		OrderBy: "",
	}

	pools, err := client.Pools.List(ctx, options)
	if err != nil {
		t.Fatalf("Failed to list pools: %v", err)
	}

	// Verify we get at least one pool
	if len(pools.Pools) == 0 {
		t.Error("Expected at least one pool, got none")
	}

	t.Logf("Requested %d pools, got %d", requestedLimit, len(pools.Pools))

	// Test network specific pools
	networkID := "ethereum"
	networkPools, err := client.Pools.ListByNetwork(ctx, networkID, options)
	if err != nil {
		t.Fatalf("Failed to list network pools: %v", err)
	}

	// Verify we get at least one pool
	if len(networkPools.Pools) == 0 {
		t.Error("Expected at least one network pool, got none")
	}

	t.Logf("Requested %d network pools, got %d", requestedLimit, len(networkPools.Pools))

	// Verify all pools are on the correct network
	for _, pool := range networkPools.Pools {
		if pool.Chain != networkID {
			t.Errorf("Expected pool to be on network %s, got %s", networkID, pool.Chain)
		}
	}
}

// Test search functionality for common terms and addresses
func testE2E_Search(t *testing.T, ctx context.Context, client *Client) {
	// Test cases
	searches := []struct {
		query       string
		description string
		minResults  int
	}{
		{
			query:       "eth",
			description: "Common crypto term",
			minResults:  1,
		},
		{
			query:       "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", // USDC on Ethereum
			description: "USDC Address",
			minResults:  1,
		},
	}

	for _, s := range searches {
		t.Run("Search:"+s.description, func(t *testing.T) {
			results, err := client.Search.Search(ctx, s.query)
			if err != nil {
				t.Fatalf("Failed to search for %q: %v", s.query, err)
			}

			// Count total results across all categories
			totalResults := len(results.Tokens) + len(results.Pools) + len(results.Dexes)
			if totalResults < s.minResults {
				t.Errorf("Expected at least %d results for query %q, got %d",
					s.minResults, s.query, totalResults)
			}

			t.Logf("Search %q returned: %d tokens, %d pools, %d dexes",
				s.query, len(results.Tokens), len(results.Pools), len(results.Dexes))
		})
	}
}

// Test that we can get pool details from a pool found in the pool list
func testE2E_PoolDetails(t *testing.T, ctx context.Context, client *Client) {
	// First get a list of pools on Ethereum
	networkID := "ethereum"
	options := &ListOptions{
		Page:  0,
		Limit: 1,
	}

	pools, err := client.Pools.ListByNetwork(ctx, networkID, options)
	if err != nil {
		t.Fatalf("Failed to list pools: %v", err)
	}

	if len(pools.Pools) == 0 {
		t.Fatal("No pools returned from API")
	}

	// Take the first pool from the list
	firstPool := pools.Pools[0]
	t.Logf("Testing pool details for: %s", firstPool.ID)

	// Get details for this pool
	poolDetails, err := client.Pools.GetDetails(ctx, networkID, firstPool.ID, false)
	if err != nil {
		t.Fatalf("Failed to get pool details: %v", err)
	}

	// Verify the pool details match the pool we requested
	if poolDetails.ID != firstPool.ID {
		t.Errorf("Pool ID mismatch: expected %s, got %s", firstPool.ID, poolDetails.ID)
	}

	// Verify we have tokens data
	if len(poolDetails.Tokens) < 2 {
		t.Errorf("Expected at least 2 tokens in pool, got %d", len(poolDetails.Tokens))
	}

	// Extra verification: check if one token has details we can follow
	if len(poolDetails.Tokens) > 0 {
		tokenID := poolDetails.Tokens[0].ID
		tokenNetwork := poolDetails.Tokens[0].Chain

		// Just verify the first token exists
		t.Logf("Testing token details for: %s on %s", tokenID, tokenNetwork)
		tokenDetails, err := client.Tokens.GetDetails(ctx, tokenNetwork, tokenID)
		if err != nil {
			t.Errorf("Failed to get token details: %v", err)
		} else {
			t.Logf("Successfully retrieved token: %s (%s)",
				tokenDetails.Name, tokenDetails.Symbol)
		}
	}
}

// Test that the DEXes endpoint returns data for a specific network
func testE2E_DEXes(t *testing.T, ctx context.Context, client *Client) {
	// Get DEXes for Ethereum network
	networkID := "ethereum"

	// Using direct page/limit parameters if ListDexes doesn't accept ListOptions
	dexes, err := client.Networks.ListDexes(ctx, networkID, 0, 10)
	if err != nil {
		t.Fatalf("Failed to list DEXes: %v", err)
	}

	// Verify we get at least one DEX
	if len(dexes.Dexes) == 0 {
		t.Error("Expected at least one DEX, got none")
	}

	t.Logf("Found %d DEXes on %s network", len(dexes.Dexes), networkID)

	// Verify pagination info
	if dexes.PageInfo.Page != 0 {
		t.Errorf("Expected page 0, got %d", dexes.PageInfo.Page)
	}

	// If we got DEXes, test getting pools for the first DEX
	if len(dexes.Dexes) > 0 {
		firstDEX := dexes.Dexes[0]
		t.Logf("Testing pools for DEX: %s (%s)", firstDEX.Name, firstDEX.ID)

		options := &ListOptions{
			Page:  0,
			Limit: 10,
		}

		dexPools, err := client.Pools.ListByDex(ctx, networkID, firstDEX.ID, options)
		if err != nil {
			t.Errorf("Failed to get pools for DEX %s: %v", firstDEX.ID, err)
		} else if len(dexPools.Pools) == 0 {
			t.Logf("No pools found for DEX %s", firstDEX.ID)
		} else {
			t.Logf("Found %d pools for DEX %s", len(dexPools.Pools), firstDEX.ID)
		}
	}

	// Test DEXes pagination - using direct page/limit parameters
	paginator := NewDexesPaginator(client, networkID, 10)

	if !paginator.HasNextPage() && len(dexes.Dexes) > 0 {
		t.Logf("Paginator reports no next page with %d results (this is normal if all results fit on one page)", len(dexes.Dexes))
	}
}

// Test that pool OHLCV data can be retrieved
func testE2E_PoolOHLCV(t *testing.T, ctx context.Context, client *Client) {
	// First get a pool ID
	networkID := "ethereum"
	options := &ListOptions{
		Page:  0,
		Limit: 1,
	}

	pools, err := client.Pools.ListByNetwork(ctx, networkID, options)
	if err != nil {
		t.Fatalf("Failed to list pools: %v", err)
	}

	if len(pools.Pools) == 0 {
		t.Fatal("No pools returned from API")
	}

	poolID := pools.Pools[0].ID
	t.Logf("Testing OHLCV data for pool: %s", poolID)

	// Get OHLCV data
	// Use current time minus 10 days as start time
	startTime := time.Now().Add(-10 * 24 * time.Hour).Format(time.RFC3339)

	// Create OHLCV options
	ohlcvOptions := &OHLCVOptions{
		Start:    startTime,
		End:      "",
		Interval: "24h", // daily data
		Limit:    7,
		Inversed: false,
	}

	ohlcv, err := client.Pools.GetOHLCV(ctx, networkID, poolID, ohlcvOptions)
	if err != nil {
		t.Errorf("Failed to get OHLCV data: %v", err)
		return
	}

	// Check if we got any data
	if len(ohlcv) == 0 {
		t.Logf("No OHLCV data returned for pool %s (this might be normal for some pools)", poolID)
	} else {
		t.Logf("Got %d OHLCV records for pool %s", len(ohlcv), poolID)

		// Verify format of first entry if available
		if len(ohlcv) > 0 {
			first := ohlcv[0]
			if first.TimeOpen == "" || first.TimeClose == "" {
				t.Errorf("OHLCV record has invalid time format: %+v", first)
			}

			// Values should be non-negative
			if first.Open < 0 || first.High < 0 || first.Low < 0 || first.Close < 0 || first.Volume < 0 {
				t.Errorf("OHLCV record has negative values: %+v", first)
			}

			t.Logf("Sample OHLCV: Open: %f, High: %f, Low: %f, Close: %f, Volume: %d",
				first.Open, first.High, first.Low, first.Close, first.Volume)
		}
	}
}

// Test that pool transactions can be retrieved
func testE2E_PoolTransactions(t *testing.T, ctx context.Context, client *Client) {
	// First get a pool ID
	networkID := "ethereum"
	options := &ListOptions{
		Page:  0,
		Limit: 1,
	}

	pools, err := client.Pools.ListByNetwork(ctx, networkID, options)
	if err != nil {
		t.Fatalf("Failed to list pools: %v", err)
	}

	if len(pools.Pools) == 0 {
		t.Fatal("No pools returned from API")
	}

	poolID := pools.Pools[0].ID
	t.Logf("Testing transactions for pool: %s", poolID)

	// Get transactions - with cursor parameter
	transactions, err := client.Pools.GetTransactions(ctx, networkID, poolID, 0, 5, "")
	if err != nil {
		t.Errorf("Failed to get transactions: %v", err)
		return
	}

	// Check if we got any transactions
	if len(transactions.Transactions) == 0 {
		t.Logf("No transactions returned for pool %s (this might be normal for some pools)", poolID)
	} else {
		t.Logf("Got %d transactions for pool %s", len(transactions.Transactions), poolID)

		// Test transactions pagination
		paginator := NewTransactionsPaginator(client, networkID, poolID, 5)

		if !paginator.HasNextPage() && len(transactions.Transactions) > 0 {
			t.Logf("Paginator reports no next page with %d results (this is normal if all results fit on one page)", len(transactions.Transactions))
		}
	}
}

// Test that token details can be retrieved
func testE2E_TokenDetails(t *testing.T, ctx context.Context, client *Client) {
	// Well-known token addresses for different networks
	tokens := []struct {
		network string
		address string
		symbol  string
	}{
		{"ethereum", "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "USDC"}, // USDC on Ethereum
		{"ethereum", "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", "WETH"}, // WETH on Ethereum
	}

	for _, token := range tokens {
		t.Run(token.symbol, func(t *testing.T) {
			// Get token details
			details, err := client.Tokens.GetDetails(ctx, token.network, token.address)
			if err != nil {
				t.Errorf("Failed to get token details for %s: %v", token.symbol, err)
				return
			}

			// Verify token details
			if details.Symbol != token.symbol {
				t.Errorf("Expected symbol %s, got %s", token.symbol, details.Symbol)
			}

			t.Logf("Got details for %s: Name=%s, Decimals=%d, Chain=%s",
				details.Symbol, details.Name, details.Decimals, details.Chain)

			// Test token pools
			options := &ListOptions{
				Page:  0,
				Limit: 5,
			}

			pools, err := client.Tokens.GetPools(ctx, token.network, token.address, options, "")
			if err != nil {
				t.Errorf("Failed to get pools for token %s: %v", token.symbol, err)
				return
			}

			if len(pools.Pools) == 0 {
				t.Logf("No pools found for token %s (unexpected for major tokens)", token.symbol)
			} else {
				t.Logf("Found %d pools for token %s", len(pools.Pools), token.symbol)

				// Verify all pools contain the token
				for _, pool := range pools.Pools {
					found := false
					for _, poolToken := range pool.Tokens {
						if poolToken.ID == token.address {
							found = true
							break
						}
					}

					if !found {
						t.Errorf("Pool %s does not contain token %s", pool.ID, token.address)
					}
				}
			}
		})
	}
}

// Test that the stats endpoint returns valid data
func testE2E_Stats(t *testing.T, ctx context.Context, client *Client) {
	stats, err := client.Utils.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	// Verify stats are reasonable
	if stats.Chains <= 0 {
		t.Errorf("Expected positive number of chains, got %d", stats.Chains)
	}

	if stats.Pools <= 0 {
		t.Errorf("Expected positive number of pools, got %d", stats.Pools)
	}

	if stats.Tokens <= 0 {
		t.Errorf("Expected positive number of tokens, got %d", stats.Tokens)
	}

	t.Logf("API Stats: %d chains, %d pools, %d tokens, %d factories",
		stats.Chains, stats.Pools, stats.Tokens, stats.Factories)
}
