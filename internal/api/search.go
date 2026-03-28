package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ErikHellman/coop-cli/internal/models"
)

// SearchProducts searches for products by query.
func (c *Client) SearchProducts(query string, take int) (*models.SearchResponse, error) {
	params := url.Values{}
	params.Set("api-version", APIVersion)
	params.Set("store", c.storeID)
	params.Set("groups", "CUSTOMER_PRIVATE")
	params.Set("device", "desktop")
	params.Set("direct", "false")

	path := fmt.Sprintf("search/products?%s", params.Encode())

	requestBody := map[string]interface{}{
		"query": query,
		"resultsOptions": map[string]interface{}{
			"skip":   0,
			"take":   take,
			"sortBy": []interface{}{},
			"facets": []interface{}{},
		},
		"relatedResultsOptions": map[string]interface{}{
			"skip": 0,
			"take": 0,
		},
		"customData": map[string]interface{}{
			"consent": false,
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doPersonalization("POST", path, string(bodyBytes))
	if err != nil {
		return nil, err
	}

	var result models.SearchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decoding search response: %w", err)
	}

	return &result, nil
}
