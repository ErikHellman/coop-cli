package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ErikHellman/coop-cli/internal/models"
)

// GetCart returns the current cart contents.
func (c *Client) GetCart() (*models.CartResponse, error) {
	path := fmt.Sprintf("users/%s/carts/current?fields=FULL",
		url.PathEscape(c.session.ShoppingUserID))

	respBody, err := c.doHybris("GET", path, "")
	if err != nil {
		return nil, err
	}

	var cart models.CartResponse
	if err := json.Unmarshal(respBody, &cart); err != nil {
		return nil, fmt.Errorf("decoding cart response: %w", err)
	}

	return &cart, nil
}

// AddToCart adds a product to the cart or updates its quantity.
func (c *Client) AddToCart(productID string, quantity int) (*models.CartModification, error) {
	params := url.Values{}
	params.Set("qty", fmt.Sprintf("%d", quantity))
	params.Set("code", productID)
	params.Set("fields", "FULL")

	path := fmt.Sprintf("users/%s/cartdata/current/products?%s",
		url.PathEscape(c.session.ShoppingUserID),
		params.Encode())

	respBody, err := c.doHybris("POST", path, "")
	if err != nil {
		return nil, err
	}

	var result models.CartModification
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decoding cart modification: %w", err)
	}

	return &result, nil
}

// RemoveFromCart removes a product by setting its quantity to 0.
func (c *Client) RemoveFromCart(productID string) error {
	_, err := c.AddToCart(productID, 0)
	return err
}

// ClearCart empties the entire cart.
func (c *Client) ClearCart() error {
	path := fmt.Sprintf("users/%s/carts/current?emptyCart=true",
		url.PathEscape(c.session.ShoppingUserID))

	_, err := c.doHybris("PUT", path, "")
	return err
}
