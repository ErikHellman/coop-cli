package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ErikHellman/coop-cli/internal/auth"
)

const (
	hybrisBaseURL      = "https://external.api.coop.se/ecommerce/coop"
	personalizationURL = "https://external.api.coop.se/personalization"
	subscriptionKey    = "3becf0ce306f41a1ae94077c16798187"
	APIVersion         = "v1"
)

type Client struct {
	session *auth.Session
	storeID string
}

func NewClient(session *auth.Session, storeID string) *Client {
	return &Client{
		session: session,
		storeID: storeID,
	}
}

type requestOpts struct {
	baseURL    string
	addAuth    bool
	apiVersion string
}

func (c *Client) doRequest(method, path, body string, opts requestOpts) ([]byte, error) {
	fullURL := opts.baseURL + "/" + path
	if opts.apiVersion != "" {
		parsed, err := url.Parse(fullURL)
		if err != nil {
			return nil, err
		}
		q := parsed.Query()
		q.Set("api-version", opts.apiVersion)
		parsed.RawQuery = q.Encode()
		fullURL = parsed.String()
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
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if opts.addAuth && c.session.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.session.Token)
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

func (c *Client) doHybris(method, path, body string) ([]byte, error) {
	return c.doRequest(method, path, body, requestOpts{
		baseURL:    hybrisBaseURL,
		addAuth:    true,
		apiVersion: APIVersion,
	})
}

func (c *Client) doPersonalization(method, path, body string) ([]byte, error) {
	return c.doRequest(method, path, body, requestOpts{
		baseURL: personalizationURL,
	})
}
