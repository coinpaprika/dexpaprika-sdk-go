package dexpaprika

import (
	"context"
	"fmt"
	"net/http"
)

// NetworksService handles communication with the networks related
// methods of the DexPaprika API.
type NetworksService struct {
	client *Client
}

// Network represents a blockchain network.
type Network struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// List returns a list of all supported blockchain networks.
// Implements the getNetworks operation from the OpenAPI spec.
func (s *NetworksService) List(ctx context.Context) ([]Network, error) {
	req, err := s.client.NewRequest(http.MethodGet, "/networks", nil)
	if err != nil {
		return nil, err
	}

	var networks []Network
	_, err = s.client.Do(ctx, req, &networks)
	if err != nil {
		return nil, err
	}

	return networks, nil
}

// Dex represents a decentralized exchange.
type Dex struct {
	ID       string `json:"dex_id"`
	Name     string `json:"dex_name"`
	Chain    string `json:"chain"`
	Protocol string `json:"protocol"`
}

// DexesResponse represents the response for the dexes endpoint.
type DexesResponse struct {
	Dexes    []Dex    `json:"dexes"`
	PageInfo PageInfo `json:"page_info"`
}

// PageInfo contains pagination information.
type PageInfo struct {
	Limit      int `json:"limit"`
	Page       int `json:"page"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// ListDexes returns a list of all available dexes on a specific network.
// Implements the getNetworkDexes operation from the OpenAPI spec.
func (s *NetworksService) ListDexes(ctx context.Context, networkID string, page, limit int) (*DexesResponse, error) {
	path := "/networks/" + networkID + "/dexes"

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// Add query parameters
	q := req.URL.Query()
	if page > 0 {
		q.Add("page", fmt.Sprintf("%d", page))
	}
	if limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", limit))
	}
	req.URL.RawQuery = q.Encode()

	var response DexesResponse
	_, err = s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
