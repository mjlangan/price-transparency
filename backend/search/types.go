package search

import "time"

// SearchRecord is a fully denormalized rate record — one per
// (billing_code, provider_group, negotiated_price) combination.
type SearchRecord struct {
	BillingCode     string   `json:"billing_code"`
	BillingCodeType string   `json:"billing_code_type"`
	Description     string   `json:"description"`
	ProviderGroupID int      `json:"provider_group_id"`
	NPIs            []int64  `json:"npis"`
	EIN             string   `json:"ein"`
	EINType         string   `json:"ein_type"`
	BusinessName    string   `json:"business_name"`
	NegotiatedRate  float64  `json:"negotiated_rate"`
	NegotiatedType  string   `json:"negotiated_type"`
	BillingClass    string   `json:"billing_class"`
	Setting         string   `json:"setting"`
	ServiceCodes    []string `json:"service_codes"`
	Modifiers       []string `json:"modifiers"`
	ExpirationDate  string   `json:"expiration_date"`
}

// SearchIndex is built once at startup and is read-only thereafter.
type SearchIndex struct {
	ByCode       map[string][]SearchRecord
	PlanNames    []string
	FileURL      string
	LoadedAt     time.Time
	TotalCodes   int
	TotalRecords int
}

// SearchResponse is the JSON shape returned by /api/search.
type SearchResponse struct {
	BillingCode     string         `json:"billing_code"`
	BillingCodeType string         `json:"billing_code_type"`
	Description     string         `json:"description"`
	ResultCount     int            `json:"result_count"`
	Results         []SearchRecord `json:"results"`
	Error           string         `json:"error,omitempty"`
}
