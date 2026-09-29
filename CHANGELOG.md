# Changelog

## [1.11.0] - 2026-09-29

Token OHLCV: USD candles for a token across every pool it trades in, not just one pool.

### Added
- `Tokens.GetOHLCV(ctx, networkID, tokenAddress, opts)` calls `GET /networks/{network}/tokens/{token_address}/ohlcv`. Each candle is a volume-weighted USD price across every pool the token trades in on that network, and `Volume` is the USD traded across all of them combined. It decodes into the same `OHLCVRecord` that `Pools.GetOHLCV` already returns.
- `TokenOHLCVOptions` carries `Start` (required), `End`, `Interval` and `Limit`, the same shapes as `OHLCVOptions` on pool OHLCV. It has no `Inversed` field: this endpoint has no such parameter, so it gets its own options type rather than reusing `OHLCVOptions` and risking that field being sent by mistake.
- This endpoint requires a Dev, Pro or Enterprise plan. A keyless or free-key call gets a `403` whose body names the required plan; `GetOHLCV` returns that message unchanged in `*APIError.Message`, matched with `errors.Is(err, ErrForbidden)`. On the Dev plan, history is limited to the last 30 days.
- Reaching it needs the same client setup Pro users already use for pool OHLCV: `WithAPIKey` plus `WithBaseURL("https://api-pro.dexpaprika.com")`. The SDK does not switch host on its own, so a client left on the default host gets the same 403 regardless of which key is attached.

## [1.10.1] - 2026-09-28

`GetTransactions` and `GetMultiPrices` keep every field the API sends, and the test suite stays under the API rate limit.

### Fixed
- `TokenPrice` has a `LastUpdated` field. The multi prices endpoint returns `last_updated` (RFC 3339, for example `"2026-09-28T13:06:30Z"`) next to every price, and the struct had nowhere to put it, so it was dropped on decode.
- `Transaction` has the 13 fields the transactions endpoint returns and the struct dropped: `Chain`, `FactoryID`, `Token0Symbol`, `Token1Symbol`, `Volume0`, `Volume1`, `Price0`, `Price1`, `Price0USD`, `Price1USD`, `CreatedAt`, `CreatedAtBlockHash` and `CanonicalChain`. Before this a row had no time other than the block number and no USD value. Numbers and the bool are pointers and the strings use `omitempty`, like the other optional fields. `Amount0` and `Amount1` still decode as `float64`; use `Volume0` and `Volume1` for display.

### Tests and CI
- Live tests are spaced to the per-IP rate limit. They used to send their requests back to back; the burst ran into 429 and whichever test was running when its retries or its 10 second context ran out failed, a different one each run (seen in CI: `TestTokensGetMultiPrices`, `TestPoolsPaginator_ForDex`). A test-only transport in `internal/livethrottle` now starts requests to `api.dexpaprika.com` at least 4.2 s apart, or 2.1 s apart when `DEXPAPRIKA_API_KEY` is set. Requests to httptest servers are not delayed.
- Live tests get a 3 minute deadline instead of 5 to 60 seconds, because the spacing counts against it.
- `make test` runs the packages one at a time (`-p 1`), since two test binaries at full pace would exceed a per-IP limit, and allows 20 minutes.
- The Test workflow passes an optional `DEXPAPRIKA_API_KEY` secret. Without it, and on pull requests from forks, the suite runs keyless.
- A full keyless run made 60 live requests (47 in `dexpaprika`, 13 in `tests`) with no 429 and took 4 min 15 s. Unpaced, the suite took 1.5 to 2 minutes, and 2 of the 4 CI runs between 25 and 28 September failed on a 429 or a context deadline.

### Notes
- A new test runs a live transactions row through `GetTransactions`, encodes it again and compares the keys with what the API sent: 24 of 24 here, 11 against 1.10.0.
- 1 new test runs the live multi prices body through `GetMultiPrices` and encodes the result again. Against 1.10.0 it fails with `last_updated is <nil> after decoding`. 4 tests cover the throttle: spacing, no delay for other hosts, a cancelled context ends the wait, and the pace follows the key.

## [1.10.0] - 2026-09-28

Time filters on transactions and search take relative times.

### Added
- `WithFrom(string)` and `WithTo(string)` for `Pools.GetTransactions`. They take a relative offset from now such as `"-1h"` or `"-24h"`, RFC 3339, `YYYY-MM-DD` or Unix seconds, which the API accepts since 2026-09-28. `WithFromTimestamp(int64)` and `WithToTimestamp(int64)` could not carry `-1h`; they stay and keep working unchanged.
- `CreatedAfter` and `CreatedBefore` on `PoolFilterOptions` and `TokenFilterOptions` were already strings; their documentation now lists the same shapes, so `CreatedAfter: "-24h"` returns what was created in the last day.

## [1.9.0] - 2026-09-25

OHLCV availability now depends on your plan. `GetOHLCV` keeps its signature; it stops cutting `Limit` to 366, and the documentation and examples now work without a key.

### API changes this release documents
- **OHLCV history depth and candle interval are per plan since 2026-09-25.** Without a key: the last 24 hours at `1h`, `6h`, `12h` and `24h`. Free key: 7 days at `10m` and longer (`1m` and `5m` are paid). Dev: 30 days at every interval. Pro and Enterprise: unlimited. A `Start` or `End` outside the window, or a finer interval than the plan allows, is answered with `403`; `GetOHLCV` returns an `*APIError` that matches `errors.Is(err, ErrForbidden)` and whose `Message` names the plan that lifts the limit. See [OHLCV limits by plan](https://docs.dexpaprika.com/knowledge-base/rate-limits#ohlcv-limits-by-plan).
- **`Start` and `End` accept a relative offset from now:** `-24h`, `-7d`, `-90m`, `-30s`. `Start: "-24h"` selects the last 24 hours, which every plan may query. The fields are strings and the SDK sends them unchanged, so this already worked in 1.8.0.
- A missing or malformed `Start` or `End` is answered with `400`.

### Changed
- **`OHLCVOptions.Limit` is capped at 1000, the API maximum.** The SDK used to lower every value above 366 to 366 without saying so, so a request for 500 hourly candles quietly returned 366.
- `OHLCVOptions` fields and `GetOHLCV` carry doc comments on the accepted formats and the per-plan window.
- The README example asked for `2023-01-01` to `2023-01-31`, which now returns 403 without a key. It uses `Start: "-24h"` with hourly candles.
- `examples/production_usage.go`, `tests/e2e_test.go` and the live `TestPools_GetOHLCV` requested yesterday's date, which the API reads as midnight UTC and therefore more than 24 hours back, so all three were answered with 403. They request `-24h`.

### Notes
- 2 new tests against an httptest server: `Limit` 1000 reaches the wire as 1000 and 5000 is capped at 1000, and `Start`/`End` relative offsets pass through unchanged. Against 1.8.0 the first fails with `limit 1000 sent as "366"`.

## [1.8.0] - 2026-08-14

### Added
- **Optional API key.** `dexpaprika.WithAPIKey("api_...")`, falling back to the `DEXPAPRIKA_API_KEY` environment variable when the option is not passed. Keyless remains the default and is unchanged: without a key the client sends exactly what it sent before. The key is transmitted as the **entire** `Authorization` value, with no `Bearer` prefix and no other scheme word, because the API checksums the raw header and a scheme word returns 401.
- `Version` and `APIKeyEnvVar` are now exported constants.
- The host is never inferred from the presence of a key. Free keys are served from `DefaultBaseURL` and only Pro moves to `api-pro.dexpaprika.com`, selected with `WithBaseURL`. Sending a free key to the Pro host returns 403, so guessing would break exactly the people who just registered.

### Changed
- **The default User-Agent carries the version.** It was the bare string `DexPaprika-SDK-Go`, which said the SDK was in use but never which version, so no rollout could be measured. It is now `DexPaprika-SDK-Go/<Version>`. `WithUserAgent` and `SetUserAgent` still override it, and the two existing tests that pinned the unversioned literal were updated.

### Notes
- 12 new tests, asserting on the headers that actually reach an httptest server rather than on struct fields: the bare-key format against five scheme words, keyless behaviour, option-beats-environment precedence, whitespace trimming, rejection of keys carrying header-injection characters, and the host rules in both directions.
- A key the API cannot read is ignored rather than rejected on the data endpoints: the call returns `200` with real data while quietly serving the keyless tier. `/usage` and its `plan` field are the way to confirm a key is landing.

## [1.7.0] - 2026-08-14

### Breaking Changes
- **API CHANGE**: DexPaprika removed `GET /networks/{network}/dexes/{dex}/pools` (now HTTP 410). `Pools.ListByDex()` now targets `GET /networks/{network}/pools/search` and sends the DEX as the `dex_name` query parameter. The method signature is unchanged.
- Despite its name, `dex_name` matches the DEX **id** (case-insensitively), which is what `Networks.ListDexes` returns as `Dex.ID`. Passing a human display name such as `Uniswap V3` (the `Dex.Name` field) returns HTTP 200 with an empty result set instead of an error, so a wrong value here fails silently.
- Pagination is cursor-based: `ListOptions.Page` is still accepted for source compatibility but is no longer sent; use `ListOptions.Cursor` (read from a response's `NextCursor`) to page. Sort fields are normalized to the canonical 24h names, since the endpoint rejects legacy values with HTTP 400.
- `PoolsPaginator.ForDex` pages DEX pools by cursor instead of by page number.
- Rows arrive under `results` rather than `pools`, and there is no `page_info`. `PoolsResponse.PageInfo` is deprecated and stays at its zero value for every pools listing.

### Changed
- `Pool.VolumeUSD` is deprecated. No pools endpoint returns a bare `volume_usd` any more; the SDK copies `volume_usd_24h` into it so existing callers keep reading a real number. Read `Pool.VolumeUSD24h` in new code.
- `Pool.Transactions` and the `Pool.LastPriceChangeUSD*` fields are deprecated and no longer populated. Use `Pool.Transactions24h` and the `Pool.PriceChangePercentage*` fields.

### Added
- `Pool.PriceChangePercentage6h`, which `/pools/search` returns but the SDK was dropping.

## [1.6.0] - 2026-07-15

### Breaking Changes
- **API CHANGE**: DexPaprika removed `GET /networks/{network}/tokens/{address}/pools` (now HTTP 410). `Tokens.GetPools()` now targets `GET /networks/{network}/pools/search` with its new `token_address` parameter. The method signature is unchanged.
- The token filter is network-scoped only: the cross-network `/pools/search` endpoint accepts `token_address` but silently ignores it, so `GetPools` still requires a network.
- Pagination is cursor-based: `Page` is still accepted for source compatibility but is no longer sent; use `ListOptions.Cursor` (read from a response's `NextCursor`) to page. Sort fields are normalized to the canonical 24h names (legacy values are rejected by the endpoint with HTTP 400).
- `TokenPoolsOptions.AdditionalTokenAddress` (pair queries) and `TokenPoolsOptions.Reorder` (pair-perspective flip) are deprecated and no longer sent: `/pools/search` has no equivalent for either, and repeating `token_address` is last-wins on the API side, not a pair filter. Filter pools client-side by their `Tokens` field to match a pair.
- An unknown token address now returns HTTP 200 with an empty result set instead of an error.
- `PoolsPaginator.ForToken` pages token pools by cursor; its `secondToken` argument is deprecated and ignored.

## [1.5.1] - 2026-07-01

### Changed
- **Deprecation errors surface the replacement**: `APIError` now has a `Replacement` field, populated from the API error body's `replacement`, and the error message includes `Use <replacement> instead.`. Also fixes a bug where the API's `message` was dropped when the body had no `error` key. `ErrGone` still matches via `errors.Is`. Generic across any error status carrying a `replacement`.

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