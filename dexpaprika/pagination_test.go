package dexpaprika

import (
	"context"
	"testing"
	"time"
)

func TestPoolsPaginator(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Create a paginator with a small limit to test pagination
	paginator := NewPoolsPaginator(client, &ListOptions{
		Limit:   5,
		Page:    0,
		OrderBy: "volume_usd",
		Sort:    "desc",
	})

	// Test initial state
	if paginator.currentResp != nil {
		t.Error("NewPoolsPaginator() initialized with non-nil currentResp")
	}

	if !paginator.HasNextPage() {
		t.Error("NewPoolsPaginator().HasNextPage() = false, want true for initial state")
	}

	// Test GetCurrentPage with no fetched data
	initialPage := paginator.GetCurrentPage()
	if initialPage != nil {
		t.Errorf("GetCurrentPage() before fetching = %v, want nil", initialPage)
	}

	// Fetch first page
	err := paginator.GetNextPage(ctx)
	if err != nil {
		t.Fatalf("GetNextPage() first page error = %v", err)
	}

	// Verify we got results
	firstPage := paginator.GetCurrentPage()
	if firstPage == nil || len(firstPage) == 0 {
		t.Fatal("GetCurrentPage() after first fetch returned no pools")
	}

	if len(firstPage) > 5 {
		t.Errorf("GetCurrentPage() returned %d pools, want <= 5", len(firstPage))
	}

	// Get the page info for comparison
	firstPageNum := 0
	if paginator.currentResp != nil && paginator.currentResp.PageInfo.TotalPages > 0 {
		firstPageNum = paginator.currentResp.PageInfo.Page
		t.Logf("First page: %d of %d total pages", firstPageNum, paginator.currentResp.PageInfo.TotalPages)
	}

	// Only try to fetch a second page if HasNextPage indicates there is one
	if paginator.HasNextPage() {
		// Fetch second page
		err = paginator.GetNextPage(ctx)
		if err != nil {
			t.Fatalf("GetNextPage() second page error = %v", err)
		}

		// Verify page incremented
		if paginator.currentResp != nil && paginator.currentResp.PageInfo.Page <= firstPageNum {
			t.Errorf("Second page number = %d, want > %d", paginator.currentResp.PageInfo.Page, firstPageNum)
		}

		secondPage := paginator.GetCurrentPage()
		if secondPage == nil || len(secondPage) == 0 {
			t.Error("GetCurrentPage() after second fetch returned no pools")
		}

		// Verify different data between pages
		if len(firstPage) > 0 && len(secondPage) > 0 && firstPage[0].ID == secondPage[0].ID {
			t.Error("First and second page contain the same first pool ID, pagination may not be working")
		}
	} else {
		t.Log("No second page available to test")
	}
}

func TestPoolsPaginator_ForNetwork(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a paginator for a specific network
	paginator := NewPoolsPaginator(client, &ListOptions{Limit: 3}).ForNetwork("ethereum")

	// Test initial state
	if paginator.networkID != "ethereum" {
		t.Errorf("ForNetwork() networkID = %q, want %q", paginator.networkID, "ethereum")
	}

	// Fetch page
	err := paginator.GetNextPage(ctx)
	if err != nil {
		t.Fatalf("GetNextPage() error = %v", err)
	}

	// Verify we got results for the correct network
	pools := paginator.GetCurrentPage()
	if pools == nil || len(pools) == 0 {
		t.Fatal("GetCurrentPage() returned no pools")
	}

	// Check that all pools are from the correct network
	for i, pool := range pools {
		if pool.Chain != "ethereum" {
			t.Errorf("Pool[%d].Chain = %q, want %q", i, pool.Chain, "ethereum")
		}
	}
}

func TestDexesPaginator(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a paginator for a specific network with small limit
	networkID := "ethereum"
	paginator := NewDexesPaginator(client, networkID, 3)

	// Test initial state
	if paginator.networkID != networkID {
		t.Errorf("NewDexesPaginator() networkID = %q, want %q", paginator.networkID, networkID)
	}

	if paginator.limit != 3 {
		t.Errorf("NewDexesPaginator() limit = %d, want %d", paginator.limit, 3)
	}

	if !paginator.HasNextPage() {
		t.Error("NewDexesPaginator().HasNextPage() = false, want true for initial state")
	}

	// Fetch first page
	err := paginator.GetNextPage(ctx)
	if err != nil {
		t.Fatalf("GetNextPage() error = %v", err)
	}

	// Verify we got results
	dexes := paginator.GetCurrentPage()
	if dexes == nil || len(dexes) == 0 {
		t.Fatal("GetCurrentPage() returned no dexes")
	}

	t.Logf("Found %d dexes on %s", len(dexes), networkID)
}

func TestTransactionsPaginator(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, get a valid pool ID from ethereum
	poolsResp, err := client.Pools.ListByNetwork(ctx, "ethereum", &ListOptions{Limit: 1})
	if err != nil || len(poolsResp.Pools) == 0 {
		t.Skip("Could not get a pool to test transactions paginator")
	}

	networkID := "ethereum"
	poolAddress := poolsResp.Pools[0].ID

	// Create a transactions paginator
	paginator := NewTransactionsPaginator(client, networkID, poolAddress, 5)

	// Test initial state
	if paginator.networkID != networkID {
		t.Errorf("NewTransactionsPaginator() networkID = %q, want %q", paginator.networkID, networkID)
	}

	if paginator.poolAddress != poolAddress {
		t.Errorf("NewTransactionsPaginator() poolAddress = %q, want %q", paginator.poolAddress, poolAddress)
	}

	if !paginator.HasNextPage() {
		t.Error("NewTransactionsPaginator().HasNextPage() = false, want true for initial state")
	}

	// Fetch first page
	err = paginator.GetNextPage(ctx)
	if err != nil {
		t.Logf("GetNextPage() error = %v (this may be normal if the pool has no transactions)", err)
		return
	}

	// Verify we got results
	txns := paginator.GetCurrentPage()
	if txns == nil {
		t.Error("GetCurrentPage() returned nil")
	}

	t.Logf("Found %d transactions for pool %s", len(txns), poolAddress)
}

func TestPoolsPaginator_ForDex(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, get a valid dex ID from ethereum
	dexesResp, err := client.Networks.ListDexes(ctx, "ethereum", 0, 1)
	if err != nil || len(dexesResp.Dexes) == 0 {
		t.Skip("Could not get a dex to test ForDex method")
	}

	networkID := "ethereum"
	dexID := dexesResp.Dexes[0].ID

	// Create a paginator for a specific dex
	paginator := NewPoolsPaginator(client, &ListOptions{Limit: 3}).ForDex(networkID, dexID)

	// Test initial state
	if paginator.networkID != networkID {
		t.Errorf("ForDex() networkID = %q, want %q", paginator.networkID, networkID)
	}

	if paginator.dexID != dexID {
		t.Errorf("ForDex() dexID = %q, want %q", paginator.dexID, dexID)
	}

	// Fetch page
	err = paginator.GetNextPage(ctx)
	if err != nil {
		t.Fatalf("GetNextPage() error = %v", err)
	}

	// Verify we got results
	pools := paginator.GetCurrentPage()
	if pools == nil || len(pools) == 0 {
		t.Skip("No pools returned for the dex, cannot fully test")
	}

	t.Logf("Found %d pools for dex %s on %s", len(pools), dexID, networkID)
}

func TestPoolsPaginator_ForToken(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use well-known token (WETH on Ethereum)
	networkID := "ethereum"
	// #nosec G101
	tokenID := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH
	secondToken := ""                                       // No second token filter

	// Create a paginator for a specific token
	paginator := NewPoolsPaginator(client, &ListOptions{Limit: 3}).ForToken(networkID, tokenID, secondToken)

	// Test initial state
	if paginator.networkID != networkID {
		t.Errorf("ForToken() networkID = %q, want %q", paginator.networkID, networkID)
	}

	if paginator.tokenID != tokenID {
		t.Errorf("ForToken() tokenID = %q, want %q", paginator.tokenID, tokenID)
	}

	if paginator.secondToken != secondToken {
		t.Errorf("ForToken() secondToken = %q, want %q", paginator.secondToken, secondToken)
	}

	// Fetch page
	err := paginator.GetNextPage(ctx)
	if err != nil {
		t.Fatalf("GetNextPage() error = %v", err)
	}

	// Verify we got results
	pools := paginator.GetCurrentPage()
	if pools == nil || len(pools) == 0 {
		t.Fatal("GetCurrentPage() returned no pools")
	}

	// Check that all pools contain the specified token
	for i, pool := range pools {
		tokenFound := false
		for _, token := range pool.Tokens {
			if token.ID == tokenID {
				tokenFound = true
				break
			}
		}
		if !tokenFound {
			t.Errorf("Pool[%d] does not contain the specified token %s", i, tokenID)
		}
	}

	t.Logf("Found %d pools containing token %s on %s", len(pools), tokenID, networkID)
}

func TestPaginator_GetError(t *testing.T) {
	// Test for PoolsPaginator
	poolsPaginator := NewPoolsPaginator(&Client{}, nil)
	if err := poolsPaginator.GetError(); err != nil {
		t.Errorf("New PoolsPaginator GetError() = %v, want nil", err)
	}

	// Test for DexesPaginator
	dexesPaginator := NewDexesPaginator(&Client{}, "ethereum", 10)
	if err := dexesPaginator.GetError(); err != nil {
		t.Errorf("New DexesPaginator GetError() = %v, want nil", err)
	}

	// Test for TransactionsPaginator
	txPaginator := NewTransactionsPaginator(&Client{}, "ethereum", "0x123", 10)
	if err := txPaginator.GetError(); err != nil {
		t.Errorf("New TransactionsPaginator GetError() = %v, want nil", err)
	}
}

func TestTransactionsPaginator_GetErrorWithBadNetwork(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a paginator with an invalid network
	paginator := NewTransactionsPaginator(client, "invalid-network", "0x123", 5)

	// This should fail
	err := paginator.GetNextPage(ctx)
	if err == nil {
		t.Fatal("GetNextPage() with invalid network returned no error, expected an error")
	}

	// Error should now be stored in the paginator
	storedErr := paginator.GetError()
	if storedErr == nil {
		t.Fatal("GetError() after error occurred returned nil, expected an error")
	}

	// Make sure the error matches what we got from GetNextPage
	if storedErr != err {
		t.Errorf("GetError() = %v, want %v", storedErr, err)
	}
}
