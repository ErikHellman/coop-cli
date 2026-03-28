package main

import (
	"os"
	"testing"

	"github.com/ErikHellman/coop-cli/internal/api"
	"github.com/ErikHellman/coop-cli/internal/auth"
	"github.com/ErikHellman/coop-cli/internal/models"
)

// testEnv holds shared state across the ordered integration test steps.
type testEnv struct {
	session  *auth.Session
	client   *api.Client
	storeID  string
	products []models.Product // products confirmed addable to cart
}

var env testEnv

func TestIntegration(t *testing.T) {
	email := os.Getenv("COOP_EMAIL")
	password := os.Getenv("COOP_PASSWORD")
	if email == "" || password == "" {
		t.Skip("COOP_EMAIL and COOP_PASSWORD must be set for integration tests")
	}

	t.Run("StoreSearch", testStoreSearch)
	t.Run("Login", func(t *testing.T) { testLogin(t, email, password) })
	t.Run("ClearCartBefore", testClearCartBefore)
	t.Run("SearchProducts", testSearchProducts)
	t.Run("AddProducts", testAddProducts)
	t.Run("VerifyCartContents", testVerifyCartContents)
	t.Run("UpdateQuantity", testUpdateQuantity)
	t.Run("RemoveProduct", testRemoveProduct)
	t.Run("ClearCartAfter", testClearCartAfter)
}

func testStoreSearch(t *testing.T) {
	stores, err := api.SearchStores("Stora Coop")
	if err != nil {
		t.Fatalf("SearchStores failed: %v", err)
	}
	if len(stores) == 0 {
		t.Fatal("Expected at least one Stora Coop store")
	}

	store := stores[0]
	if store.LedgerAccountNumber == "" {
		t.Fatal("Store is missing LedgerAccountNumber")
	}
	if store.Name == "" {
		t.Fatal("Store is missing Name")
	}
	if store.City == "" {
		t.Fatal("Store is missing City")
	}
	if store.Address == "" {
		t.Fatal("Store is missing Address")
	}
	if store.PostalCode == "" {
		t.Fatal("Store is missing PostalCode")
	}
	if store.Phone == "" {
		t.Fatal("Store is missing Phone")
	}

	env.storeID = store.LedgerAccountNumber
	t.Logf("Selected store: %s (%s, %s, %s %s)",
		store.Name, store.LedgerAccountNumber, store.Address, store.PostalCode, store.City)
}

func testLogin(t *testing.T, email, password string) {
	if env.storeID == "" {
		t.Skip("No store selected")
	}

	session, err := auth.Login(email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if session.Token == "" {
		t.Fatal("Session token is empty")
	}
	if session.ShoppingUserID == "" {
		t.Fatal("ShoppingUserID is empty")
	}

	env.session = session
	env.client = api.NewClient(session, env.storeID)
	t.Logf("Logged in as %s", session.ShoppingUserID)
}

func testClearCartBefore(t *testing.T) {
	requireClient(t)

	err := env.client.ClearCart()
	if err != nil {
		t.Fatalf("ClearCart failed: %v", err)
	}

	cart, err := env.client.GetCart()
	if err != nil {
		t.Fatalf("GetCart failed: %v", err)
	}
	if len(cart.Entries) != 0 {
		t.Fatalf("Expected empty cart, got %d entries", len(cart.Entries))
	}
	t.Log("Cart cleared before tests")
}

func testSearchProducts(t *testing.T) {
	requireClient(t)

	result, err := env.client.SearchProducts("mjölk", 10)
	if err != nil {
		t.Fatalf("SearchProducts failed: %v", err)
	}
	if result.Results.Count == 0 {
		t.Fatal("Expected search results for 'mjölk'")
	}
	if len(result.Results.Items) == 0 {
		t.Fatal("Expected at least one product item")
	}

	// Verify product fields on the first result.
	p := result.Results.Items[0]
	if p.ID == "" {
		t.Error("Product ID is empty")
	}
	if p.Name == "" {
		t.Error("Product Name is empty")
	}
	if p.ManufacturerName == "" {
		t.Error("Product ManufacturerName is empty")
	}
	if p.PackageSizeInfo == "" {
		t.Error("Product PackageSizeInfo is empty")
	}
	if p.SalesPriceData.B2CPrice <= 0 {
		t.Error("Product price should be positive")
	}
	if p.ComparativePriceData.B2CPrice <= 0 {
		t.Error("Product comparative price should be positive")
	}
	if p.ComparativePriceUnit.Text == "" {
		t.Error("Product comparative price unit text is empty")
	}
	if len(p.NavCategories) == 0 {
		t.Error("Product should have at least one category")
	}
	if models.CategoryPath(p.NavCategories) == "" {
		t.Error("CategoryPath should not be empty")
	}

	// Try adding each product to the cart to find ones that are actually
	// available at this store. The search API marks products as availableOnline
	// globally, but individual stores may not stock them.
	for _, item := range result.Results.Items {
		if !item.AvailableOnline {
			continue
		}
		_, err := env.client.AddToCart(item.ID, 1)
		if err != nil {
			t.Logf("Product %s (%s) not available at this store, skipping", item.ID, item.Name)
			continue
		}
		env.products = append(env.products, item)
		t.Logf("Product %s (%s) confirmed available", item.ID, item.Name)
		if len(env.products) >= 2 {
			break
		}
	}

	// Clear the cart after probing so later tests start clean.
	if err := env.client.ClearCart(); err != nil {
		t.Fatalf("ClearCart after probing failed: %v", err)
	}

	if len(env.products) < 2 {
		t.Fatalf("Need at least 2 available products for cart tests, found %d", len(env.products))
	}

	t.Logf("Search returned %d results, %d confirmed available at store", result.Results.Count, len(env.products))
}

func testAddProducts(t *testing.T) {
	requireClient(t)
	requireProducts(t, 2)

	// Add first product with quantity 2.
	first := env.products[0]
	cart, err := env.client.AddToCart(first.ID, 2)
	if err != nil {
		t.Fatalf("AddToCart(%s, 2) failed: %v", first.ID, err)
	}

	entry := findEntry(cart, first.ID)
	if entry == nil {
		t.Fatalf("Product %s not found in cart after adding", first.ID)
	}
	if entry.Quantity != 2 {
		t.Errorf("Expected quantity 2, got %d", entry.Quantity)
	}
	t.Logf("Added %s (qty 2)", first.Name)

	// Add second product with quantity 1.
	second := env.products[1]
	cart, err = env.client.AddToCart(second.ID, 1)
	if err != nil {
		t.Fatalf("AddToCart(%s, 1) failed: %v", second.ID, err)
	}

	entry = findEntry(cart, second.ID)
	if entry == nil {
		t.Fatalf("Product %s not found in cart after adding", second.ID)
	}
	if entry.Quantity != 1 {
		t.Errorf("Expected quantity 1, got %d", entry.Quantity)
	}
	if len(cart.Entries) < 2 {
		t.Errorf("Expected at least 2 cart entries, got %d", len(cart.Entries))
	}
	t.Logf("Added %s (qty 1)", second.Name)
}

func testVerifyCartContents(t *testing.T) {
	requireClient(t)
	requireProducts(t, 2)

	cart, err := env.client.GetCart()
	if err != nil {
		t.Fatalf("GetCart failed: %v", err)
	}

	first := findEntry(cart, env.products[0].ID)
	if first == nil {
		t.Errorf("First product %s not in cart", env.products[0].ID)
	} else {
		if first.Quantity != 2 {
			t.Errorf("First product quantity: expected 2, got %d", first.Quantity)
		}
		if first.BasePrice.Value <= 0 {
			t.Error("First product base price should be positive")
		}
		if first.TotalPrice.Value <= 0 {
			t.Error("First product total price should be positive")
		}
	}

	second := findEntry(cart, env.products[1].ID)
	if second == nil {
		t.Errorf("Second product %s not in cart", env.products[1].ID)
	} else if second.Quantity != 1 {
		t.Errorf("Second product quantity: expected 1, got %d", second.Quantity)
	}

	if cart.TotalPrice.Value <= 0 {
		t.Error("Cart total price should be positive")
	}
	if cart.SubTotal.Value <= 0 {
		t.Error("Cart subtotal should be positive")
	}
	if cart.TotalPrice.FormattedValue == "" {
		t.Error("Cart total formatted value is empty")
	}
	if cart.CoopStore.Name == "" {
		t.Error("Cart store name is empty")
	}

	t.Logf("Cart verified: %d items, subtotal %s, total %s at %s",
		cart.TotalItems, cart.SubTotal.FormattedValue, cart.TotalPrice.FormattedValue, cart.CoopStore.Name)
}

func testUpdateQuantity(t *testing.T) {
	requireClient(t)
	requireProducts(t, 1)

	product := env.products[0]
	cart, err := env.client.AddToCart(product.ID, 5)
	if err != nil {
		t.Fatalf("AddToCart (update qty) failed: %v", err)
	}

	entry := findEntry(cart, product.ID)
	if entry == nil {
		t.Fatalf("Product %s not in cart after quantity update", product.ID)
	}
	if entry.Quantity != 5 {
		t.Errorf("Expected quantity 5, got %d", entry.Quantity)
	}

	t.Logf("Updated %s quantity to 5", product.Name)
}

func testRemoveProduct(t *testing.T) {
	requireClient(t)
	requireProducts(t, 2)

	// Remove first product.
	product := env.products[0]
	cart, err := env.client.RemoveFromCart(product.ID)
	if err != nil {
		t.Fatalf("RemoveFromCart failed: %v", err)
	}

	if findEntry(cart, product.ID) != nil {
		t.Errorf("Product %s should have been removed", product.ID)
	}

	// Second product should still be there.
	if findEntry(cart, env.products[1].ID) == nil {
		t.Errorf("Second product %s should still be in cart", env.products[1].ID)
	}

	t.Logf("Removed %s, %d entries remaining", product.Name, len(cart.Entries))
}

func testClearCartAfter(t *testing.T) {
	requireClient(t)

	err := env.client.ClearCart()
	if err != nil {
		t.Fatalf("ClearCart failed: %v", err)
	}

	cart, err := env.client.GetCart()
	if err != nil {
		t.Fatalf("GetCart after clear failed: %v", err)
	}
	if len(cart.Entries) != 0 {
		t.Errorf("Expected empty cart after clear, got %d entries", len(cart.Entries))
	}
	if cart.TotalPrice.Value != 0 {
		t.Errorf("Expected zero total after clear, got %f", cart.TotalPrice.Value)
	}

	t.Log("Cart cleared after tests")
}

func requireClient(t *testing.T) {
	t.Helper()
	if env.client == nil {
		t.Skip("No authenticated client")
	}
}

func requireProducts(t *testing.T, n int) {
	t.Helper()
	if len(env.products) < n {
		t.Skipf("Need %d products, have %d", n, len(env.products))
	}
}

func findEntry(cart *models.CartResponse, productID string) *models.CartEntry {
	for i := range cart.Entries {
		if cart.Entries[i].Product.Code == productID {
			return &cart.Entries[i]
		}
	}
	return nil
}
