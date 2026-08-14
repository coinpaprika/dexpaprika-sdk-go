[![Check & test & build](https://github.com/coinpaprika/dexpaprika-sdk-go/actions/workflows/main.yml/badge.svg)](https://github.com/coinpaprika/dexpaprika-sdk-go/actions/workflows/main.yml) 

# DexPaprika Go SDK

[![Go Tests & Linting](https://github.com/coinpaprika/dexpaprika-sdk-go/actions/workflows/main.yml/badge.svg)](https://github.com/coinpaprika/dexpaprika-sdk-go/actions/workflows/main.yml)

A production-ready Go client for the DexPaprika API, providing access to decentralized exchange (DEX) data across multiple blockchain networks.

## Overview

The DexPaprika API lets you access data on decentralized exchanges across multiple blockchains, including pools, tokens, transactions, and pricing information. This SDK provides a clean, idiomatic Go interface to that API.

## Installation

```bash
go get github.com/coinpaprika/dexpaprika-sdk-go
```

## Requirements

- Go 1.26 or higher
- No API key required to start. A free key or a paid plan raises the quota, see https://docs.dexpaprika.com/knowledge-base/rate-limits

## Testing the SDK

The repository includes a comprehensive test executable that verifies all SDK functionality against the live DexPaprika API:

```bash
# Run the pre-compiled comprehensive test
make test
```

This executable tests all major features, including:
- Networks endpoints
- Pools endpoints
- Tokens endpoints
- Search functionality
- API statistics
- Pagination
- Error handling

You can also rebuild this test executable from the source:

```bash
go build
```

Running this test is a quick way to verify the SDK is working correctly in your environment.

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

## Getting Started

Here's a quick example to get you started with the SDK:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/coinpaprika/dexpaprika-sdk-go/dexpaprika"
)

func main() {
    // Create a new client
    client := dexpaprika.NewClient()

    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Get a list of networks (blockchains)
    networks, err := client.Networks.List(ctx)
    if err != nil {
        log.Fatalf("Error fetching networks: %v", err)
    }
    fmt.Printf("Available networks: %d\n", len(networks))
    
    // Example: Print the first few networks
    for i, network := range networks {
        if i >= 3 {
            break
        }
        fmt.Printf("  - %s (%s)\n", network.DisplayName, network.ID)
    }
    
    // Get top trading pools from Ethereum network.
    // ListByNetwork targets the unified /pools/search endpoint. It is
    // cursor-paginated and reports 24h volume in VolumeUSD24h (a pointer).
    pools, err := client.Pools.ListByNetwork(ctx, "ethereum", &dexpaprika.ListOptions{
        Limit:   5,
        OrderBy: "volume_usd_24h",
        Sort:    "desc",
    })
    if err != nil {
        log.Fatalf("Error fetching pools: %v", err)
    }
    
    fmt.Println("\nTop trading pools on Ethereum:")
    for _, pool := range pools.Pools {
        vol24h := 0.0
        if pool.VolumeUSD24h != nil {
            vol24h = *pool.VolumeUSD24h
        }
        fmt.Printf("  - %s on %s (24h Volume: $%.2f)\n",
            pool.DexName,
            pool.Chain,
            vol24h)
    }
}
```

## Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/coinpaprika/dexpaprika-sdk-go/dexpaprika"
)

func main() {
    // Create a new client with default settings
    client := dexpaprika.NewClient()

    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Get top pools from Ethereum network.
    // ListByNetwork uses the cursor-paginated /pools/search endpoint; Page is
    // ignored, use the Cursor option to page through results.
    poolsOpts := &dexpaprika.ListOptions{
        Limit:   10,
        OrderBy: "volume_usd_24h",
        Sort:    "desc",
    }
    pools, err := client.Pools.ListByNetwork(ctx, "ethereum", poolsOpts)
    if err != nil {
        log.Fatalf("Failed to get pools: %v", err)
    }
    fmt.Printf("Found %d pools on Ethereum\n", len(pools.Pools))
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

The SDK provides a caching layer to improve performance and reduce API calls:

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

// Get top pools (will be cached)
pools, err := cachedClient.GetPools(ctx, &dexpaprika.ListOptions{
    Limit: 10,
    OrderBy: "volume_usd",
    Sort: "desc",
})
if err != nil {
    log.Fatalf("Failed to get pools: %v", err)
}

// Get networks again (will be served from cache)
networks, err = cachedClient.GetNetworks(ctx)
```

## Pagination Helpers

For endpoints that return large collections, the SDK provides pagination helpers:

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
        vol24h := 0.0
        if pool.VolumeUSD24h != nil {
            vol24h = *pool.VolumeUSD24h
        }
        fmt.Printf("Pool: %s on %s (24h Volume: $%.2f)\n",
            pool.DexName,
            pool.Chain,
            vol24h)
    }
}
```

## Handling Errors

The SDK provides detailed error types to help you handle different failure scenarios:

```go
// Use network-specific endpoint
pools, err := client.Pools.ListByNetwork(ctx, "ethereum", poolsOpts)
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
        } else if errors.Is(err, dexpaprika.ErrGone) {
            fmt.Println("This endpoint has been deprecated")
            // The error message will contain migration guidance
            fmt.Println(apiErr.Message)
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
// DEPRECATED: Global pools method has been removed in API v1.3.0
// pools, err := client.Pools.List(ctx, opts) // This will return 410 Gone

// Use network-specific endpoints instead:
// Get pools on a specific network
networkPools, err := client.Pools.ListByNetwork(ctx, "ethereum", &dexpaprika.ListOptions{
    Limit:   10,
    OrderBy: "volume_usd",
    Sort:    "desc",
})

// Get pools on a specific DEX.
// /networks/{network}/dexes/{dex}/pools was removed (HTTP 410), so this now
// targets /pools/search with a dex_name filter. It is cursor-paginated: pass
// Cursor (from a response's NextCursor) to fetch the next page.
dexPools, err := client.Pools.ListByDex(ctx, "ethereum", "uniswap_v3", &dexpaprika.ListOptions{
    Limit:   10,
    OrderBy: "volume_usd_24h",
    Sort:    "desc",
})

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

Note on DEX pools: the API removed `/networks/{network}/dexes/{dex}/pools` and it
now answers HTTP 410. `Pools.ListByDex` keeps the same signature but sends the DEX
as the `dex_name` filter on `/networks/{network}/pools/search`. Two consequences
for callers:

- Pagination is cursor-based. `Page` is ignored, use `Cursor` and read
  `HasNextPage` / `NextCursor`.
- The response has no bare `volume_usd`. Read `VolumeUSD24h`. The SDK copies it
  into the deprecated `VolumeUSD` field so older code keeps working, but
  `Transactions` and the `LastPriceChangeUSD*` fields are no longer sent at all;
  use `Transactions24h` and the `PriceChangePercentage*` fields.

Despite its name, `dex_name` matches the DEX id, case-insensitively: `curve`,
`CURVE` and `uniswap_v3` all work. It does not match the human display name.
Pass `Dex.ID` from `Networks.ListDexes`, never `Dex.Name`: a display name such as
`Uniswap V3` returns HTTP 200 with an empty result set instead of an error, so
the mistake looks like a DEX with no pools.

### Tokens

```go
// Get details about a specific token
tokenDetails, err := client.Tokens.GetDetails(ctx, "ethereum", "0xtoken_address")

// Get pools that contain a specific token.
// This targets /networks/{network}/pools/search with token_address, so the
// filter is network-scoped and results are cursor-paginated: pass
// Cursor (from a response's NextCursor) to fetch the next page.
tokenPools, err := client.Tokens.GetPools(ctx, "ethereum", "0xtoken_address", &dexpaprika.TokenPoolsOptions{
    ListOptions: &dexpaprika.ListOptions{
        Limit:   10,
        OrderBy: "volume_usd_24h",
        Sort:    "desc",
    },
})
```

Note: the removed token-pools endpoint supported pair queries (a second token
address) and metric reordering (`reorder`). `/pools/search` has no equivalent
for either, so `AdditionalTokenAddress` and `Reorder` are deprecated and no
longer sent. To restrict results to a pair, filter the returned pools
client-side by their `Tokens` field.

### Pool Filtering

`Pools.Filter` targets the unified `/pools/search` endpoint. It is cursor-paginated
(`Cursor` / `NextCursor`), and legacy sort values are normalized automatically.

```go
// Find high-volume pools on Ethereum
vol := 100000.0
filtered, err := client.Pools.Filter(ctx, "ethereum", &dexpaprika.PoolFilterOptions{
    Limit:        10,
    Volume24hMin: &vol,
    SortBy:       "volume_usd_24h",
    SortDir:      "desc",
})
fmt.Printf("Found %d pools matching criteria (more pages: %v)\n", len(filtered.Results), filtered.HasNextPage)
```

### Top Tokens & Token Filtering

`Tokens.GetTop` and `Tokens.Filter` target the unified `/tokens/search` endpoint and
return the flat row shape (`FilteredToken`): address, chain, price, volume, liquidity,
FDV, transactions, and 24h price change. The API no longer returns token name/symbol
here. Both are cursor-paginated.

```go
// Get top tokens by volume
topTokens, err := client.Tokens.GetTop(ctx, "ethereum", &dexpaprika.TopTokensOptions{
    Limit:   10,
    OrderBy: "volume_usd_24h",
})

// Filter tokens by criteria
vol := 100000.0
fdv := 1000000.0
filteredTokens, err := client.Tokens.Filter(ctx, "ethereum", &dexpaprika.TokenFilterOptions{
    Limit:        10,
    Volume24hMin: &vol,
    FDVMin:       &fdv,
})
```

### Batch Token Prices

```go
// Get prices for multiple tokens in one request (max 10)
prices, err := client.Tokens.GetMultiPrices(ctx, "ethereum", []string{
    "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", // WETH
    "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", // USDC
})
for _, p := range prices {
    fmt.Printf("%s: $%.4f\n", p.ID, *p.PriceUSD)
}
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

## Versioning

This SDK follows [Semantic Versioning](https://semver.org/). 

- **Major version** changes indicate breaking API changes
- **Minor version** changes add functionality in a backwards-compatible manner
- **Patch version** changes fix bugs without changing the API

See the [CHANGELOG.md](CHANGELOG.md) file for a detailed version history.

## Resources

- [Official Documentation](https://docs.dexpaprika.com) - Comprehensive API reference
- [DexPaprika Website](https://dexpaprika.com) - Main product website
- [CoinPaprika](https://coinpaprika.com) - Related cryptocurrency data platform
- [Discord Community](https://discord.gg/DhJge5TUGM) - Get support and connect with other developers

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Development and Contribution

### Dependency Management

This project uses [Dependabot](https://github.com/features/security) to keep dependencies up to date. Dependabot automatically creates pull requests to update dependencies when new versions are available.

The Dependabot configuration can be found in `.github/dependabot.yml` and includes:
- Weekly updates for Go modules
- Weekly updates for GitHub Actions workflows
- Automatic grouping of minor and patch updates

### Continuous Integration

The project uses GitHub Actions for continuous integration:

1. **Go CI Workflow**: Runs on every push and pull request to main branches
   - Builds and tests the code on multiple Go versions
   - Runs linting checks
   - Generates and uploads code coverage reports

2. **Security Scanning**: Automatically scans for security vulnerabilities
   - Uses Gosec to identify security issues in the code
   - Runs govulncheck to check for vulnerabilities in dependencies
   - Performs dependency review on pull requests

## License

This project is licensed under the MIT License. 
