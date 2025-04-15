package dexpaprika

import (
	"context"
	"net/http"
)

// UtilsService handles communication with the utility methods
// of the DexPaprika API.
type UtilsService struct {
	client *Client
}

// Stats represents high-level statistics about the DexPaprika ecosystem.
type Stats struct {
	Chains    int `json:"chains"`
	Factories int `json:"factories"`
	Pools     int `json:"pools"`
	Tokens    int `json:"tokens"`
}

// GetStats retrieves high-level statistics about the DexPaprika ecosystem.
// Implements the getStats operation from the OpenAPI spec.
func (s *UtilsService) GetStats(ctx context.Context) (*Stats, error) {
	req, err := s.client.NewRequest(http.MethodGet, "/stats", nil)
	if err != nil {
		return nil, err
	}

	var stats Stats
	_, err = s.client.Do(ctx, req, &stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
