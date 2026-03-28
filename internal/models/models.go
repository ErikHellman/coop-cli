package models

// Product represents a Coop product from search results.
type Product struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	ManufacturerName     string            `json:"manufacturerName"`
	PackageSizeInfo      string            `json:"packageSizeInformation"`
	PackageSize          float64           `json:"packageSize"`
	PackageSizeUnit      string            `json:"packageSizeUnit"`
	SalesPriceData       PriceData         `json:"salesPriceData"`
	ComparativePriceData PriceData         `json:"comparativePriceData"`
	ComparativePriceUnit ComparativeUnit   `json:"comparativePriceUnit"`
	NavCategories        []NavCategory     `json:"navCategories"`
	AvailableOnline      bool              `json:"availableOnline"`
}

// PriceData holds b2c and b2b prices.
type PriceData struct {
	B2CPrice float64 `json:"b2cPrice"`
	B2BPrice float64 `json:"b2bPrice"`
}

// ComparativeUnit holds the comparison price unit.
type ComparativeUnit struct {
	Unit string `json:"unit"`
	Text string `json:"text"`
}

// NavCategory represents a product's navigation category.
type NavCategory struct {
	Code            string        `json:"code"`
	Name            string        `json:"name"`
	SuperCategories []NavCategory `json:"superCategories"`
}

// CategoryPath returns the full category path as a string.
func CategoryPath(cats []NavCategory) string {
	if len(cats) == 0 {
		return ""
	}
	cat := cats[0]
	parts := collectCategoryNames(cat)
	// Reverse to get top-down order.
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " > "
		}
		result += p
	}
	return result
}

func collectCategoryNames(cat NavCategory) []string {
	names := []string{cat.Name}
	if len(cat.SuperCategories) > 0 {
		names = append(names, collectCategoryNames(cat.SuperCategories[0])...)
	}
	return names
}

// SearchResponse is the top-level response from the product search API.
type SearchResponse struct {
	QueryUsed string        `json:"queryUsed"`
	Results   SearchResults `json:"results"`
}

// SearchResults contains the search result items and count.
type SearchResults struct {
	Count int       `json:"count"`
	Items []Product `json:"items"`
}

// CartResponse is the response from the cart API.
type CartResponse struct {
	Code       string      `json:"code"`
	GUID       string      `json:"guid"`
	Entries    []CartEntry `json:"entries"`
	TotalItems int         `json:"totalItems"`
	TotalPrice PriceValue  `json:"totalPrice"`
	SubTotal   PriceValue  `json:"subTotal"`
	CoopStore  CoopStore   `json:"coopStore"`
}

// CartEntry is a single item in the cart.
type CartEntry struct {
	EntryNumber int          `json:"entryNumber"`
	Quantity    int          `json:"quantity"`
	TotalPrice  PriceValue   `json:"totalPrice"`
	BasePrice   PriceValue   `json:"basePrice"`
	Product     CartProduct  `json:"product"`
}

// CartProduct is the product info embedded in a cart entry.
type CartProduct struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// PriceValue holds a formatted price.
type PriceValue struct {
	Value          float64 `json:"value"`
	FormattedValue string  `json:"formattedValue"`
	CurrencyISO    string  `json:"currencyIso"`
}

// CoopStore identifies a Coop store.
type CoopStore struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CartModification is the response when adding/updating cart items.
type CartModification struct {
	StatusCode string    `json:"statusCode"`
	Quantity   int       `json:"quantity"`
	Entry      CartEntry `json:"entry"`
}
