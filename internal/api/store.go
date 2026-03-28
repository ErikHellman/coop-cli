package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ErikHellman/coop-cli/internal/models"
)

// SearchStores fetches all stores and filters by query (name, city, address, or store ID).
// Does not require authentication.
func SearchStores(query string) ([]models.Store, error) {
	req, err := http.NewRequest("GET", storeBaseURL+"/stores/map?api-version=v2", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", storeSubKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var stores []models.Store
	if err := json.Unmarshal(respBody, &stores); err != nil {
		return nil, fmt.Errorf("decoding stores: %w", err)
	}

	if query == "" {
		return stores, nil
	}

	q := strings.ToLower(query)
	var filtered []models.Store
	for _, s := range stores {
		if strings.Contains(strings.ToLower(s.Name), q) ||
			strings.Contains(strings.ToLower(s.City), q) ||
			strings.Contains(strings.ToLower(s.Address), q) ||
			s.LedgerAccountNumber == query {
			filtered = append(filtered, s)
		}
	}

	return filtered, nil
}
