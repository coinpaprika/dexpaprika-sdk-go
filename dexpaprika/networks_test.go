package dexpaprika

import (
	"context"
	"testing"
	"time"
)

func TestNetworks_List(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test getting networks
	networks, err := client.Networks.List(ctx)
	if err != nil {
		t.Fatalf("Networks.List returned error: %v", err)
	}

	if len(networks) == 0 {
		t.Error("Networks.List returned empty list, expected some networks")
	}

	// Check basic properties of networks
	for _, network := range networks {
		if network.ID == "" {
			t.Error("Networks.List returned network with empty ID")
		}
		if network.DisplayName == "" {
			t.Error("Networks.List returned network with empty DisplayName")
		}
	}
}

func TestNetworks_ListDexes(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// We need networks to test ListDexes
	networks, err := client.Networks.List(ctx)
	if err != nil {
		t.Fatalf("Networks.List returned error: %v", err)
	}

	if len(networks) == 0 {
		t.Skip("No networks available to test ListDexes")
	}

	// Use the first network for testing
	networkID := networks[0].ID

	// Test getting DEXes for a network
	dexes, err := client.Networks.ListDexes(ctx, networkID, 0, 5)
	if err != nil {
		t.Fatalf("Networks.ListDexes returned error: %v", err)
	}

	if dexes == nil {
		t.Fatal("Networks.ListDexes returned nil, expected a DexList")
	}

	// If we got dexes back, check their properties
	for _, dex := range dexes.Dexes {
		if dex.ID == "" {
			t.Error("Networks.ListDexes returned dex with empty ID (dex_id field)")
		}
		if dex.Name == "" {
			t.Error("Networks.ListDexes returned dex with empty Name (dex_name field)")
		}
		if dex.Chain == "" {
			t.Error("Networks.ListDexes returned dex with empty Chain")
		}
		if dex.Protocol == "" {
			t.Error("Networks.ListDexes returned dex with empty Protocol")
		}
	}
}

func TestCachedClient_Networks(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a cached client
	cachedClient := NewCachedClient(client, nil, 5*time.Minute)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test networks retrieval from cache
	networks, err := cachedClient.GetNetworks(ctx)
	if err != nil {
		t.Fatalf("CachedClient.GetNetworks returned error: %v", err)
	}

	if len(networks) == 0 {
		t.Error("CachedClient.GetNetworks returned empty list, expected some networks")
	}

	// Get networks again to test cache
	start := time.Now()
	networksAgain, err := cachedClient.GetNetworks(ctx)
	if err != nil {
		t.Fatalf("CachedClient.GetNetworks (again) returned error: %v", err)
	}
	duration := time.Since(start)

	// Cached response should be very fast
	if duration > 100*time.Millisecond {
		t.Logf("Warning: Cached response took longer than expected: %v", duration)
	}

	// Result should be the same length
	if len(networksAgain) != len(networks) {
		t.Errorf("Cache inconsistency: got %d networks on second call, want %d", len(networksAgain), len(networks))
	}
}
