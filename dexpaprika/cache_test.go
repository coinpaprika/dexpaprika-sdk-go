package dexpaprika

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryCache(t *testing.T) {
	cache := NewInMemoryCache()

	// Test setting and getting an item
	key := "test-key"
	value := "test-value"
	ttl := 1 * time.Second

	// Set the item
	cache.Set(key, value, ttl)

	// Get the item
	got, found := cache.Get(key)
	if !found {
		t.Error("Get() found = false, want true")
	}

	if got != value {
		t.Errorf("Get() got = %v, want %v", got, value)
	}

	// Test expiration
	time.Sleep(ttl + 100*time.Millisecond)
	_, found = cache.Get(key)
	if found {
		t.Error("Get() after expiration found = true, want false")
	}

	// Test delete
	cache.Set(key, value, 10*time.Minute)
	cache.Delete(key)
	_, found = cache.Get(key)
	if found {
		t.Error("Get() after Delete() found = true, want false")
	}

	// Test clear
	cache.Set("key1", "value1", 10*time.Minute)
	cache.Set("key2", "value2", 10*time.Minute)
	cache.Clear()
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	if found1 || found2 {
		t.Error("Get() after Clear() found items, want none")
	}
}

func TestCachedClient(t *testing.T) {
	// Create a standard client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a cached client with a short TTL for testing
	cachedClient := NewCachedClient(client, nil, 500*time.Millisecond)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test GetNetworks caching
	firstNetworks, err := cachedClient.GetNetworks(ctx)
	if err != nil {
		t.Fatalf("GetNetworks() first call error = %v", err)
	}

	startTime := time.Now()
	secondNetworks, err := cachedClient.GetNetworks(ctx)
	duration := time.Since(startTime)
	if err != nil {
		t.Fatalf("GetNetworks() second call error = %v", err)
	}

	// The second call should be very fast since it's cached
	if duration > 50*time.Millisecond {
		t.Logf("Warning: Cached call took %v, expected < 50ms", duration)
	}

	// Check that we got the same data
	if len(firstNetworks) != len(secondNetworks) {
		t.Errorf("GetNetworks() returned different data sizes: %d vs %d",
			len(firstNetworks), len(secondNetworks))
	}

	// Wait for cache to expire
	time.Sleep(600 * time.Millisecond)

	// This call should fetch from the API again
	startTime = time.Now()
	_, err = cachedClient.GetNetworks(ctx)
	duration = time.Since(startTime)
	if err != nil {
		t.Fatalf("GetNetworks() after expiration error = %v", err)
	}

	// This call should take longer since cache expired
	if duration < 20*time.Millisecond {
		t.Logf("Warning: API call after cache expiration took %v, expected > 20ms", duration)
	}
}

func TestCachedClient_WithCustomCache(t *testing.T) {
	// Create a standard client
	client := NewClient()

	// Create a custom cache
	cache := NewInMemoryCache()

	// Create a cached client with the custom cache
	cachedClient := NewCachedClient(client, cache, 10*time.Minute)

	// Ensure the custom cache is used
	if cachedClient.cache != cache {
		t.Error("NewCachedClient() did not use provided cache")
	}
}

func TestCachedClient_GetTokenDetails(t *testing.T) {
	// Create a standard client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a cached client with a short TTL for testing
	cachedClient := NewCachedClient(client, nil, 500*time.Millisecond)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test token details for a well-known token (WETH on Ethereum)
	tokenChain := "ethereum"
	// #nosec G101
	tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH

	// First call should hit the API
	token1, err := cachedClient.GetTokenDetails(ctx, tokenChain, tokenAddress)
	if err != nil {
		t.Fatalf("GetTokenDetails() first call error = %v", err)
	}

	if token1 == nil {
		t.Fatal("GetTokenDetails() first call returned nil")
	}

	// Second call should hit the cache
	startTime := time.Now()
	token2, err := cachedClient.GetTokenDetails(ctx, tokenChain, tokenAddress)
	duration := time.Since(startTime)
	if err != nil {
		t.Fatalf("GetTokenDetails() second call error = %v", err)
	}

	// Should be very fast from cache
	if duration > 50*time.Millisecond {
		t.Logf("Warning: Cached GetTokenDetails call took %v, expected < 50ms", duration)
	}

	// Data should be the same
	if token1.ID != token2.ID || token1.Symbol != token2.Symbol {
		t.Errorf("GetTokenDetails() returned different data: %v vs %v", token1, token2)
	}
}

func TestCachedClient_GetPoolDetails(t *testing.T) {
	// Create a standard client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a cached client
	cachedClient := NewCachedClient(client, nil, 500*time.Millisecond)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get a pool to test with
	pools, err := client.Pools.ListByNetwork(ctx, "ethereum", &ListOptions{Limit: 1})
	if err != nil || len(pools.Pools) == 0 {
		t.Skip("Could not get a pool to test caching")
	}

	poolID := pools.Pools[0].ID
	networkID := "ethereum"

	// First call should hit the API
	pool1, err := cachedClient.GetPoolDetails(ctx, networkID, poolID, false)
	if err != nil {
		t.Fatalf("GetPoolDetails() first call error = %v", err)
	}

	// Second call should hit the cache
	startTime := time.Now()
	pool2, err := cachedClient.GetPoolDetails(ctx, networkID, poolID, false)
	duration := time.Since(startTime)
	if err != nil {
		t.Fatalf("GetPoolDetails() second call error = %v", err)
	}

	// Should be very fast from cache
	if duration > 50*time.Millisecond {
		t.Logf("Warning: Cached GetPoolDetails call took %v, expected < 50ms", duration)
	}

	// Data should be the same
	if pool1.ID != pool2.ID {
		t.Errorf("GetPoolDetails() returned different IDs: %v vs %v", pool1.ID, pool2.ID)
	}
}

func TestCachedClient_GetDexes(t *testing.T) {
	// Create a standard client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a cached client with a short TTL for testing
	cachedClient := NewCachedClient(client, nil, 500*time.Millisecond)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test for a specific network
	networkID := "ethereum"
	page := 0
	limit := 5

	// First call should hit the API
	dexes1, err := cachedClient.GetDexes(ctx, networkID, page, limit)
	if err != nil {
		t.Fatalf("GetDexes() first call error = %v", err)
	}

	if dexes1 == nil || dexes1.Dexes == nil {
		t.Fatal("GetDexes() first call returned nil")
	}

	// Second call should hit the cache
	startTime := time.Now()
	dexes2, err := cachedClient.GetDexes(ctx, networkID, page, limit)
	duration := time.Since(startTime)
	if err != nil {
		t.Fatalf("GetDexes() second call error = %v", err)
	}

	// Should be very fast from cache
	if duration > 50*time.Millisecond {
		t.Logf("Warning: Cached GetDexes call took %v, expected < 50ms", duration)
	}

	// Data should be the same
	if len(dexes1.Dexes) != len(dexes2.Dexes) {
		t.Errorf("GetDexes() returned different counts: %d vs %d", len(dexes1.Dexes), len(dexes2.Dexes))
	}
}

func TestCachedClient_GetPools(t *testing.T) {
	// Create a standard client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a cached client with a short TTL for testing
	cachedClient := NewCachedClient(client, nil, 500*time.Millisecond)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test with specific options
	opts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}

	// First call should hit the API
	pools1, err := cachedClient.GetPools(ctx, opts)
	if err != nil {
		t.Fatalf("GetPools() first call error = %v", err)
	}

	if pools1 == nil || pools1.Pools == nil {
		t.Fatal("GetPools() first call returned nil")
	}

	// Second call should hit the cache
	startTime := time.Now()
	pools2, err := cachedClient.GetPools(ctx, opts)
	duration := time.Since(startTime)
	if err != nil {
		t.Fatalf("GetPools() second call error = %v", err)
	}

	// Should be very fast from cache
	if duration > 50*time.Millisecond {
		t.Logf("Warning: Cached GetPools call took %v, expected < 50ms", duration)
	}

	// Data should be the same
	if len(pools1.Pools) != len(pools2.Pools) {
		t.Errorf("GetPools() returned different counts: %d vs %d", len(pools1.Pools), len(pools2.Pools))
	}
}

func TestCachedClient_GetNetworkPools(t *testing.T) {
	// Create a standard client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a cached client with a short TTL for testing
	cachedClient := NewCachedClient(client, nil, 500*time.Millisecond)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test with specific network and options
	networkID := "ethereum"
	opts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}

	// First call should hit the API
	pools1, err := cachedClient.GetNetworkPools(ctx, networkID, opts)
	if err != nil {
		t.Fatalf("GetNetworkPools() first call error = %v", err)
	}

	if pools1 == nil || pools1.Pools == nil {
		t.Fatal("GetNetworkPools() first call returned nil")
	}

	// Second call should hit the cache
	startTime := time.Now()
	pools2, err := cachedClient.GetNetworkPools(ctx, networkID, opts)
	duration := time.Since(startTime)
	if err != nil {
		t.Fatalf("GetNetworkPools() second call error = %v", err)
	}

	// Should be very fast from cache
	if duration > 50*time.Millisecond {
		t.Logf("Warning: Cached GetNetworkPools call took %v, expected < 50ms", duration)
	}

	// Data should be the same
	if len(pools1.Pools) != len(pools2.Pools) {
		t.Errorf("GetNetworkPools() returned different counts: %d vs %d", len(pools1.Pools), len(pools2.Pools))
	}

	// Verify all pools are from the correct network
	for _, pool := range pools1.Pools {
		if pool.Chain != networkID {
			t.Errorf("GetNetworkPools returned a pool from wrong network: %s, expected %s", pool.Chain, networkID)
		}
	}
}

func TestCachedClient_GetTokenPools(t *testing.T) {
	// Create a standard client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a cached client with a short TTL for testing
	cachedClient := NewCachedClient(client, nil, 500*time.Millisecond)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test with a specific token
	networkID := "ethereum"
	// #nosec G101
	tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH
	opts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}

	// First call should hit the API
	pools1, err := cachedClient.GetTokenPools(ctx, networkID, tokenAddress, opts, "")
	if err != nil {
		t.Fatalf("GetTokenPools() first call error = %v", err)
	}

	if pools1 == nil || pools1.Pools == nil {
		t.Fatal("GetTokenPools() first call returned nil")
	}

	// Second call should hit the cache
	startTime := time.Now()
	pools2, err := cachedClient.GetTokenPools(ctx, networkID, tokenAddress, opts, "")
	duration := time.Since(startTime)
	if err != nil {
		t.Fatalf("GetTokenPools() second call error = %v", err)
	}

	// Should be very fast from cache
	if duration > 50*time.Millisecond {
		t.Logf("Warning: Cached GetTokenPools call took %v, expected < 50ms", duration)
	}

	// Data should be the same
	if len(pools1.Pools) != len(pools2.Pools) {
		t.Errorf("GetTokenPools() returned different counts: %d vs %d", len(pools1.Pools), len(pools2.Pools))
	}

	// Verify all pools contain the token
	for _, pool := range pools1.Pools {
		tokenFound := false
		for _, token := range pool.Tokens {
			if token.ID == tokenAddress {
				tokenFound = true
				break
			}
		}
		if !tokenFound {
			t.Errorf("GetTokenPools returned a pool without the specified token: %s", pool.ID)
		}
	}
}

func TestCachedClient_GetStats(t *testing.T) {
	// Create a standard client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a cached client with a short TTL for testing
	cachedClient := NewCachedClient(client, nil, 500*time.Millisecond)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First call should hit the API
	stats1, err := cachedClient.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats() first call error = %v", err)
	}

	if stats1 == nil {
		t.Fatal("GetStats() first call returned nil")
	}

	// Second call should hit the cache
	startTime := time.Now()
	stats2, err := cachedClient.GetStats(ctx)
	duration := time.Since(startTime)
	if err != nil {
		t.Fatalf("GetStats() second call error = %v", err)
	}

	// Should be very fast from cache
	if duration > 50*time.Millisecond {
		t.Logf("Warning: Cached GetStats call took %v, expected < 50ms", duration)
	}

	// Data should be the same
	if stats1.Chains != stats2.Chains || stats1.Pools != stats2.Pools || stats1.Tokens != stats2.Tokens {
		t.Errorf("GetStats() returned different data: %+v vs %+v", stats1, stats2)
	}
}
