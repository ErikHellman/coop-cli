package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ErikHellman/coop-cli/internal/models"
)

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

func (c *Client) AddToCart(productID string, quantity int) (*models.CartResponse, error) {
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

	var cart models.CartResponse
	if err := json.Unmarshal(respBody, &cart); err != nil {
		return nil, fmt.Errorf("decoding cart response: %w", err)
	}

	return &cart, nil
}

func (c *Client) RemoveFromCart(productID string) (*models.CartResponse, error) {
	return c.AddToCart(productID, 0)
}

func (c *Client) ClearCart() error {
	path := fmt.Sprintf("users/%s/carts/current?emptyCart=true",
		url.PathEscape(c.session.ShoppingUserID))

	_, err := c.doHybris("PUT", path, "")
	return err
}
