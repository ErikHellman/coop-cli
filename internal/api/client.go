package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hellman/coop-cli/internal/auth"
)

const (
	hybrisBaseURL        = "https://external.api.coop.se/ecommerce/coop"
	personalizationURL   = "https://external.api.coop.se/personalization"
	subscriptionKey      = "3becf0ce306f41a1ae94077c16798187"
	hybrisAPIVersion     = "v1"
)

// Client is the Coop API client.
type Client struct {
	session *auth.Session
	storeID string
}

// NewClient creates a new API client from an authenticated session.
func NewClient(session *auth.Session, storeID string) *Client {
	return &Client{
		session: session,
		storeID: storeID,
	}
}

func (c *Client) doHybris(method, path string, body string) ([]byte, error) {
	fullURL := fmt.Sprintf("%s/%s", hybrisBaseURL, path)
	if strings.Contains(path, "?") {
		fullURL += "&api-version=" + hybrisAPIVersion
	} else {
		fullURL += "?api-version=" + hybrisAPIVersion
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", subscriptionKey)
	req.Header.Set("Accept", "application/json")
	if c.session.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.session.Token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.session.Client.Do(req)
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

	return respBody, nil
}

func (c *Client) doPersonalization(method, path string, body string) ([]byte, error) {
	fullURL := fmt.Sprintf("%s/%s", personalizationURL, path)

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", subscriptionKey)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.session.Client.Do(req)
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

	return respBody, nil
}
