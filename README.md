# DexPaprika Go SDK

A production-ready Go client for the DexPaprika API, providing access to decentralized exchange (DEX) data across multiple blockchain networks.

## Installation

```bash
go get github.com/donbagger/dexpaprika-sdk-go
```

## Features

- **Complete API Coverage**: Access all DexPaprika API endpoints
- **Production-Ready**:
  - Automatic retry mechanism with exponential backoff
  - Comprehensive error handling with typed errors
  - Rate limiting support
  - Pagination helpers
  - Caching layer for improved performance
  - Flexible configuration via functional options
- **Easy to Use**: Simple, intuitive interfaces for working with DEX data

## Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/donbagger/dexpaprika-sdk-go/dexpaprika"
)

func main() {
    // Create a new client with default settings
    client := dexpaprika.NewClient()

    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Get a list of all supported networks
    networks, err := client.Networks.List(ctx)
    if err != nil {
        log.Fatalf("Failed to get networks: %v", err)
    }
    fmt.Printf("Found %d supported networks\n", len(networks))

    // Get top pools with pagination
    poolsOpts := &dexpaprika.ListOptions{
        Limit:   10,
        Page:    0,
        OrderBy: "volume_usd",
        Sort:    "desc",
    }
    pools, err := client.Pools.List(ctx, poolsOpts)
    if err != nil {
        log.Fatalf("Failed to get pools: %v", err)
    }
    fmt.Printf("Found %d pools\n", len(pools.Pools))
}
```

## Advanced Configuration

```go
// Create a client with custom settings
client := dexpaprika.NewClient(
    // Custom HTTP client with longer timeout
    dexpaprika.WithHTTPClient(&http.Client{
        Timeout: 60 * time.Second,
    }),
    // Custom retry settings
    dexpaprika.WithRetryConfig(5, 2*time.Second, 30*time.Second),
    // Rate limiting to 5 requests per second
    dexpaprika.WithRateLimit(5.0),
    // Custom base URL (for testing or using a proxy)
    dexpaprika.WithBaseURL("https://your-proxy.example.com/dexpaprika"),
)
```

## Using Caching

```go
// Create a basic client
client := dexpaprika.NewClient()

// Create a cached client with default settings (in-memory cache, 5-minute TTL)
cachedClient := dexpaprika.NewCachedClient(client, nil, 0)

// Get networks (will be cached)
networks, err := cachedClient.GetNetworks(ctx)
if err != nil {
    log.Fatalf("Failed to get networks: %v", err)
}

// Get pools (will be cached)
pools, err := cachedClient.GetPools(ctx, poolsOpts)
if err != nil {
    log.Fatalf("Failed to get pools: %v", err)
}

// Get networks again (will be served from cache)
networks, err = cachedClient.GetNetworks(ctx)
```

## Pagination Helpers

```go
// Create a pools paginator
paginator := dexpaprika.NewPoolsPaginator(client, &dexpaprika.ListOptions{
    Limit: 100,
    OrderBy: "volume_usd",
    Sort: "desc",
})

// Specify that we want Ethereum pools
paginator.ForNetwork("ethereum")

// Process all pages
for paginator.HasNextPage() {
    if err := paginator.GetNextPage(ctx); err != nil {
        log.Fatalf("Failed to get page: %v", err)
    }
    
    // Process the current page
    pools := paginator.GetCurrentPage()
    for _, pool := range pools {
        fmt.Printf("Pool: %s on %s (Volume: $%.2f)\n", 
            pool.DexName, 
            pool.Chain, 
            pool.VolumeUSD)
    }
}
```

## Handling Errors

```go
pools, err := client.Pools.List(ctx, poolsOpts)
if err != nil {
    var apiErr *dexpaprika.APIError
    
    // Check for specific error types
    if errors.As(err, &apiErr) {
        // Access details about the API error
        fmt.Printf("API Error: %s (Status Code: %d)\n", 
            apiErr.Message, 
            apiErr.StatusCode)
        
        // Check for specific error conditions
        if errors.Is(err, dexpaprika.ErrRateLimit) {
            fmt.Println("Rate limit exceeded, try again later")
        } else if errors.Is(err, dexpaprika.ErrNotFound) {
            fmt.Println("Resource not found")
        }
    } else {
        fmt.Printf("Other error: %v\n", err)
    }
    return
}
```

## API Documentation

### Networks

```go
// Get a list of all supported blockchain networks
networks, err := client.Networks.List(ctx)

// Get DEXes on a specific network
dexes, err := client.Networks.ListDexes(ctx, "ethereum", 0, 10)
```

### Pools

```go
// Get top pools from all networks
pools, err := client.Pools.List(ctx, &dexpaprika.ListOptions{
    Limit:   10,
    OrderBy: "volume_usd",
    Sort:    "desc",
})

// Get pools on a specific network
networkPools, err := client.Pools.ListByNetwork(ctx, "ethereum", opts)

// Get pools on a specific DEX
dexPools, err := client.Pools.ListByDex(ctx, "ethereum", "uniswap_v3", opts)

// Get details about a specific pool
poolDetails, err := client.Pools.GetDetails(ctx, "ethereum", "0xpool_address", false)

// Get OHLCV data for a pool
ohlcv, err := client.Pools.GetOHLCV(ctx, "ethereum", "0xpool_address", &dexpaprika.OHLCVOptions{
    Start:    "2023-01-01",
    End:      "2023-01-31",
    Interval: "24h",
    Limit:    30,
})

// Get transactions for a pool
transactions, err := client.Pools.GetTransactions(ctx, "ethereum", "0xpool_address", 0, 10, "")
```

### Tokens

```go
// Get details about a specific token
tokenDetails, err := client.Tokens.GetDetails(ctx, "ethereum", "0xtoken_address")

// Get pools that contain a specific token
tokenPools, err := client.Tokens.GetPools(ctx, "ethereum", "0xtoken_address", opts, "")

// Get pools that contain a pair of tokens
pairPools, err := client.Tokens.GetPools(ctx, "ethereum", "0xtoken1_address", opts, "0xtoken2_address")
```

### Search

```go
// Search for tokens, pools, and DEXes
results, err := client.Search.Search(ctx, "query")
```

### Utils

```go
// Get global stats
stats, err := client.Utils.GetStats(ctx)
```

## License

This project is licensed under the MIT License. 