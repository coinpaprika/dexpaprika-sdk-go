package dexpaprika

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTokens_GetDetails(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(3, 1*time.Second, 8*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test token details for a well-known token (WETH on Ethereum)
	tokenChain := "ethereum"
	// #nosec G101
	tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH

	// Get token details
	token, err := client.Tokens.GetDetails(ctx, tokenChain, tokenAddress)
	if err != nil {
		t.Fatalf("Tokens.GetDetails returned error: %v", err)
	}

	if token == nil {
		t.Fatal("Tokens.GetDetails returned nil, expected token details")
	}

	// Check token details
	if token.ID != tokenAddress {
		t.Errorf("Tokens.GetDetails returned wrong ID: got %s, want %s", token.ID, tokenAddress)
	}
	if token.Chain != tokenChain {
		t.Errorf("Tokens.GetDetails returned wrong chain: got %s, want %s", token.Chain, tokenChain)
	}
	if token.Symbol == "" {
		t.Error("Tokens.GetDetails returned empty symbol")
	}
	if token.Name == "" {
		t.Error("Tokens.GetDetails returned empty name")
	}
	if token.Decimals == 0 {
		t.Error("Tokens.GetDetails returned decimals = 0, expected > 0 for WETH")
	}
}

func TestTokens_GetPools(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(3, 1*time.Second, 8*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test token pools for a well-known token (WETH on Ethereum)
	tokenChain := "ethereum"
	// #nosec G101
	tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH

	// Get pools for the token
	opts := &TokenPoolsOptions{
		ListOptions: &ListOptions{
			Limit:   5,
			OrderBy: "volume_usd",
			Sort:    "desc",
		},
	}

	tokenPools, err := client.Tokens.GetPools(ctx, tokenChain, tokenAddress, opts)
	if err != nil {
		t.Fatalf("Tokens.GetPools returned error: %v", err)
	}

	if tokenPools == nil {
		t.Fatal("Tokens.GetPools returned nil, expected a PoolList")
	}

	// Check that we got some pools back
	if len(tokenPools.Pools) == 0 {
		t.Error("Tokens.GetPools returned empty list, expected some pools for WETH")
	}

	// All pools should contain the specified token
	for _, pool := range tokenPools.Pools {
		tokenFound := false
		for _, token := range pool.Tokens {
			if token.ID == tokenAddress {
				tokenFound = true
				break
			}
		}
		if !tokenFound {
			t.Errorf("Tokens.GetPools returned pool without the specified token: %s", pool.ID)
		}
	}
}

func TestTokens_GetPoolsCursorPagination(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(3, 1*time.Second, 8*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tokenChain := "ethereum"
	// #nosec G101
	tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH

	// The deprecated pair/reorder options must be ignored, not sent upstream.
	firstPage, err := client.Tokens.GetPools(ctx, tokenChain, tokenAddress, &TokenPoolsOptions{
		ListOptions:            &ListOptions{Limit: 2},
		AdditionalTokenAddress: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", // #nosec G101
		Reorder:                true,
	})
	if err != nil {
		t.Fatalf("Tokens.GetPools returned error: %v", err)
	}
	if len(firstPage.Pools) == 0 {
		t.Fatal("Tokens.GetPools returned empty first page, expected pools for WETH")
	}
	if !firstPage.HasNextPage || firstPage.NextCursor == nil || *firstPage.NextCursor == "" {
		t.Fatal("Tokens.GetPools first page has no next cursor, expected more WETH pools")
	}

	// Follow the cursor and make sure we get a different page.
	secondPage, err := client.Tokens.GetPools(ctx, tokenChain, tokenAddress, &TokenPoolsOptions{
		ListOptions: &ListOptions{Limit: 2, Cursor: *firstPage.NextCursor},
	})
	if err != nil {
		t.Fatalf("Tokens.GetPools (cursor) returned error: %v", err)
	}
	if len(secondPage.Pools) == 0 {
		t.Fatal("Tokens.GetPools (cursor) returned empty page")
	}
	if secondPage.Pools[0].ID == firstPage.Pools[0].ID {
		t.Errorf("Tokens.GetPools (cursor) returned the same first pool %s, expected the next page", secondPage.Pools[0].ID)
	}

	// Every returned pool must still contain the requested token.
	for _, pool := range append(firstPage.Pools, secondPage.Pools...) {
		tokenFound := false
		for _, token := range pool.Tokens {
			if token.ID == tokenAddress {
				tokenFound = true
				break
			}
		}
		if !tokenFound {
			t.Errorf("Tokens.GetPools returned pool without the specified token: %s", pool.ID)
		}
	}
}

func TestCachedClient_Tokens(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(3, 1*time.Second, 8*time.Second),
	)

	// Create a cached client
	cachedClient := NewCachedClient(client, nil, 5*time.Minute)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test token details caching
	tokenChain := "ethereum"
	// #nosec G101
	tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH

	// Get token details
	token, err := cachedClient.GetTokenDetails(ctx, tokenChain, tokenAddress)
	if err != nil {
		t.Fatalf("CachedClient.GetTokenDetails returned error: %v", err)
	}

	if token == nil {
		t.Fatal("CachedClient.GetTokenDetails returned nil, expected token details")
	}

	// Get token details again to test cache
	start := time.Now()
	tokenAgain, err := cachedClient.GetTokenDetails(ctx, tokenChain, tokenAddress)
	if err != nil {
		t.Fatalf("CachedClient.GetTokenDetails (again) returned error: %v", err)
	}
	duration := time.Since(start)

	// Cached response should be very fast
	if duration > 100*time.Millisecond {
		t.Logf("Warning: Cached response took longer than expected: %v", duration)
	}

	// Same token should be returned
	if tokenAgain.ID != token.ID || tokenAgain.Symbol != token.Symbol {
		t.Error("Cache inconsistency: token details changed between calls")
	}
}

func TestTokens_ValidationErrors(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Test GetDetails with empty network ID
	_, err := client.Tokens.GetDetails(ctx, "", "0xtoken")
	if err == nil || err.Error() != "network ID is required" {
		t.Errorf("Expected 'network ID is required' error, got: %v", err)
	}

	// Test GetDetails with empty token address
	_, err = client.Tokens.GetDetails(ctx, "ethereum", "")
	if err == nil || err.Error() != "token address is required" {
		t.Errorf("Expected 'token address is required' error, got: %v", err)
	}

	// Test GetPools with empty network ID
	_, err = client.Tokens.GetPools(ctx, "", "0xtoken", nil)
	if err == nil || err.Error() != "network ID is required" {
		t.Errorf("Expected 'network ID is required' error, got: %v", err)
	}

	// Test GetPools with empty token address
	_, err = client.Tokens.GetPools(ctx, "ethereum", "", nil)
	if err == nil || err.Error() != "token address is required" {
		t.Errorf("Expected 'token address is required' error, got: %v", err)
	}
}

// TestTokens_FilterSendsPriceChangeBounds covers the one price-change window
// tokens/search actually honours. I originally wrote all four off as pool-only,
// which was wrong: 24h works on both endpoints, only 6h, 1h and 5m are
// pool-only. Checked live, and the bound is real rather than ignored:
//
//	no filter                          [95.7, -0.07, 0.23, -0.02, 0.4]
//	price_change_percentage_24h_min=20 [95.7, 36.62, 15876.53, 22.96, 32.88]
//	misspelled name                    [95.7, -0.07, 0.23, -0.02, 0.4]
func TestTokens_FilterSendsPriceChangeBounds(t *testing.T) {
	var got string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"results":[],"has_next_page":false}`)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRetryConfig(0, 1*time.Millisecond, 1*time.Millisecond),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	min24h := 20.0
	max24h := -20.0
	_, err := client.Tokens.Filter(ctx, "ethereum", &TokenFilterOptions{
		PriceChange24hMin: &min24h,
		PriceChange24hMax: &max24h,
	})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}

	for _, want := range []string{
		"price_change_percentage_24h_min=20",
		"price_change_percentage_24h_max=-20",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("query %q is missing %q", got, want)
		}
	}
}
