package dexpaprika

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// TestSearch_SearchWithMock tests the Search method using a mock server
func TestSearch_SearchWithMock(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		query          string
		response       string
		statusCode     int
		expectError    bool
		expectedTokens int
		expectedPools  int
		expectedDexes  int
	}{
		{
			name:           "valid query with results",
			query:          "uniswap",
			response:       `{"tokens": [{"name": "Uniswap", "symbol": "UNI"}], "pools": [{"name": "UNI-ETH"}], "dexes": [{"name": "Uniswap V2"}, {"name": "Uniswap V3"}]}`,
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedTokens: 1,
			expectedPools:  1,
			expectedDexes:  2,
		},
		{
			name:           "valid query with no results",
			query:          "nonexistent",
			response:       `{"tokens": [], "pools": [], "dexes": []}`,
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedTokens: 0,
			expectedPools:  0,
			expectedDexes:  0,
		},
		{
			name:           "empty query",
			query:          "",
			response:       `{"error": "Invalid query param"}`,
			statusCode:     http.StatusBadRequest,
			expectError:    true,
			expectedTokens: 0,
			expectedPools:  0,
			expectedDexes:  0,
		},
		{
			name:           "server error",
			query:          "error",
			response:       `{"error": "Internal server error"}`,
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
			expectedTokens: 0,
			expectedPools:  0,
			expectedDexes:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check that the request is for the search endpoint
				if r.URL.Path != "/search" {
					t.Errorf("Expected request to '/search', got '%s'", r.URL.Path)
				}

				// Check that the query parameter is correct
				query := r.URL.Query().Get("query")
				if query != tc.query {
					t.Errorf("Expected query parameter to be '%s', got '%s'", tc.query, query)
				}

				// Set response headers
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				fmt.Fprintln(w, tc.response)
			}))
			defer server.Close()

			// Create a client that uses the test server
			client := NewClient(
				WithBaseURL(server.URL),
				WithRetryConfig(0, 1*time.Millisecond, 1*time.Millisecond), // No retries for faster tests
			)

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Perform the search
			result, err := client.Search.Search(ctx, tc.query)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If we don't expect an error, check the results
			if !tc.expectError {
				if result == nil {
					t.Fatal("Expected non-nil result but got nil")
				}

				if len(result.Tokens) != tc.expectedTokens {
					t.Errorf("Expected %d tokens, got %d", tc.expectedTokens, len(result.Tokens))
				}

				if len(result.Pools) != tc.expectedPools {
					t.Errorf("Expected %d pools, got %d", tc.expectedPools, len(result.Pools))
				}

				if len(result.Dexes) != tc.expectedDexes {
					t.Errorf("Expected %d dexes, got %d", tc.expectedDexes, len(result.Dexes))
				}
			}
		})
	}
}

// TestSearch_CanceledContext tests behavior with a canceled context
func TestSearch_CanceledContext(t *testing.T) {
	// Create a client
	client := NewClient(
		WithRetryConfig(0, 1*time.Millisecond, 1*time.Millisecond), // No retries for faster tests
	)

	// Create a context and cancel it immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context before making the request

	// Perform the search with a canceled context
	result, err := client.Search.Search(ctx, "test")

	// Expect an error due to canceled context
	if err == nil {
		t.Fatal("Search with canceled context returned no error, expected context canceled error")
	}

	// Check that the error is a context error
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Search returned error %v, want context.Canceled", err)
	}

	// Result should be nil
	if result != nil {
		t.Error("Search with canceled context returned non-nil result, expected nil")
	}
}
