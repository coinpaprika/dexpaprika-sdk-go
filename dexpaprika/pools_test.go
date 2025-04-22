package dexpaprika

import (
	"context"
	"testing"
	"time"
)

func TestPools_List(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test getting top pools
	poolsOpts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}
	pools, err := client.Pools.List(ctx, poolsOpts)
	if err != nil {
		t.Fatalf("Pools.List returned error: %v", err)
	}

	if pools == nil {
		t.Fatal("Pools.List returned nil, expected a PoolList")
	}

	if len(pools.Pools) == 0 {
		t.Error("Pools.List returned empty list, expected some pools")
	}

	// Check basic properties of pools
	for _, pool := range pools.Pools {
		if pool.ID == "" {
			t.Error("Pools.List returned pool with empty ID")
		}
		if pool.Chain == "" {
			t.Error("Pools.List returned pool with empty Chain")
		}
		if pool.DexName == "" {
			t.Error("Pools.List returned pool with empty DexName")
		}
	}
}

func TestPools_ListByNetwork(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test getting network-specific pools (using Ethereum as example)
	networkID := "ethereum"
	poolsOpts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}

	pools, err := client.Pools.ListByNetwork(ctx, networkID, poolsOpts)
	if err != nil {
		t.Fatalf("Pools.ListByNetwork returned error: %v", err)
	}

	if pools == nil {
		t.Fatal("Pools.ListByNetwork returned nil, expected a PoolList")
	}

	// All pools should be on the specified network
	for _, pool := range pools.Pools {
		if pool.Chain != networkID {
			t.Errorf("Pools.ListByNetwork returned pool with wrong chain: got %s, want %s", pool.Chain, networkID)
		}
	}
}

func TestPools_ListByDex(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test getting DEX-specific pools (using Uniswap V3 on Ethereum as example)
	networkID := "ethereum"
	dexID := "uniswap_v3"
	poolsOpts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}

	pools, err := client.Pools.ListByDex(ctx, networkID, dexID, poolsOpts)
	if err != nil {
		t.Fatalf("Pools.ListByDex returned error: %v", err)
	}

	if pools == nil {
		t.Fatal("Pools.ListByDex returned nil, expected a PoolList")
	}

	// All pools should be on the specified network and DEX
	for _, pool := range pools.Pools {
		if pool.Chain != networkID {
			t.Errorf("Pools.ListByDex returned pool with wrong chain: got %s, want %s", pool.Chain, networkID)
		}
		// DEX names might not exactly match the DEX ID, so we don't check this strictly
	}
}

func TestPools_GetDetails(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use a well-known Ethereum USDC-WETH Uniswap V3 pool
	networkID := "ethereum"
	poolID := "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"

	// Test getting pool details
	details, err := client.Pools.GetDetails(ctx, networkID, poolID, false)
	if err != nil {
		t.Fatalf("Pools.GetDetails returned error: %v", err)
	}

	if details == nil {
		t.Fatal("Pools.GetDetails returned nil, expected pool details")
	}

	// Check the pool details contain relevant information
	if details.ID != poolID {
		t.Errorf("Pools.GetDetails returned wrong pool ID: got %s, want %s", details.ID, poolID)
	}
	if details.Chain != networkID {
		t.Errorf("Pools.GetDetails returned wrong chain: got %s, want %s", details.Chain, networkID)
	}

	// Verify tokens
	hasUSDC := false
	hasWETH := false
	for _, token := range details.Tokens {
		if token.Symbol == "USDC" {
			hasUSDC = true
		}
		if token.Symbol == "WETH" {
			hasWETH = true
		}
	}

	if !hasUSDC {
		t.Error("Pool details missing USDC token")
	}
	if !hasWETH {
		t.Error("Pool details missing WETH token")
	}
}

func TestPools_GetOHLCV(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use a well-known Ethereum USDC-WETH Uniswap V3 pool
	networkID := "ethereum"
	poolID := "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"

	// Use specific date format for the API
	// Get yesterday and today in the format YYYY-MM-DD
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	// Format the dates properly
	startDate := yesterday.Format("2006-01-02")
	endDate := today.Format("2006-01-02")

	t.Logf("Getting OHLCV data from %s to %s", startDate, endDate)

	// Test getting OHLCV data with specific date format
	ohlcvOpts := &OHLCVOptions{
		Start:    startDate,
		End:      endDate,
		Interval: "1h",
		Limit:    3,
	}

	ohlcv, err := client.Pools.GetOHLCV(ctx, networkID, poolID, ohlcvOpts)
	if err != nil {
		t.Fatalf("Pools.GetOHLCV returned error: %v", err)
	}

	// The API should return data for this timeframe
	if len(ohlcv) == 0 {
		t.Log("Warning: No OHLCV data available for the selected timeframe")
	} else {
		t.Logf("Retrieved %d OHLCV records", len(ohlcv))

		// Check basic properties of OHLCV data
		for _, record := range ohlcv {
			if record.TimeOpen == "" {
				t.Error("Pools.GetOHLCV returned record with empty TimeOpen")
			}
		}
	}
}
