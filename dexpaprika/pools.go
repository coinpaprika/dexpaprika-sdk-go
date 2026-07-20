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
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Symbol      string   `json:"symbol"`
	Chain       string   `json:"chain"`
	Decimals    int      `json:"decimals"`
	AddedAt     string   `json:"added_at"`
	FDV         *float64 `json:"fdv,omitempty"`
	TotalSupply *float64 `json:"total_supply,omitempty"`
	Description string   `json:"description,omitempty"`
	Website     string   `json:"website,omitempty"`
	Type        string   `json:"type,omitempty"`
	Status      string   `json:"status,omitempty"`
	HasImage    *bool    `json:"has_image,omitempty"`
}

// Pool represents a liquidity pool.
type Pool struct {
	ID                    string   `json:"id"`
	DexID                 string   `json:"dex_id"`
	DexName               string   `json:"dex_name"`
	Chain                 string   `json:"chain"`
	VolumeUSD             float64  `json:"volume_usd"`
	CreatedAt             string   `json:"created_at"`
	CreatedAtBlockNumber  int64    `json:"created_at_block_number"`
	Transactions          int      `json:"transactions"`
	PriceUSD              float64  `json:"price_usd"`
	LastPriceChangeUSD5m  *float64 `json:"last_price_change_usd_5m"`
	LastPriceChangeUSD1h  *float64 `json:"last_price_change_usd_1h"`
	LastPriceChangeUSD24h *float64 `json:"last_price_change_usd_24h"`
	Fee                   *float64 `json:"fee"`
	Tokens                []Token  `json:"tokens"`
	// VolumeUSD/Transactions/LastPriceChangeUSD* are returned by the DEX-pools and
	// token-pools endpoints. The /pools/search endpoint instead returns the
	// timeframe-split and percentage-named fields below; whichever endpoint
	// produced this value leaves the other set nil/zero.
	VolumeUSD24h             *float64 `json:"volume_usd_24h,omitempty"`
	VolumeUSD7d              *float64 `json:"volume_usd_7d,omitempty"`
	VolumeUSD30d             *float64 `json:"volume_usd_30d,omitempty"`
	LiquidityUSD             *float64 `json:"liquidity_usd,omitempty"`
	Transactions24h          int      `json:"transactions_24h,omitempty"`
	PriceChangePercentage5m  *float64 `json:"price_change_percentage_5m,omitempty"`
	PriceChangePercentage1h  *float64 `json:"price_change_percentage_1h,omitempty"`
	PriceChangePercentage24h *float64 `json:"price_change_percentage_24h,omitempty"`
}

// PoolsResponse represents the response for the pools endpoints.
//
// The DEX-pools and token-pools endpoints return rows under "pools" with
// page-based PageInfo. The /pools/search endpoint (used by ListByNetwork)
// returns rows under "results" with cursor-based HasNextPage/NextCursor;
// ListByNetwork copies those rows into Pools so callers keep reading .Pools.
type PoolsResponse struct {
	Pools       []Pool   `json:"pools"`
	Results     []Pool   `json:"results,omitempty"`
	PageInfo    PageInfo `json:"page_info"`
	HasNextPage bool     `json:"has_next_page,omitempty"`
	NextCursor  *string  `json:"next_cursor,omitempty"`
}

// ListOptions contains common options for listing pools.
//
// Page is honored only by the legacy/DEX-pool endpoints. The cursor-paginated
// /pools/search endpoint ignores Page; use Cursor to fetch the next page.
type ListOptions struct {
	Page    int
	Limit   int
	Sort    string
	OrderBy string
	Cursor  string
}

// validateNetworkID validates that a network ID is not empty
func validateNetworkID(networkID string) error {
	if networkID == "" {
		return fmt.Errorf("network ID is required")
	}
	return nil
}

// validatePoolAddress validates that a pool address is not empty
func validatePoolAddress(poolAddress string) error {
	if poolAddress == "" {
		return fmt.Errorf("pool address is required")
	}
	return nil
}

// addOptions adds the parameters in opts as URL query parameters to s.
func addOptions(s string, opts interface{}) (string, error) {
	v := url.Values{}

	if o, ok := opts.(*ListOptions); ok {
		if o.Page > 0 {
			v.Add("page", fmt.Sprintf("%d", o.Page))
		}
		if o.Limit > 0 {
			// Validate limit constraints (max 100 as per API v1.3.0)
			limit := o.Limit
			if limit > 100 {
				limit = 100
			}
			v.Add("limit", fmt.Sprintf("%d", limit))
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
//
// Deprecated: This method calls the deprecated /pools endpoint which has been removed in API v1.3.0.
// Use ListByNetwork(networkID, opts) instead to get pools from a specific network.
//
// Example migration:
//
//	// Old (deprecated):
//	pools, err := client.Pools.List(ctx, opts)
//
//	// New (required):
//	pools, err := client.Pools.ListByNetwork(ctx, "ethereum", opts)
//	pools, err := client.Pools.ListByNetwork(ctx, "solana", opts)
func (s *PoolsService) List(ctx context.Context, opts *ListOptions) (*PoolsResponse, error) {
	// The global /pools endpoint was removed in API v1.3.0 (HTTP 410). Rather
	// than issue a request that always fails, fail client-side with the same
	// structured error the server would have returned, so callers get an
	// immediate, actionable message: errors.Is(err, ErrGone) reports true and
	// errors.As(err, &APIError{}) exposes StatusCode 410 and the replacement.
	return nil, &APIError{
		StatusCode:  http.StatusGone,
		Message:     "the /pools endpoint was removed in API v1.3.0; use ListByNetwork(ctx, networkID, opts) for a specific network",
		Replacement: "/networks/{network}/pools/search",
		Err:         ErrGone,
	}
}

// ListByNetwork returns a list of top pools on a specific network.
//
// Since the DexPaprika API removed /networks/{network}/pools (HTTP 410), this
// method targets the unified /networks/{network}/pools/search endpoint. That
// endpoint is cursor-paginated and rejects the legacy sort values, so the sort
// field is normalized and the Page option is not sent upstream. Pass
// ListOptions.Cursor (taken from a previous response's NextCursor) to page.
func (s *PoolsService) ListByNetwork(ctx context.Context, networkID string, opts *ListOptions) (*PoolsResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("/networks/%s/pools/search", networkID), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	orderBy := ""
	if opts != nil {
		orderBy = opts.OrderBy
	}
	q.Add("order_by", mapPoolSortField(orderBy))
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

// ListByDex returns a list of top pools on a specific network's DEX.
// Implements the getDexPools operation from the OpenAPI spec.
func (s *PoolsService) ListByDex(ctx context.Context, networkID, dexID string, opts *ListOptions) (*PoolsResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}
	if dexID == "" {
		return nil, fmt.Errorf("dex ID is required")
	}

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
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}
	if err := validatePoolAddress(poolAddress); err != nil {
		return nil, err
	}

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
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}
	if err := validatePoolAddress(poolAddress); err != nil {
		return nil, err
	}

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
			// Validate OHLCV limit constraints (max 366)
			limit := opts.Limit
			if limit > 366 {
				limit = 366
			}
			q.Add("limit", fmt.Sprintf("%d", limit))
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

// TransactionOption configures optional parameters for GetTransactions.
type TransactionOption func(q *url.Values)

// WithFromTimestamp filters transactions starting from this UNIX timestamp (inclusive).
// Results are always capped to the last 7 days.
func WithFromTimestamp(from int64) TransactionOption {
	return func(q *url.Values) {
		q.Add("from", fmt.Sprintf("%d", from))
	}
}

// WithToTimestamp filters transactions up to this UNIX timestamp (exclusive).
func WithToTimestamp(to int64) TransactionOption {
	return func(q *url.Values) {
		q.Add("to", fmt.Sprintf("%d", to))
	}
}

// GetTransactions returns transactions of a pool on a network.
// Implements the getPoolTransactions operation from the OpenAPI spec.
// Use WithFromTimestamp and WithToTimestamp to filter by time range.
func (s *PoolsService) GetTransactions(ctx context.Context, networkID, poolAddress string, page, limit int, cursor string, opts ...TransactionOption) (*TransactionsResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}
	if err := validatePoolAddress(poolAddress); err != nil {
		return nil, err
	}

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
		// Validate limit constraints (max 100 as per API v1.3.0)
		if limit > 100 {
			limit = 100
		}
		q.Add("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		q.Add("cursor", cursor)
	}
	for _, opt := range opts {
		opt(&q)
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

// PoolFilterOptions contains options for filtering pools on a network.
//
// Page is accepted for backward compatibility but is not sent to the
// cursor-paginated /pools/search endpoint; use Cursor to page instead.
type PoolFilterOptions struct {
	Page          int
	Limit         int
	SortBy        string
	SortDir       string
	Cursor        string
	Volume24hMin  *float64
	Volume24hMax  *float64
	Volume7dMin   *float64
	Volume7dMax   *float64
	LiquidityMin  *float64
	LiquidityMax  *float64
	Txns24hMin    *int
	CreatedAfter  string
	CreatedBefore string
}

// PoolFilterResponse represents the response from the /pools/search endpoint.
type PoolFilterResponse struct {
	Results     []Pool  `json:"results"`
	HasNextPage bool    `json:"has_next_page"`
	NextCursor  *string `json:"next_cursor,omitempty"`
}

// Filter returns pools on a network filtered by volume, liquidity, transactions, and creation date.
//
// Since /networks/{network}/pools/filter was removed (HTTP 410), this targets
// the unified /networks/{network}/pools/search endpoint: filters use canonical
// query parameter names, sorting uses order_by (normalized) plus sort
// (direction), and pagination is cursor-based (Page is ignored, use Cursor).
func (s *PoolsService) Filter(ctx context.Context, networkID string, opts *PoolFilterOptions) (*PoolFilterResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("/networks/%s/pools/search", networkID), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	sortBy := ""
	if opts != nil {
		sortBy = opts.SortBy
	}
	q.Add("order_by", mapPoolSortField(sortBy))
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
		if opts.Volume7dMin != nil {
			q.Add("volume_usd_7d_min", fmt.Sprintf("%f", *opts.Volume7dMin))
		}
		if opts.Volume7dMax != nil {
			q.Add("volume_usd_7d_max", fmt.Sprintf("%f", *opts.Volume7dMax))
		}
		if opts.LiquidityMin != nil {
			q.Add("liquidity_usd_min", fmt.Sprintf("%f", *opts.LiquidityMin))
		}
		if opts.LiquidityMax != nil {
			q.Add("liquidity_usd_max", fmt.Sprintf("%f", *opts.LiquidityMax))
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

	var response PoolFilterResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}
