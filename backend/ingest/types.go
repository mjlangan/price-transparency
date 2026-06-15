package ingest

// Index file (table of contents)

type IndexFile struct {
	ReportingEntityName string               `json:"reporting_entity_name"`
	ReportingEntityType string               `json:"reporting_entity_type"`
	LastUpdatedOn       string               `json:"last_updated_on"`
	Version             string               `json:"version"`
	ReportingStructure  []ReportingStructure `json:"reporting_structure"`
}

type ReportingStructure struct {
	ReportingPlans    []ReportingPlan `json:"reporting_plans"`
	InNetworkFiles    []FileRef       `json:"in_network_files"`
	AllowedAmountFile *FileRef        `json:"allowed_amount_file"`
}

type ReportingPlan struct {
	PlanName       string `json:"plan_name"`
	PlanIDType     string `json:"plan_id_type"`
	PlanID         string `json:"plan_id"`
	PlanMarketType string `json:"plan_market_type"`
	IssuerName     string `json:"issuer_name"`
}

type FileRef struct {
	Description string `json:"description"`
	Location    string `json:"location"`
}

// In-network rate file

type RateFile struct {
	ReportingEntityName string              `json:"reporting_entity_name"`
	ReportingEntityType string              `json:"reporting_entity_type"`
	LastUpdatedOn       string              `json:"last_updated_on"`
	Version             string              `json:"version"`
	ProviderReferences  []ProviderReference `json:"provider_references"`
	InNetwork           []InNetworkItem     `json:"in_network"`
}

type ProviderReference struct {
	ProviderGroupID int             `json:"provider_group_id"`
	ProviderGroups  []ProviderGroup `json:"provider_groups"`
	// Location is a URL to an external provider list; not supported — logged as warning if present.
	Location string `json:"location"`
}

type ProviderGroup struct {
	NPI []int64 `json:"npi"`
	TIN TIN     `json:"tin"`
}

type TIN struct {
	Type         string `json:"type"`
	Value        string `json:"value"`
	BusinessName string `json:"business_name"`
}

type InNetworkItem struct {
	NegotiationArrangement string           `json:"negotiation_arrangement"`
	Name                   string           `json:"name"`
	BillingCodeType        string           `json:"billing_code_type"`
	BillingCodeTypeVersion string           `json:"billing_code_type_version"`
	BillingCode            string           `json:"billing_code"`
	Description            string           `json:"description"`
	NegotiatedRates        []NegotiatedRate `json:"negotiated_rates"`
}

type NegotiatedRate struct {
	ProviderReferences []int             `json:"provider_references"`
	NegotiatedPrices   []NegotiatedPrice `json:"negotiated_prices"`
}

type NegotiatedPrice struct {
	NegotiatedType      string   `json:"negotiated_type"`
	NegotiatedRate      float64  `json:"negotiated_rate"`
	ExpirationDate      string   `json:"expiration_date"`
	BillingClass        string   `json:"billing_class"`
	ServiceCode         []string `json:"service_code"`
	BillingCodeModifier []string `json:"billing_code_modifier"`
	Setting             string   `json:"setting"`
}
