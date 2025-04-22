package dexpaprika

import (
	"context"
	"testing"
	"time"
)

func TestTokens_GetDetails(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test token details for a well-known token (WETH on Ethereum)
	tokenChain := "ethereum"
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
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test token pools for a well-known token (WETH on Ethereum)
	tokenChain := "ethereum"
	tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH

	// Get pools for the token
	opts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}

	tokenPools, err := client.Tokens.GetPools(ctx, tokenChain, tokenAddress, opts, "")
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

func TestTokens_GetPoolsWithPair(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test token pair pools (WETH-USDC on Ethereum)
	tokenChain := "ethereum"
	tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"     // WETH
	pairTokenAddress := "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" // USDC

	// Get pools for the token pair
	opts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}

	pairPools, err := client.Tokens.GetPools(ctx, tokenChain, tokenAddress, opts, pairTokenAddress)
	if err != nil {
		t.Fatalf("Tokens.GetPools (with pair) returned error: %v", err)
	}

	if pairPools == nil {
		t.Fatal("Tokens.GetPools (with pair) returned nil, expected a PoolList")
	}

	// All pools should contain both tokens
	for _, pool := range pairPools.Pools {
		firstTokenFound := false
		secondTokenFound := false

		for _, token := range pool.Tokens {
			if token.ID == tokenAddress {
				firstTokenFound = true
			}
			if token.ID == pairTokenAddress {
				secondTokenFound = true
			}
		}

		if !firstTokenFound || !secondTokenFound {
			t.Errorf("Tokens.GetPools (with pair) returned pool without both tokens: %s", pool.ID)
		}
	}
}

func TestCachedClient_Tokens(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a cached client
	cachedClient := NewCachedClient(client, nil, 5*time.Minute)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test token details caching
	tokenChain := "ethereum"
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
