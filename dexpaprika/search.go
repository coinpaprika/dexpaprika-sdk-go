package dexpaprika

import (
	"context"
	"net/http"
	"net/url"
)

// SearchService handles communication with the search related
// methods of the DexPaprika API.
type SearchService struct {
	client *Client
}

// DexInfo represents basic information about a DEX in search results.
type DexInfo struct {
	ID           string  `json:"id"`
	DexID        string  `json:"dex_id"`
	DexName      string  `json:"dex_name"`
	Chain        string  `json:"chain"`
	VolumeUSD24h float64 `json:"volume_usd_24h"`
	Txns24h      int     `json:"txns_24h"`
	PoolsCount   int     `json:"pools_count"`
	Protocol     string  `json:"protocol"`
	CreatedAt    string  `json:"created_at"`
}

// SearchResult represents the structure of a search response.
type SearchResult struct {
	Tokens []TokenDetails `json:"tokens"`
	Pools  []Pool         `json:"pools"`
	Dexes  []DexInfo      `json:"dexes"`
}

// Search performs a search across tokens, pools, and DEXes.
func (s *SearchService) Search(ctx context.Context, query string) (*SearchResult, error) {
	req, err := s.client.NewRequest(http.MethodGet, "/search", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("query", url.QueryEscape(query))
	req.URL.RawQuery = q.Encode()

	var result SearchResult
	_, err = s.client.Do(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
