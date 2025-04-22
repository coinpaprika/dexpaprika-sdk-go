# Changelog

## [1.2.0] - 2025-04-22

### Changed
- Corrected Dex struct JSON field mapping to match API response format
- Improved reliability of API tests with proper error handling
- Enhanced test coverage with more comprehensive assertions

### Added
- Added proper Go idiomatic tests using the standard testing package
- Added MIT license
- Added GitHub Actions workflow for CI/CD
- Added golangci-lint configuration for code quality
- Added status badge to README.md for build status

### Fixed
- Fixed field mappings in Dex struct to correctly parse API responses
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