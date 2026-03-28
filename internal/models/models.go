package models

import "strings"

type Product struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	ManufacturerName     string          `json:"manufacturerName"`
	PackageSizeInfo      string          `json:"packageSizeInformation"`
	PackageSize          float64         `json:"packageSize"`
	PackageSizeUnit      string          `json:"packageSizeUnit"`
	SalesPriceData       PriceData       `json:"salesPriceData"`
	ComparativePriceData PriceData       `json:"comparativePriceData"`
	ComparativePriceUnit ComparativeUnit `json:"comparativePriceUnit"`
	NavCategories        []NavCategory   `json:"navCategories"`
	AvailableOnline      bool            `json:"availableOnline"`
}

type PriceData struct {
	B2CPrice float64 `json:"b2cPrice"`
	B2BPrice float64 `json:"b2bPrice"`
}

type ComparativeUnit struct {
	Unit string `json:"unit"`
	Text string `json:"text"`
}

type NavCategory struct {
	Code            string        `json:"code"`
	Name            string        `json:"name"`
	SuperCategories []NavCategory `json:"superCategories"`
}

// CategoryPath returns the full category hierarchy (e.g. "Mejeri & Ägg > Mjölk > Mellanmjölk").
func CategoryPath(cats []NavCategory) string {
	if len(cats) == 0 {
		return ""
	}
	parts := collectCategoryNames(cats[0])
	// Reverse to get top-down order.
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, " > ")
}

func collectCategoryNames(cat NavCategory) []string {
	names := []string{cat.Name}
	if len(cat.SuperCategories) > 0 {
		names = append(names, collectCategoryNames(cat.SuperCategories[0])...)
	}
	return names
}

type SearchResponse struct {
	QueryUsed string        `json:"queryUsed"`
	Results   SearchResults `json:"results"`
}

type SearchResults struct {
	Count int       `json:"count"`
	Items []Product `json:"items"`
}

type CartResponse struct {
	Code       string      `json:"code"`
	GUID       string      `json:"guid"`
	Entries    []CartEntry `json:"entries"`
	TotalItems int         `json:"totalItems"`
	TotalPrice PriceValue  `json:"totalPrice"`
	SubTotal   PriceValue  `json:"subTotal"`
	CoopStore  CoopStore   `json:"coopStore"`
}

type CartEntry struct {
	EntryNumber int         `json:"entryNumber"`
	Quantity    int         `json:"quantity"`
	TotalPrice  PriceValue  `json:"totalPrice"`
	BasePrice   PriceValue  `json:"basePrice"`
	Product     CartProduct `json:"product"`
}

type CartProduct struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type PriceValue struct {
	Value          float64 `json:"value"`
	FormattedValue string  `json:"formattedValue"`
	CurrencyISO    string  `json:"currencyIso"`
}

type CoopStore struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Store struct {
	StoreID             int     `json:"storeId"`
	LedgerAccountNumber string  `json:"ledgerAccountNumber"`
	Name                string  `json:"name"`
	Address             string  `json:"address"`
	City                string  `json:"city"`
	PostalCode          string  `json:"postalCode"`
	Phone               string  `json:"phone"`
	OpeningHoursToday   string  `json:"openingHoursToday"`
	Latitude            float64 `json:"latitude"`
	Longitude           float64 `json:"longitude"`
	URL                 string  `json:"url"`
}

