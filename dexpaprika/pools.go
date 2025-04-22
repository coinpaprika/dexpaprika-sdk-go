package dexpaprika

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// PoolsService handles communication with the pools related
// methods of the DexPaprika API.
type PoolsService struct {
	client *Client
}

// Token represents a token in a pool.
type Token struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Symbol   string   `json:"symbol"`
	Chain    string   `json:"chain"`
	Decimals int      `json:"decimals"`
	AddedAt  string   `json:"added_at"`
	FDV      *float64 `json:"fdv,omitempty"`
}

// Pool represents a liquidity pool.
type Pool struct {
	ID                    string  `json:"id"`
	DexID                 string  `json:"dex_id"`
	DexName               string  `json:"dex_name"`
	Chain                 string  `json:"chain"`
	VolumeUSD             float64 `json:"volume_usd"`
	CreatedAt             string  `json:"created_at"`
	CreatedAtBlockNumber  int64   `json:"created_at_block_number"`
	Transactions          int     `json:"transactions"`
	PriceUSD              float64 `json:"price_usd"`
	LastPriceChangeUSD5m  float64 `json:"last_price_change_usd_5m"`
	LastPriceChangeUSD1h  float64 `json:"last_price_change_usd_1h"`
	LastPriceChangeUSD24h float64 `json:"last_price_change_usd_24h"`
	Fee                   float64 `json:"fee"`
	Tokens                []Token `json:"tokens"`
}

// PoolsResponse represents the response for the pools endpoint.
type PoolsResponse struct {
	Pools    []Pool   `json:"pools"`
	PageInfo PageInfo `json:"page_info"`
}

// ListOptions contains common options for listing pools.
type ListOptions struct {
	Page    int
	Limit   int
	Sort    string
	OrderBy string
}

// addOptions adds the parameters in opts as URL query parameters to s.
func addOptions(s string, opts interface{}) (string, error) {
	v := url.Values{}

	if o, ok := opts.(*ListOptions); ok {
		if o.Page > 0 {
			v.Add("page", fmt.Sprintf("%d", o.Page))
		}
		if o.Limit > 0 {
			v.Add("limit", fmt.Sprintf("%d", o.Limit))
		}
		if o.Sort != "" {
			v.Add("sort", o.Sort)
		}
		if o.OrderBy != "" {
			v.Add("order_by", o.OrderBy)
		}
	}

	if len(v) > 0 {
		return s + "?" + v.Encode(), nil
	}
	return s, nil
}

// List returns a list of top pools from all networks.
// Implements the getTopPools operation from the OpenAPI spec.
func (s *PoolsService) List(ctx context.Context, opts *ListOptions) (*PoolsResponse, error) {
	path, err := addOptions("/pools", opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response PoolsResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

// ListByNetwork returns a list of top pools on a specific network.
// Implements the getNetworkPools operation from the OpenAPI spec.
func (s *PoolsService) ListByNetwork(ctx context.Context, networkID string, opts *ListOptions) (*PoolsResponse, error) {
	path, err := addOptions(fmt.Sprintf("/networks/%s/pools", networkID), opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response PoolsResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

// ListByDex returns a list of top pools on a specific network's DEX.
// Implements the getDexPools operation from the OpenAPI spec.
func (s *PoolsService) ListByDex(ctx context.Context, networkID, dexID string, opts *ListOptions) (*PoolsResponse, error) {
	path, err := addOptions(fmt.Sprintf("/networks/%s/dexes/%s/pools", networkID, dexID), opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response PoolsResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

// TimeIntervalMetrics represents metrics for a specific time interval.
type TimeIntervalMetrics struct {
	LastPriceUSDChange float64 `json:"last_price_usd_change"`
	VolumeUSD          float64 `json:"volume_usd"`
	BuyUSD             float64 `json:"buy_usd"`
	SellUSD            float64 `json:"sell_usd"`
	Sells              int     `json:"sells"`
	Buys               int     `json:"buys"`
	Txns               int     `json:"txns"`
}

// PoolDetails represents detailed information about a pool.
type PoolDetails struct {
	ID                   string              `json:"id"`
	CreatedAtBlockNumber int64               `json:"created_at_block_number"`
	Chain                string              `json:"chain"`
	CreatedAt            string              `json:"created_at"`
	FactoryID            string              `json:"factory_id"`
	DexID                string              `json:"dex_id"`
	DexName              string              `json:"dex_name"`
	Tokens               []Token             `json:"tokens"`
	LastPrice            float64             `json:"last_price"`
	LastPriceUSD         float64             `json:"last_price_usd"`
	Fee                  float64             `json:"fee"`
	PriceTime            string              `json:"price_time"`
	Day                  TimeIntervalMetrics `json:"24h"`
	Hour6                TimeIntervalMetrics `json:"6h"`
	Hour1                TimeIntervalMetrics `json:"1h"`
	Minute30             TimeIntervalMetrics `json:"30m"`
	Minute15             TimeIntervalMetrics `json:"15m"`
	Minute5              TimeIntervalMetrics `json:"5m"`
}

// GetDetails returns details about a specific pool on a network.
// Implements the getPoolDetails operation from the OpenAPI spec.
func (s *PoolsService) GetDetails(ctx context.Context, networkID, poolAddress string, inversed bool) (*PoolDetails, error) {
	path := fmt.Sprintf("/networks/%s/pools/%s", networkID, poolAddress)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	if inversed {
		q := req.URL.Query()
		q.Add("inversed", "true")
		req.URL.RawQuery = q.Encode()
	}

	var response PoolDetails
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

// OHLCVRecord represents a single OHLCV (Open-High-Low-Close-Volume) data point.
type OHLCVRecord struct {
	TimeOpen  string  `json:"time_open"`
	TimeClose string  `json:"time_close"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
}

// OHLCVOptions contains options for retrieving OHLCV data.
type OHLCVOptions struct {
	Start    string
	End      string
	Limit    int
	Interval string
	Inversed bool
}

// GetOHLCV returns OHLCV data for a specific pool.
// Implements the getPoolOHLCV operation from the OpenAPI spec.
func (s *PoolsService) GetOHLCV(ctx context.Context, networkID, poolAddress string, opts *OHLCVOptions) ([]OHLCVRecord, error) {
	path := fmt.Sprintf("/networks/%s/pools/%s/ohlcv", networkID, poolAddress)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if opts != nil {
		if opts.Start != "" {
			q.Add("start", opts.Start)
		}
		if opts.End != "" {
			q.Add("end", opts.End)
		}
		if opts.Limit > 0 {
			q.Add("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Interval != "" {
			q.Add("interval", opts.Interval)
		}
		if opts.Inversed {
			q.Add("inversed", "true")
		}
	}
	req.URL.RawQuery = q.Encode()

	var response []OHLCVRecord
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return response, nil
}

// Transaction represents a transaction of a pool.
type Transaction struct {
	ID                   string      `json:"id"`
	LogIndex             int         `json:"log_index"`
	TransactionIndex     int         `json:"transaction_index"`
	PoolID               string      `json:"pool_id"`
	Sender               string      `json:"sender"`
	Recipient            string      `json:"recipient"`
	Token0               string      `json:"token_0"`
	Token1               string      `json:"token_1"`
	Amount0              interface{} `json:"amount_0"`
	Amount1              interface{} `json:"amount_1"`
	CreatedAtBlockNumber int64       `json:"created_at_block_number"`
}

// TransactionsResponse represents the response for the transactions endpoint.
type TransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
	PageInfo     PageInfo      `json:"page_info"`
}

// GetTransactions returns transactions of a pool on a network.
// Implements the getPoolTransactions operation from the OpenAPI spec.
func (s *PoolsService) GetTransactions(ctx context.Context, networkID, poolAddress string, page, limit int, cursor string) (*TransactionsResponse, error) {
	path := fmt.Sprintf("/networks/%s/pools/%s/transactions", networkID, poolAddress)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if page > 0 {
		q.Add("page", fmt.Sprintf("%d", page))
	}
	if limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		q.Add("cursor", cursor)
	}
	req.URL.RawQuery = q.Encode()

	var response TransactionsResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}
