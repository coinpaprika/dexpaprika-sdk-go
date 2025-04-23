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

- Go 1.24 or higher
- No API key required (service is in public beta)

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
    
    // Get top trading pools
    pools, err := client.Pools.List(ctx, &dexpaprika.ListOptions{
        Limit: 5,
        OrderBy: "volume_usd",
        Sort: "desc",
    })
    if err != nil {
        log.Fatalf("Error fetching pools: %v", err)
    }
    
    fmt.Println("\nTop trading pools:")
    for _, pool := range pools.Pools {
        fmt.Printf("  - %s on %s (Volume: $%.2f)\n", 
            pool.DexName, 
            pool.Chain, 
            pool.VolumeUSD)
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
        fmt.Printf("Pool: %s on %s (Volume: $%.2f)\n", 
            pool.DexName, 
            pool.Chain, 
            pool.VolumeUSD)
    }
}
```

## Handling Errors

The SDK provides detailed error types to help you handle different failure scenarios:

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

### Code Ownership

The project uses a CODEOWNERS file to automatically request reviews from the appropriate team members when a pull request is opened.

## License

This project is licensed under the MIT License. 
