package dexpaprika

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Cache is an interface for types that can be used as caches
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
}

// InMemoryCache provides a simple in-memory cache
type InMemoryCache struct {
	items map[string]*cacheItem
	mu    sync.RWMutex
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		items: make(map[string]*cacheItem),
	}

	// Start a cleanup routine
	go cache.cleanup()

	return cache
}

// Get retrieves an item from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Check if the item has expired
	if time.Now().After(item.expiresAt) {
		return nil, false
	}

	return item.value, true
}

// Set adds an item to the cache with a TTL
func (c *InMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete removes an item from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
}

// cleanup periodically removes expired items from the cache
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		for key, item := range c.items {
			if time.Now().After(item.expiresAt) {
				delete(c.items, key)
			}
		}

		c.mu.Unlock()
	}
}

// CachedClient wraps a Client with caching functionality
type CachedClient struct {
	client *Client
	cache  Cache
	ttl    time.Duration
}

// NewCachedClient creates a new client with caching
func NewCachedClient(client *Client, cache Cache, ttl time.Duration) *CachedClient {
	if cache == nil {
		cache = NewInMemoryCache()
	}

	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	return &CachedClient{
		client: client,
		cache:  cache,
		ttl:    ttl,
	}
}

// GetNetworks retrieves networks with caching
func (c *CachedClient) GetNetworks(ctx context.Context) ([]Network, error) {
	cacheKey := "networks"

	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if networks, ok := cachedValue.([]Network); ok {
			return networks, nil
		}
	}

	// If not in cache or wrong type, fetch from API
	networks, err := c.client.Networks.List(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(cacheKey, networks, c.ttl)

	return networks, nil
}

// GetDexes retrieves DEXes with caching
func (c *CachedClient) GetDexes(ctx context.Context, networkID string, page, limit int) (*DexesResponse, error) {
	cacheKey := fmt.Sprintf("dexes:%s:%d:%d", networkID, page, limit)

	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if dexes, ok := cachedValue.(*DexesResponse); ok {
			return dexes, nil
		}
	}

	// If not in cache or wrong type, fetch from API
	dexes, err := c.client.Networks.ListDexes(ctx, networkID, page, limit)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(cacheKey, dexes, c.ttl)

	return dexes, nil
}

// GetPools retrieves pools with caching
func (c *CachedClient) GetPools(ctx context.Context, opts *ListOptions) (*PoolsResponse, error) {
	var optsPage, optsLimit int
	var optsSort, optsOrderBy string

	if opts != nil {
		optsPage = opts.Page
		optsLimit = opts.Limit
		optsSort = opts.Sort
		optsOrderBy = opts.OrderBy
	}

	cacheKey := fmt.Sprintf("pools:%d:%d:%s:%s", optsPage, optsLimit, optsSort, optsOrderBy)

	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if pools, ok := cachedValue.(*PoolsResponse); ok {
			return pools, nil
		}
	}

	// If not in cache or wrong type, fetch from API
	pools, err := c.client.Pools.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(cacheKey, pools, c.ttl)

	return pools, nil
}

// GetNetworkPools retrieves network pools with caching
func (c *CachedClient) GetNetworkPools(ctx context.Context, networkID string, opts *ListOptions) (*PoolsResponse, error) {
	var optsPage, optsLimit int
	var optsSort, optsOrderBy string

	if opts != nil {
		optsPage = opts.Page
		optsLimit = opts.Limit
		optsSort = opts.Sort
		optsOrderBy = opts.OrderBy
	}

	cacheKey := fmt.Sprintf("network_pools:%s:%d:%d:%s:%s", networkID, optsPage, optsLimit, optsSort, optsOrderBy)

	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if pools, ok := cachedValue.(*PoolsResponse); ok {
			return pools, nil
		}
	}

	// If not in cache or wrong type, fetch from API
	pools, err := c.client.Pools.ListByNetwork(ctx, networkID, opts)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(cacheKey, pools, c.ttl)

	return pools, nil
}

// GetPoolDetails retrieves pool details with caching
func (c *CachedClient) GetPoolDetails(ctx context.Context, networkID, poolAddress string, inversed bool) (*PoolDetails, error) {
	cacheKey := fmt.Sprintf("pool_details:%s:%s:%t", networkID, poolAddress, inversed)

	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if details, ok := cachedValue.(*PoolDetails); ok {
			return details, nil
		}
	}

	// If not in cache or wrong type, fetch from API
	details, err := c.client.Pools.GetDetails(ctx, networkID, poolAddress, inversed)
	if err != nil {
		return nil, err
	}

	// Store in cache for a shorter time since prices change frequently
	c.cache.Set(cacheKey, details, c.ttl/5)

	return details, nil
}

// GetTokenDetails retrieves token details with caching
func (c *CachedClient) GetTokenDetails(ctx context.Context, networkID, tokenAddress string) (*TokenDetails, error) {
	cacheKey := fmt.Sprintf("token_details:%s:%s", networkID, tokenAddress)

	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if details, ok := cachedValue.(*TokenDetails); ok {
			return details, nil
		}
	}

	// If not in cache or wrong type, fetch from API
	details, err := c.client.Tokens.GetDetails(ctx, networkID, tokenAddress)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(cacheKey, details, c.ttl)

	return details, nil
}

// GetTokenPools retrieves token pools with caching
func (c *CachedClient) GetTokenPools(ctx context.Context, networkID, tokenAddress string, opts *ListOptions, additionalTokenAddress string) (*PoolsResponse, error) {
	var optsPage, optsLimit int
	var optsSort, optsOrderBy string

	if opts != nil {
		optsPage = opts.Page
		optsLimit = opts.Limit
		optsSort = opts.Sort
		optsOrderBy = opts.OrderBy
	}

	cacheKey := fmt.Sprintf("token_pools:%s:%s:%d:%d:%s:%s:%s", networkID, tokenAddress, optsPage, optsLimit, optsSort, optsOrderBy, additionalTokenAddress)

	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if pools, ok := cachedValue.(*PoolsResponse); ok {
			return pools, nil
		}
	}

	// If not in cache or wrong type, fetch from API
	pools, err := c.client.Tokens.GetPools(ctx, networkID, tokenAddress, opts, additionalTokenAddress)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(cacheKey, pools, c.ttl)

	return pools, nil
}

// GetStats retrieves DexPaprika stats with caching
func (c *CachedClient) GetStats(ctx context.Context) (*Stats, error) {
	cacheKey := "stats"

	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if stats, ok := cachedValue.(*Stats); ok {
			return stats, nil
		}
	}

	// If not in cache or wrong type, fetch from API
	stats, err := c.client.Utils.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(cacheKey, stats, c.ttl)

	return stats, nil
}
