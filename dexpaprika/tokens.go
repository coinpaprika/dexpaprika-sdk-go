package dexpaprika

import (
	"context"
	"fmt"
	"net/http"
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
	Explorer    string        `json:"explorer"`
	AddedAt     string        `json:"added_at"`
	Summary     *TokenSummary `json:"summary,omitempty"`
	LastUpdated string        `json:"last_updated"` // RFC3339/ISO8601 date-time format when token data was last updated
}

// GetDetails returns detailed information about a specific token on a network.
// Implements the getTokenDetails operation from the OpenAPI spec.
func (s *TokensService) GetDetails(ctx context.Context, networkID, tokenAddress string) (*TokenDetails, error) {
	path := fmt.Sprintf("/networks/%s/tokens/%s", networkID, tokenAddress)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response TokenDetails
	_, err = s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetPools returns a list of top liquidity pools for a specific token on a network.
// Implements the getTokenPools operation from the OpenAPI spec.
func (s *TokensService) GetPools(ctx context.Context, networkID, tokenAddress string, opts *ListOptions, additionalTokenAddress string) (*PoolsResponse, error) {
	path := fmt.Sprintf("/networks/%s/tokens/%s/pools", networkID, tokenAddress)

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
			q.Add("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Sort != "" {
			q.Add("sort", opts.Sort)
		}
		if opts.OrderBy != "" {
			q.Add("order_by", opts.OrderBy)
		}
	}
	if additionalTokenAddress != "" {
		q.Add("address", additionalTokenAddress)
	}
	req.URL.RawQuery = q.Encode()

	var response PoolsResponse
	_, err = s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
