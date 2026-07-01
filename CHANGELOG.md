# Changelog

## [1.5.0] - 2026-06-30

### Breaking Changes
- **API CHANGE**: DexPaprika removed four REST endpoints (now HTTP 410) and replaced them with the unified search endpoints:
  - `GET /networks/{network}/pools` and `GET /networks/{network}/pools/filter` -> `GET /networks/{network}/pools/search`
  - `GET /networks/{network}/tokens/top` and `GET /networks/{network}/tokens/filter` -> `GET /networks/{network}/tokens/search`
- `Pools.ListByNetwork()`, `Pools.Filter()`, `Tokens.GetTop()`, and `Tokens.Filter()` now target the search endpoints. Method signatures are unchanged.
- The search endpoints are cursor-paginated and do not accept `page`. The `Page` option is still accepted for source compatibility but is no longer sent; use the new `Cursor` option (read from a response's `NextCursor`) to page.
- Sorting now uses `order_by` (field) plus `sort` (direction). Filter methods no longer send `sort_by`/`sort_dir`. Legacy sort values and filter parameter names are mapped to canonical values automatically (the search endpoints reject legacy values with HTTP 400).
- `Tokens.GetTop()` now returns the flat search row shape. `TopToken` is now an alias for `FilteredToken`; the legacy `name`, `symbol`, pool count, and nested timeframe metrics are no longer returned by the API. `TopTokenTimeMetrics` was removed.
- `TokenFilterResponse` rows are now read from `results` (previously `data`).

### Changed
- Response types now expose cursor pagination: `PoolsResponse`, `PoolFilterResponse`, `TopTokensResponse`, and `TokenFilterResponse` carry `HasNextPage` and `NextCursor`. `ListByNetwork` exposes search rows via the existing `.Pools` field for backward compatibility.
- Extended `Pool` with the search item fields `Transactions24h` and `PriceChangePercentage5m/1h/24h`.
- Added `Cursor` to `ListOptions`, `PoolFilterOptions`, `TopTokensOptions`, and `TokenFilterOptions`.

## [1.4.0] - 2026-03-31

### Added
- **Pool filtering**: `Pools.Filter()` method for advanced pool filtering by volume, liquidity, transactions, and creation date
- **Top tokens**: `Tokens.GetTop()` method for discovering top tokens on a network ranked by volume, price, liquidity, or other metrics
- **Token filtering**: `Tokens.Filter()` method for filtering tokens by volume, liquidity, FDV, transactions, and creation date
- **Batch prices**: `Tokens.GetMultiPrices()` method for getting prices of up to 10 tokens in a single request
- New types: `PoolFilterOptions`, `PoolFilterResponse`, `TopToken`, `TopTokenTimeMetrics`, `TopTokensResponse`, `TopTokensOptions`, `FilteredToken`, `TokenFilterResponse`, `TokenFilterOptions`, `TokenPrice`
- Extended `Token` struct with `TotalSupply`, `Description`, `Website`, `Type`, `Status`, `HasImage` fields
- Extended `Pool` struct with `VolumeUSD7d`, `LiquidityUSD` fields
- Test coverage for all new endpoints

### Changed
- Pool price change fields (`LastPriceChangeUSD5m/1h/24h`, `Fee`) are now pointer types to handle null API responses
- Updated SDK version to 1.4.0

## [1.3.0] - 2025-01-27

### Breaking Changes
- **DEPRECATED**: `Pools.List()` method due to API endpoint removal (returns 410 Gone)
- **REQUIRED**: All pool operations now require network parameter
- **MIGRATION**: Update `client.Pools.List(ctx, opts)` calls to `client.Pools.ListByNetwork(ctx, network, opts)`
- **API CHANGE**: Updated to DexPaprika API v1.3.0 with network-specific endpoints

### Added
- Network parameter validation for all pool and token methods
- Improved error handling for 410 Gone responses with migration guidance
- Support for new token pools parameters: `reorder` and `address` filtering
- Enhanced parameter validation with automatic limit constraints (max 100 for most endpoints, max 366 for OHLCV)
- New `TokenPoolsOptions` struct for better token pool configuration
- Comprehensive validation tests for all new parameter requirements

### Changed
- `Tokens.GetPools()` method signature updated to use `TokenPoolsOptions` struct
- All network-related methods now validate network ID parameter
- All pool-related methods now validate pool address parameter
- Limit parameters automatically capped at API maximums (100 for pools, 366 for OHLCV)
- Enhanced error messages for deprecated endpoints with migration examples

### Migration Guide
```go
// Before (deprecated):
pools, err := client.Pools.List(ctx, &dexpaprika.ListOptions{Limit: 10})

// After (required):
pools, err := client.Pools.ListByNetwork(ctx, "ethereum", &dexpaprika.ListOptions{Limit: 10})
pools, err := client.Pools.ListByNetwork(ctx, "solana", &dexpaprika.ListOptions{Limit: 10})

// Token pools before:
pools, err := client.Tokens.GetPools(ctx, network, token, opts, additionalToken)

// Token pools after:
pools, err := client.Tokens.GetPools(ctx, network, token, &dexpaprika.TokenPoolsOptions{
    ListOptions: opts,
    AdditionalTokenAddress: additionalToken,
    Reorder: false,
})
```

## [1.2.0] - 2025-04-22

### Changed
- Corrected Dex struct JSON field mapping to match API response format
- Improved reliability of API tests with proper error handling
- Enhanced test coverage from 63.3% to 83.6% with comprehensive test suite

### Added
- Implemented dual testing strategy with mock-based comprehensive tests and actual API e2e tests
- Added extensive unit tests for utils, search, cache, and pagination services
- Added tests for error handling, timeouts, and edge cases
- Added MIT license
- Added GitHub Actions workflow for CI/CD
- Added golangci-lint configuration for code quality
- Added status badge to README.md for build status

### Fixed
- Fixed linter errors in search_test.go related to client initialization
- Fixed method call to client.Tokens.GetPools by adding missing parameter
- Removed redundant stable_test.go as functionality is covered by other tests
- Fixed OHLCV tests with proper date formatting

## [1.1.0] - 2025-04-15

### Changed
- Updated the SDK to align with OpenAPI 3.1.0 specification
- Added operationId references to all API methods for better traceability
- Updated TokenDetails.LastUpdated field documentation to indicate date-time format
- Improved code documentation
- Enhanced API error reporting

### Added
- Added support for explicit HTTP error handling for 400 and 500 responses
- Added CHANGELOG.md for tracking version changes

## [1.0.0] - 2025-03-10

### Added
- Initial release of the DexPaprika Go SDK
- Complete support for all DexPaprika API endpoints
- Caching layer for improved performance
- Pagination helpers for all collection endpoints
- Comprehensive error handling
- Production-ready client with retry and rate limiting 