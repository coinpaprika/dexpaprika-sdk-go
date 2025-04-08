package dexpaprika

import (
	"context"
	"fmt"
)

// Paginator is an interface for types that can be paginated
type Paginator interface {
	HasNextPage() bool
	GetNextPage(ctx context.Context) error
}

// PoolsPaginator provides pagination for pools
type PoolsPaginator struct {
	client      *Client
	networkID   string // Optional, for network-specific queries
	dexID       string // Optional, for dex-specific queries
	tokenID     string // Optional, for token-specific queries
	secondToken string // Optional, for filtering token pairs
	options     *ListOptions
	currentResp *PoolsResponse
	err         error
}

// NewPoolsPaginator creates a new paginator for pools
func NewPoolsPaginator(client *Client, opts *ListOptions) *PoolsPaginator {
	if opts == nil {
		opts = &ListOptions{Page: 0, Limit: 50}
	}
	if opts.Limit <= 0 {
		opts.Limit = 50
	}
	return &PoolsPaginator{
		client:  client,
		options: opts,
	}
}

// ForNetwork sets the paginator to fetch pools for a specific network
func (p *PoolsPaginator) ForNetwork(networkID string) *PoolsPaginator {
	p.networkID = networkID
	return p
}

// ForDex sets the paginator to fetch pools for a specific DEX on a network
func (p *PoolsPaginator) ForDex(networkID, dexID string) *PoolsPaginator {
	p.networkID = networkID
	p.dexID = dexID
	return p
}

// ForToken sets the paginator to fetch pools containing a specific token
func (p *PoolsPaginator) ForToken(networkID, tokenID string, secondToken string) *PoolsPaginator {
	p.networkID = networkID
	p.tokenID = tokenID
	p.secondToken = secondToken
	return p
}

// HasNextPage returns true if there are more pages to fetch
func (p *PoolsPaginator) HasNextPage() bool {
	if p.currentResp == nil {
		return true // First page
	}

	if p.err != nil {
		return false
	}

	// Check if we've received fewer items than requested, indicating last page
	if len(p.currentResp.Pools) < p.options.Limit {
		return false
	}

	// Or if the API explicitly tells us there are no more pages
	if p.currentResp.PageInfo.Page+1 >= p.currentResp.PageInfo.TotalPages {
		return false
	}

	return true
}

// GetNextPage fetches the next page of results
func (p *PoolsPaginator) GetNextPage(ctx context.Context) error {
	if !p.HasNextPage() {
		return fmt.Errorf("no more pages")
	}

	// Increment page number if not the first page
	if p.currentResp != nil {
		p.options.Page++
	}

	var resp *PoolsResponse
	var err error

	// Determine which API endpoint to call based on the set parameters
	if p.tokenID != "" {
		// Token pools
		resp, err = p.client.Tokens.GetPools(ctx, p.networkID, p.tokenID, p.options, p.secondToken)
	} else if p.dexID != "" {
		// DEX pools
		resp, err = p.client.Pools.ListByDex(ctx, p.networkID, p.dexID, p.options)
	} else if p.networkID != "" {
		// Network pools
		resp, err = p.client.Pools.ListByNetwork(ctx, p.networkID, p.options)
	} else {
		// All pools
		resp, err = p.client.Pools.List(ctx, p.options)
	}

	if err != nil {
		p.err = err
		return err
	}

	p.currentResp = resp
	return nil
}

// GetCurrentPage returns the current page of results
func (p *PoolsPaginator) GetCurrentPage() []Pool {
	if p.currentResp == nil {
		return nil
	}
	return p.currentResp.Pools
}

// GetError returns any error that occurred while fetching pages
func (p *PoolsPaginator) GetError() error {
	return p.err
}

// DexesPaginator provides pagination for DEXes
type DexesPaginator struct {
	client      *Client
	networkID   string
	page        int
	limit       int
	currentResp *DexesResponse
	err         error
}

// NewDexesPaginator creates a new paginator for DEXes
func NewDexesPaginator(client *Client, networkID string, limit int) *DexesPaginator {
	if limit <= 0 {
		limit = 50
	}
	return &DexesPaginator{
		client:    client,
		networkID: networkID,
		page:      0,
		limit:     limit,
	}
}

// HasNextPage returns true if there are more pages to fetch
func (p *DexesPaginator) HasNextPage() bool {
	if p.currentResp == nil {
		return true // First page
	}

	if p.err != nil {
		return false
	}

	// Check if we've received fewer items than requested, indicating last page
	if len(p.currentResp.Dexes) < p.limit {
		return false
	}

	// Or if the API explicitly tells us there are no more pages
	if p.currentResp.PageInfo.Page+1 >= p.currentResp.PageInfo.TotalPages {
		return false
	}

	return true
}

// GetNextPage fetches the next page of results
func (p *DexesPaginator) GetNextPage(ctx context.Context) error {
	if !p.HasNextPage() {
		return fmt.Errorf("no more pages")
	}

	resp, err := p.client.Networks.ListDexes(ctx, p.networkID, p.page, p.limit)
	if err != nil {
		p.err = err
		return err
	}

	p.currentResp = resp
	p.page++ // Increment page for next call

	return nil
}

// GetCurrentPage returns the current page of results
func (p *DexesPaginator) GetCurrentPage() []Dex {
	if p.currentResp == nil {
		return nil
	}
	return p.currentResp.Dexes
}

// GetError returns any error that occurred while fetching pages
func (p *DexesPaginator) GetError() error {
	return p.err
}

// TransactionsPaginator provides pagination for transactions
type TransactionsPaginator struct {
	client      *Client
	networkID   string
	poolAddress string
	page        int
	limit       int
	cursor      string // Some APIs use cursor-based pagination
	currentResp *TransactionsResponse
	err         error
}

// NewTransactionsPaginator creates a new paginator for transactions
func NewTransactionsPaginator(client *Client, networkID, poolAddress string, limit int) *TransactionsPaginator {
	if limit <= 0 {
		limit = 50
	}
	return &TransactionsPaginator{
		client:      client,
		networkID:   networkID,
		poolAddress: poolAddress,
		page:        0,
		limit:       limit,
	}
}

// HasNextPage returns true if there are more pages to fetch
func (p *TransactionsPaginator) HasNextPage() bool {
	if p.currentResp == nil {
		return true // First page
	}

	if p.err != nil {
		return false
	}

	// Check if we've received fewer items than requested, indicating last page
	if len(p.currentResp.Transactions) < p.limit {
		return false
	}

	// Or if the API explicitly tells us there are no more pages
	if p.currentResp.PageInfo.Page+1 >= p.currentResp.PageInfo.TotalPages {
		return false
	}

	return true
}

// GetNextPage fetches the next page of results
func (p *TransactionsPaginator) GetNextPage(ctx context.Context) error {
	if !p.HasNextPage() {
		return fmt.Errorf("no more pages")
	}

	resp, err := p.client.Pools.GetTransactions(ctx, p.networkID, p.poolAddress, p.page, p.limit, p.cursor)
	if err != nil {
		p.err = err
		return err
	}

	p.currentResp = resp
	p.page++ // Increment page for next call

	// If the API provides a cursor for the next page, use that instead of page number
	if p.currentResp != nil && len(p.currentResp.Transactions) > 0 {
		lastTx := p.currentResp.Transactions[len(p.currentResp.Transactions)-1]
		p.cursor = lastTx.ID // Some APIs use the last ID as cursor
	}

	return nil
}

// GetCurrentPage returns the current page of results
func (p *TransactionsPaginator) GetCurrentPage() []Transaction {
	if p.currentResp == nil {
		return nil
	}
	return p.currentResp.Transactions
}

// GetError returns any error that occurred while fetching pages
func (p *TransactionsPaginator) GetError() error {
	return p.err
}
