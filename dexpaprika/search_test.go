package dexpaprika

import (
	"context"
	"testing"
	"time"
)

func TestSearch_Search(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with longer timeout (30 seconds instead of 10)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test searching for a common term
	query := "ethereum"
	results, err := client.Search.Search(ctx, query)
	if err != nil {
		t.Fatalf("Search.Search returned error: %v", err)
	}

	if results == nil {
		t.Fatal("Search.Search returned nil, expected search results")
	}

	// Check that we got some results back
	if len(results.Tokens) == 0 && len(results.Pools) == 0 && len(results.Dexes) == 0 {
		t.Error("Search.Search returned empty results for query 'ethereum'")
	}

	// Test specific query handling with individual timeouts
	queries := []struct {
		query   string
		timeout time.Duration
		desc    string
	}{
		{"uniswap", 10 * time.Second, "should find DEXes"},
		{"bitcoin", 10 * time.Second, "should find tokens"},
		{"eth", 15 * time.Second, "should find tokens"},              // Allow more time for this common term
		{"0xc02aaa39b2", 15 * time.Second, "partial address search"}, // Allow more time for address search
	}

	for _, q := range queries {
		t.Run("Query:"+q.query, func(t *testing.T) {
			// Create a separate context with appropriate timeout for each query
			queryCtx, queryCancel := context.WithTimeout(context.Background(), q.timeout)
			defer queryCancel()

			results, err := client.Search.Search(queryCtx, q.query)
			if err != nil {
				t.Errorf("Search.Search(%q) returned error: %v", q.query, err)
				return
			}

			if results == nil {
				t.Errorf("Search.Search(%q) returned nil results", q.query)
				return
			}

			t.Logf("Query %q (%s) returned: %d tokens, %d pools, %d dexes",
				q.query, q.desc, len(results.Tokens), len(results.Pools), len(results.Dexes))
		})
	}
}

func TestSearch_InvalidQuery(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test with empty query
	results, err := client.Search.Search(ctx, "")
	if err != nil {
		// The API might reject empty queries
		t.Logf("Empty query rejected as expected: %v", err)
		return
	}

	// If the API accepts empty queries, we should still get a valid (possibly empty) result
	if results == nil {
		t.Error("Search.Search with empty query returned nil, expected empty results")
	}
}
