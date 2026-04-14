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
type TokenPoolsOptions struct {
	*ListOptions
	// AdditionalTokenAddress filters pools that contain this additional token address
	AdditionalTokenAddress string
	// Reorder reorders the pool so that the specified token becomes the primary token for all metrics
	Reorder bool
}

// GetPools returns a list of top liquidity pools for a specific token on a network.
// Implements the getTokenPools operation from the OpenAPI spec.
func (s *TokensService) GetPools(ctx context.Context, networkID, tokenAddress string, opts *TokenPoolsOptions) (*PoolsResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}

	path := fmt.Sprintf("/networks/%s/tokens/%s/pools", networkID, tokenAddress)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if opts != nil {
		if opts.ListOptions != nil {
			if opts.Page > 0 {
				q.Add("page", fmt.Sprintf("%d", opts.Page))
			}
			if opts.Limit > 0 {
				// Validate limit constraints (max 100 as per API v1.3.0)
				limit := opts.Limit
				if limit > 100 {
					limit = 100
				}
				q.Add("limit", fmt.Sprintf("%d", limit))
			}
			if opts.Sort != "" {
				q.Add("sort", opts.Sort)
			}
			if opts.OrderBy != "" {
				q.Add("order_by", opts.OrderBy)
			}
		}
		if opts.AdditionalTokenAddress != "" {
			q.Add("address", opts.AdditionalTokenAddress)
		}
		if opts.Reorder {
			q.Add("reorder", "true")
		}
	}
	req.URL.RawQuery = q.Encode()

	var response PoolsResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

// TopTokenTimeMetrics represents time interval metrics for top tokens.
type TopTokenTimeMetrics struct {
	VolumeUSD          float64  `json:"volume_usd"`
	Txns               int      `json:"txns"`
	LastPriceUSDChange *float64 `json:"last_price_usd_change,omitempty"`
	Buys               *int     `json:"buys,omitempty"`
	Sells              *int     `json:"sells,omitempty"`
}

// TopToken represents a token from the top tokens endpoint.
type TopToken struct {
	Address      string               `json:"address"`
	Name         string               `json:"name"`
	Symbol       string               `json:"symbol"`
	Chain        string               `json:"chain"`
	Decimals     int                  `json:"decimals"`
	HasImage     *bool                `json:"has_image,omitempty"`
	PriceUSD     *float64             `json:"price_usd,omitempty"`
	FDV          *float64             `json:"fdv,omitempty"`
	LiquidityUSD *float64             `json:"liquidity_usd,omitempty"`
	Pools        *int                 `json:"pools,omitempty"`
	Day          *TopTokenTimeMetrics `json:"24h,omitempty"`
	Hour1        *TopTokenTimeMetrics `json:"1h,omitempty"`
	Minute5      *TopTokenTimeMetrics `json:"5m,omitempty"`
}

// TopTokensResponse represents the response from the top tokens endpoint.
type TopTokensResponse struct {
	Tokens   []TopToken `json:"tokens"`
	PageInfo PageInfo   `json:"page_info"`
}

// TopTokensOptions contains options for the top tokens endpoint.
type TopTokensOptions struct {
	Page    int
	Limit   int
	OrderBy string
	Sort    string
}

// GetTop returns top tokens on a network ranked by volume, price, liquidity, or other metrics.
func (s *TokensService) GetTop(ctx context.Context, networkID string, opts *TopTokensOptions) (*TopTokensResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/networks/%s/tokens/top", networkID)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if opts != nil {
		if opts.Page > 0 {
			q.Add("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.Limit > 0 {
			limit := opts.Limit
			if limit > 100 {
				limit = 100
			}
			q.Add("limit", fmt.Sprintf("%d", limit))
		}
		if opts.OrderBy != "" {
			q.Add("order_by", opts.OrderBy)
		}
		if opts.Sort != "" {
			q.Add("sort", opts.Sort)
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

// FilteredToken represents a token from the token filter endpoint.
type FilteredToken struct {
	Chain        string   `json:"chain"`
	Address      string   `json:"address"`
	PriceUSD     *float64 `json:"price_usd,omitempty"`
	VolumeUSD24h *float64 `json:"volume_usd_24h,omitempty"`
	VolumeUSD7d  *float64 `json:"volume_usd_7d,omitempty"`
	LiquidityUSD *float64 `json:"liquidity_usd,omitempty"`
	FDVUSD       *float64 `json:"fdv_usd,omitempty"`
	Txns24h      *int     `json:"txns_24h,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
}

// TokenFilterResponse represents the response from the token filter endpoint.
type TokenFilterResponse struct {
	Results  []FilteredToken `json:"results"`
	PageInfo PageInfo        `json:"page_info"`
}

// TokenFilterOptions contains options for filtering tokens on a network.
type TokenFilterOptions struct {
	Page          int
	Limit         int
	SortBy        string
	SortDir       string
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
func (s *TokensService) Filter(ctx context.Context, networkID string, opts *TokenFilterOptions) (*TokenFilterResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/networks/%s/tokens/filter", networkID)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if opts != nil {
		if opts.Page > 0 {
			q.Add("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.Limit > 0 {
			limit := opts.Limit
			if limit > 100 {
				limit = 100
			}
			q.Add("limit", fmt.Sprintf("%d", limit))
		}
		if opts.SortBy != "" {
			q.Add("sort_by", opts.SortBy)
		}
		if opts.SortDir != "" {
			q.Add("sort_dir", opts.SortDir)
		}
		if opts.Volume24hMin != nil {
			q.Add("volume_24h_min", fmt.Sprintf("%f", *opts.Volume24hMin))
		}
		if opts.Volume24hMax != nil {
			q.Add("volume_24h_max", fmt.Sprintf("%f", *opts.Volume24hMax))
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
