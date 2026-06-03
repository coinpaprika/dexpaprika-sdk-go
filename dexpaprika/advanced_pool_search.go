package dexpaprika

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AdvancedSearchOptions contains options for the advanced pool search endpoints
// (GET /frontend/v1/pools and GET /frontend/v1/networks/{network}/pools).
//
// Sorting uses the canonical SortBy/SortDir field names on this surface. They are
// translated to the wire parameters order_by and sort respectively before the
// request goes out, so callers never deal with the backend's inverted naming.
type AdvancedSearchOptions struct {
	// Limit caps the number of results returned (defaults to the API default when 0).
	Limit int
	// Cursor is the opaque pagination token. Pass the NextCursor from a previous
	// response to fetch the next page. This endpoint is cursor-based, not page-based.
	Cursor string
	// SortBy is the field to order by. One of: volume_usd_24h, volume_usd_7d,
	// volume_usd_30d, liquidity_usd, txns_24h, price_usd, price_change_percentage_24h,
	// created_at. Sent on the wire as order_by. Defaults to volume_usd_24h.
	SortBy string
	// SortDir is the sort direction, "asc" or "desc". Sent on the wire as sort.
	// Defaults to desc.
	SortDir string
	// Detailed, when true, makes each token in a result carry its FDV and the
	// per-timeframe metric blocks (1m, 5m, 15m, 30m, 1h, 6h, 24h).
	Detailed bool

	// Optional filters. Each is sent under the same name on the wire. Pointer fields
	// are only included in the request when non-nil so the zero value stays meaningful.
	Volume24hMin *float64
	Volume24hMax *float64
	Volume7dMin  *float64
	Volume7dMax  *float64
	LiquidityMin *float64
	LiquidityMax *float64
	Txns24hMin   *int

	PriceUSDMin *float64
	PriceUSDMax *float64

	PriceChange24hMin *float64
	PriceChange24hMax *float64

	DexName       string
	CreatedAfter  string
	CreatedBefore string
}

// AdvancedTokenTimeframe holds the per-timeframe metrics attached to a token when
// AdvancedSearchOptions.Detailed is true. Every field is a pointer because the API
// may omit any of them (notably LastPriceUSDChange is frequently null).
type AdvancedTokenTimeframe struct {
	VolumeUSD          *float64 `json:"volume_usd,omitempty"`
	Buys               *int     `json:"buys,omitempty"`
	Sells              *int     `json:"sells,omitempty"`
	Txns               *int     `json:"txns,omitempty"`
	LastPriceUSDChange *float64 `json:"last_price_usd_change,omitempty"`
}

// AdvancedToken is a token as returned by the advanced pool search endpoints.
//
// In the non-detailed responses tokens carry only a subset of these fields
// (id, chain, has_image), so every field beyond ID is optional. When Detailed is
// requested the FDV and timeframe blocks are populated.
type AdvancedToken struct {
	ID          string   `json:"id"`
	Name        string   `json:"name,omitempty"`
	Symbol      string   `json:"symbol,omitempty"`
	Chain       string   `json:"chain,omitempty"`
	Decimals    *int     `json:"decimals,omitempty"`
	AddedAt     string   `json:"added_at,omitempty"`
	Status      string   `json:"status,omitempty"`
	HasImage    *bool    `json:"has_image,omitempty"`
	FDV         *float64 `json:"fdv,omitempty"`
	TotalSupply *float64 `json:"total_supply,omitempty"`
	Website     string   `json:"website,omitempty"`
	Description string   `json:"description,omitempty"`

	// Timeframe metric blocks, populated only when Detailed is true.
	Minute1  *AdvancedTokenTimeframe `json:"1m,omitempty"`
	Minute5  *AdvancedTokenTimeframe `json:"5m,omitempty"`
	Minute15 *AdvancedTokenTimeframe `json:"15m,omitempty"`
	Minute30 *AdvancedTokenTimeframe `json:"30m,omitempty"`
	Hour1    *AdvancedTokenTimeframe `json:"1h,omitempty"`
	Hour6    *AdvancedTokenTimeframe `json:"6h,omitempty"`
	Day      *AdvancedTokenTimeframe `json:"24h,omitempty"`
}

// PoolRow is a single pool returned by the advanced pool search endpoints.
//
// Metric fields are pointers so a value the API drops stays distinguishable from a
// genuine zero. This mirrors the lesson from the pool/token filter fix: never
// hard-require a field the backend is allowed to omit.
type PoolRow struct {
	ID                   string   `json:"id"`
	DexID                string   `json:"dex_id"`
	DexName              string   `json:"dex_name"`
	Chain                string   `json:"chain"`
	Fee                  *float64 `json:"fee"`
	CreatedAt            string   `json:"created_at"`
	CreatedAtBlockNumber int64    `json:"created_at_block_number"`

	PriceUSD        *float64 `json:"price_usd,omitempty"`
	Transactions24h *int     `json:"transactions_24h,omitempty"`
	VolumeUSD24h    *float64 `json:"volume_usd_24h,omitempty"`
	VolumeUSD7d     *float64 `json:"volume_usd_7d,omitempty"`
	VolumeUSD30d    *float64 `json:"volume_usd_30d,omitempty"`
	LiquidityUSD    *float64 `json:"liquidity_usd,omitempty"`

	PriceChangePercentage5m  *float64 `json:"price_change_percentage_5m,omitempty"`
	PriceChangePercentage1h  *float64 `json:"price_change_percentage_1h,omitempty"`
	PriceChangePercentage24h *float64 `json:"price_change_percentage_24h,omitempty"`

	Tokens []AdvancedToken `json:"tokens"`
}

// AdvancedSearchResponse is the cursor-paginated response shape of the advanced
// pool search endpoints.
type AdvancedSearchResponse struct {
	Results     []PoolRow `json:"results"`
	HasNextPage bool      `json:"has_next_page"`
	// NextCursor is nil on the final page; pass it back via AdvancedSearchOptions.Cursor
	// to fetch the next page.
	NextCursor *string `json:"next_cursor"`
	// Query echoes the parameters the backend actually applied (using its wire names,
	// e.g. order_by and sort). Kept as raw JSON so new echo fields never break decoding.
	Query json.RawMessage `json:"query,omitempty"`
}

// buildAdvancedSearchQuery applies opts to the request, translating the canonical
// SortBy/SortDir names to the wire parameters order_by/sort.
func (o *AdvancedSearchOptions) apply(req *http.Request) {
	if o == nil {
		return
	}
	q := req.URL.Query()

	if o.Limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", o.Limit))
	}
	if o.Cursor != "" {
		q.Add("cursor", o.Cursor)
	}
	// Canonical -> wire translation. This is the inverse of intuition: the surface
	// speaks sort_by/sort_dir, the backend speaks order_by/sort.
	if o.SortBy != "" {
		q.Add("order_by", o.SortBy)
	}
	if o.SortDir != "" {
		q.Add("sort", o.SortDir)
	}
	if o.Detailed {
		q.Add("detailed", "true")
	}

	if o.Volume24hMin != nil {
		q.Add("volume_24h_min", fmt.Sprintf("%v", *o.Volume24hMin))
	}
	if o.Volume24hMax != nil {
		q.Add("volume_24h_max", fmt.Sprintf("%v", *o.Volume24hMax))
	}
	if o.Volume7dMin != nil {
		q.Add("volume_7d_min", fmt.Sprintf("%v", *o.Volume7dMin))
	}
	if o.Volume7dMax != nil {
		q.Add("volume_7d_max", fmt.Sprintf("%v", *o.Volume7dMax))
	}
	if o.LiquidityMin != nil {
		q.Add("liquidity_usd_min", fmt.Sprintf("%v", *o.LiquidityMin))
	}
	if o.LiquidityMax != nil {
		q.Add("liquidity_usd_max", fmt.Sprintf("%v", *o.LiquidityMax))
	}
	if o.Txns24hMin != nil {
		q.Add("txns_24h_min", fmt.Sprintf("%d", *o.Txns24hMin))
	}
	if o.PriceUSDMin != nil {
		q.Add("price_usd_min", fmt.Sprintf("%v", *o.PriceUSDMin))
	}
	if o.PriceUSDMax != nil {
		q.Add("price_usd_max", fmt.Sprintf("%v", *o.PriceUSDMax))
	}
	if o.PriceChange24hMin != nil {
		q.Add("price_change_percentage_24h_min", fmt.Sprintf("%v", *o.PriceChange24hMin))
	}
	if o.PriceChange24hMax != nil {
		q.Add("price_change_percentage_24h_max", fmt.Sprintf("%v", *o.PriceChange24hMax))
	}
	if o.DexName != "" {
		q.Add("dex_name", o.DexName)
	}
	if o.CreatedAfter != "" {
		q.Add("created_after", o.CreatedAfter)
	}
	if o.CreatedBefore != "" {
		q.Add("created_before", o.CreatedBefore)
	}

	req.URL.RawQuery = q.Encode()
}

// AdvancedSearch searches pools across all networks via GET /frontend/v1/pools.
//
// It supports rich filtering, canonical SortBy/SortDir sorting, optional detailed
// token metrics, and cursor pagination. Use the returned NextCursor (when
// HasNextPage is true) as opts.Cursor on the next call to page through results.
//
// Example:
//
//	minPrice := 0.5
//	res, err := client.Pools.AdvancedSearch(ctx, &dexpaprika.AdvancedSearchOptions{
//	    Limit:       10,
//	    SortBy:      "volume_usd_24h",
//	    SortDir:     "desc",
//	    PriceUSDMin: &minPrice,
//	    DexName:     "uniswap_v3",
//	    Detailed:    true,
//	})
func (s *PoolsService) AdvancedSearch(ctx context.Context, opts *AdvancedSearchOptions) (*AdvancedSearchResponse, error) {
	req, err := s.client.NewRequest(http.MethodGet, "/frontend/v1/pools", nil)
	if err != nil {
		return nil, err
	}
	opts.apply(req)

	var response AdvancedSearchResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

// AdvancedSearchByNetwork searches pools on a single network via
// GET /frontend/v1/networks/{network}/pools. It accepts the same options as
// AdvancedSearch.
func (s *PoolsService) AdvancedSearchByNetwork(ctx context.Context, networkID string, opts *AdvancedSearchOptions) (*AdvancedSearchResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("/frontend/v1/networks/%s/pools", networkID), nil)
	if err != nil {
		return nil, err
	}
	opts.apply(req)

	var response AdvancedSearchResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}
