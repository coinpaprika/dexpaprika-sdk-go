# Changelog

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