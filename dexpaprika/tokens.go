package dexpaprika

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// TokensService handles communication with the tokens related
// methods of the DexPaprika API.
type TokensService struct {
	client *Client
}

// TokenSummary contains token summary metrics.
type TokenSummary struct {
	PriceUSD     float64              `json:"price_usd"`
	FDV          float64              `json:"fdv"`
	LiquidityUSD float64              `json:"liquidity_usd"`
	Pools        *int                 `json:"pools,omitempty"`
	Day          *TimeIntervalMetrics `json:"24h,omitempty"`
	Hour6        *TimeIntervalMetrics `json:"6h,omitempty"`
	Hour1        *TimeIntervalMetrics `json:"1h,omitempty"`
	Minute30     *TimeIntervalMetrics `json:"30m,omitempty"`
	Minute15     *TimeIntervalMetrics `json:"15m,omitempty"`
	Minute5      *TimeIntervalMetrics `json:"5m,omitempty"`
	Minute1      *TimeIntervalMetrics `json:"1m,omitempty"`
}

// TokenDetails represents detailed information about a token.
type TokenDetails struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Symbol      string        `json:"symbol"`
	Chain       string        `json:"chain"`
	Decimals    int           `json:"decimals"`
	TotalSupply float64       `json:"total_supply"`
	Description string        `json:"description"`
	Website     string        `json:"website"`
	Telegram    string        `json:"telegram"`
	Twitter     string        `json:"twitter"`
	Explorer    string        `json:"explorer"`
	AddedAt     string        `json:"added_at"`
	Summary     *TokenSummary `json:"summary,omitempty"`
	LastUpdated string        `json:"last_updated"` // RFC3339/ISO8601 date-time format when token data was last updated
}

// GetDetails returns detailed information about a specific token on a network.
// Implements the getTokenDetails operation from the OpenAPI spec.
func (s *TokensService) GetDetails(ctx context.Context, networkID, tokenAddress string) (*TokenDetails, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}

	path := fmt.Sprintf("/networks/%s/tokens/%s", networkID, tokenAddress)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response TokenDetails
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

// TokenPoolsOptions contains options for retrieving token pools.
//
// Page is accepted for backward compatibility but is not sent to the
// cursor-paginated /pools/search endpoint; use ListOptions.Cursor to page.
type TokenPoolsOptions struct {
	*ListOptions
	// AdditionalTokenAddress used to filter pools that also contain a second
	// token (pair queries) on the removed token-pools endpoint.
	//
	// Deprecated: /pools/search has no pair filter and its token_address
	// parameter is last-wins when repeated, so this option is no longer sent.
	// Filter the returned pools client-side to match a pair.
	AdditionalTokenAddress string
	// Reorder used to flip the pool's pair perspective so the requested token
	// became the primary token for all metrics.
	//
	// Deprecated: /pools/search has no equivalent, so this option is no longer
	// sent. Metrics are returned from the pool's own perspective.
	Reorder bool
}

// GetPools returns a list of top liquidity pools that contain a specific token
// on a network.
//
// Since the DexPaprika API removed /networks/{network}/tokens/{address}/pools
// (HTTP 410), this method targets the unified /networks/{network}/pools/search
// endpoint with its token_address parameter. The filter is network-scoped
// only: the cross-network /pools/search endpoint accepts token_address but
// silently ignores it, so a network is always required here. The endpoint is
// cursor-paginated and rejects legacy sort values, so the sort field is
// normalized and the Page option is not sent upstream; pass ListOptions.Cursor
// (taken from a previous response's NextCursor) to page. An unknown token
// address returns an empty result set, not an error.
func (s *TokensService) GetPools(ctx context.Context, networkID, tokenAddress string, opts *TokenPoolsOptions) (*PoolsResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}

	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("/networks/%s/pools/search", networkID), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("token_address", tokenAddress)
	orderBy := ""
	if opts != nil && opts.ListOptions != nil {
		orderBy = opts.OrderBy
	}
	q.Add("order_by", mapPoolSortField(orderBy))
	if opts != nil && opts.ListOptions != nil {
		if opts.Limit > 0 {
			limit := opts.Limit
			if limit > 100 {
				limit = 100
			}
			q.Add("limit", fmt.Sprintf("%d", limit))
		}
		if opts.Sort != "" {
			q.Add("sort", opts.Sort)
		}
		if opts.Cursor != "" {
			q.Add("cursor", opts.Cursor)
		}
	}
	req.URL.RawQuery = q.Encode()

	var response PoolsResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// /pools/search returns rows under "results"; expose them via .Pools.
	if len(response.Pools) == 0 {
		response.Pools = response.Results
	}

	return &response, nil
}

// TopToken is an alias for FilteredToken, kept for backward compatibility.
// Since the API removed /networks/{network}/tokens/top, GetTop now returns the
// flat /tokens/search row shape (FilteredToken). The legacy name/symbol and
// nested timeframe metrics are no longer returned by the API.
type TopToken = FilteredToken

// TopTokensResponse represents the response from the /tokens/search endpoint.
type TopTokensResponse struct {
	Tokens      []TopToken `json:"results"`
	HasNextPage bool       `json:"has_next_page"`
	NextCursor  *string    `json:"next_cursor,omitempty"`
}

// TopTokensOptions contains options for ranking tokens.
//
// Page is accepted for backward compatibility but is not sent to the
// cursor-paginated /tokens/search endpoint; use Cursor to page instead.
type TopTokensOptions struct {
	Page    int
	Limit   int
	OrderBy string
	Sort    string
	Cursor  string
}

// GetTop returns top tokens on a network ranked by volume, liquidity, FDV, or
// other metrics.
//
// Since /networks/{network}/tokens/top was removed (HTTP 410), this targets the
// unified /networks/{network}/tokens/search endpoint: the sort field is
// normalized to a canonical value and pagination is cursor-based (Page is
// ignored, use Cursor).
func (s *TokensService) GetTop(ctx context.Context, networkID string, opts *TopTokensOptions) (*TopTokensResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("/networks/%s/tokens/search", networkID), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	orderBy := ""
	if opts != nil {
		orderBy = opts.OrderBy
	}
	q.Add("order_by", mapTokenSortField(orderBy))
	if opts != nil {
		if opts.Limit > 0 {
			limit := opts.Limit
			if limit > 100 {
				limit = 100
			}
			q.Add("limit", fmt.Sprintf("%d", limit))
		}
		if opts.Sort != "" {
			q.Add("sort", opts.Sort)
		}
		if opts.Cursor != "" {
			q.Add("cursor", opts.Cursor)
		}
	}
	req.URL.RawQuery = q.Encode()

	var response TopTokensResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

// FilteredToken represents a token row from the /tokens/search endpoint.
type FilteredToken struct {
	Chain                    string   `json:"chain"`
	Address                  string   `json:"address"`
	PriceUSD                 *float64 `json:"price_usd,omitempty"`
	VolumeUSD24h             *float64 `json:"volume_usd_24h,omitempty"`
	VolumeUSD7d              *float64 `json:"volume_usd_7d,omitempty"`
	VolumeUSD30d             *float64 `json:"volume_usd_30d,omitempty"`
	LiquidityUSD             *float64 `json:"liquidity_usd,omitempty"`
	FDVUSD                   *float64 `json:"fdv_usd,omitempty"`
	Txns24h                  *int     `json:"txns_24h,omitempty"`
	PriceChangePercentage24h *float64 `json:"price_change_percentage_24h,omitempty"`
	CreatedAt                string   `json:"created_at,omitempty"`
}

// TokenFilterResponse represents the response from the /tokens/search endpoint.
type TokenFilterResponse struct {
	Results     []FilteredToken `json:"results"`
	HasNextPage bool            `json:"has_next_page"`
	NextCursor  *string         `json:"next_cursor,omitempty"`
}

// TokenFilterOptions contains options for filtering tokens on a network.
//
// Page is accepted for backward compatibility but is not sent to the
// cursor-paginated /tokens/search endpoint; use Cursor to page instead.
type TokenFilterOptions struct {
	Page          int
	Limit         int
	SortBy        string
	SortDir       string
	Cursor        string
	Volume24hMin  *float64
	Volume24hMax  *float64
	LiquidityMin  *float64
	FDVMin        *float64
	FDVMax        *float64
	Txns24hMin    *int
	CreatedAfter  string
	CreatedBefore string
}

// Filter returns tokens on a network filtered by volume, liquidity, FDV, transactions, and creation date.
//
// Since /networks/{network}/tokens/filter was removed (HTTP 410), this targets
// the unified /networks/{network}/tokens/search endpoint: filters use canonical
// query parameter names, sorting uses order_by (normalized) plus sort
// (direction), and pagination is cursor-based (Page is ignored, use Cursor).
func (s *TokensService) Filter(ctx context.Context, networkID string, opts *TokenFilterOptions) (*TokenFilterResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("/networks/%s/tokens/search", networkID), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	sortBy := ""
	if opts != nil {
		sortBy = opts.SortBy
	}
	q.Add("order_by", mapTokenSortField(sortBy))
	if opts != nil {
		if opts.Limit > 0 {
			limit := opts.Limit
			if limit > 100 {
				limit = 100
			}
			q.Add("limit", fmt.Sprintf("%d", limit))
		}
		if opts.SortDir != "" {
			q.Add("sort", opts.SortDir)
		}
		if opts.Cursor != "" {
			q.Add("cursor", opts.Cursor)
		}
		if opts.Volume24hMin != nil {
			q.Add("volume_usd_24h_min", fmt.Sprintf("%f", *opts.Volume24hMin))
		}
		if opts.Volume24hMax != nil {
			q.Add("volume_usd_24h_max", fmt.Sprintf("%f", *opts.Volume24hMax))
		}
		if opts.LiquidityMin != nil {
			q.Add("liquidity_usd_min", fmt.Sprintf("%f", *opts.LiquidityMin))
		}
		if opts.FDVMin != nil {
			q.Add("fdv_min", fmt.Sprintf("%f", *opts.FDVMin))
		}
		if opts.FDVMax != nil {
			q.Add("fdv_max", fmt.Sprintf("%f", *opts.FDVMax))
		}
		if opts.Txns24hMin != nil {
			q.Add("txns_24h_min", fmt.Sprintf("%d", *opts.Txns24hMin))
		}
		if opts.CreatedAfter != "" {
			q.Add("created_after", opts.CreatedAfter)
		}
		if opts.CreatedBefore != "" {
			q.Add("created_before", opts.CreatedBefore)
		}
	}
	req.URL.RawQuery = q.Encode()

	var response TokenFilterResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

// TokenPrice represents a token price from the multi-prices endpoint.
type TokenPrice struct {
	Chain    string   `json:"chain"`
	ID       string   `json:"id"`
	PriceUSD *float64 `json:"price_usd,omitempty"`
}

// GetMultiPrices returns batch prices for multiple tokens on a network.
// The tokens parameter accepts up to 10 token addresses.
func (s *TokensService) GetMultiPrices(ctx context.Context, networkID string, tokens []string) ([]TokenPrice, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("tokens list is required and must not be empty")
	}
	if len(tokens) > 10 {
		return nil, fmt.Errorf("tokens list must contain at most 10 addresses")
	}

	path := fmt.Sprintf("/networks/%s/multi/prices", networkID)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("tokens", strings.Join(tokens, ","))
	req.URL.RawQuery = q.Encode()

	var response []TokenPrice
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return response, nil
}
