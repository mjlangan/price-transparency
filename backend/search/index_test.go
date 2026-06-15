package search

import (
	"testing"

	"price-transparency/backend/ingest"
)

// fixture builds a minimal RateFile with two billing codes and two provider groups.
func fixture() *ingest.RateFile {
	return &ingest.RateFile{
		ProviderReferences: []ingest.ProviderReference{
			{
				ProviderGroupID: 1,
				ProviderGroups: []ingest.ProviderGroup{
					{NPI: []int64{1902960099}, TIN: ingest.TIN{Type: "ein", Value: "11-2700051", BusinessName: "Acme Medical"}},
				},
			},
			{
				ProviderGroupID: 2,
				ProviderGroups: []ingest.ProviderGroup{
					{NPI: []int64{9876543210}, TIN: ingest.TIN{Type: "ein", Value: "22-3456789", BusinessName: "Beta Clinic"}},
				},
			},
		},
		InNetwork: []ingest.InNetworkItem{
			{
				BillingCode:     "99213",
				BillingCodeType: "CPT",
				Description:     "Office visit",
				NegotiatedRates: []ingest.NegotiatedRate{
					{
						ProviderReferences: []int{1},
						NegotiatedPrices: []ingest.NegotiatedPrice{
							{NegotiatedType: "negotiated", NegotiatedRate: 139.39, ExpirationDate: "9999-12-31", BillingClass: "professional", ServiceCode: []string{"11"}},
						},
					},
					{
						ProviderReferences: []int{2},
						NegotiatedPrices: []ingest.NegotiatedPrice{
							{NegotiatedType: "negotiated", NegotiatedRate: 160.00, ExpirationDate: "2027-06-30", BillingClass: "professional"},
						},
					},
				},
			},
			{
				BillingCode:     "99214",
				BillingCodeType: "CPT",
				Description:     "Office visit, moderate",
				NegotiatedRates: []ingest.NegotiatedRate{
					{
						ProviderReferences: []int{1},
						NegotiatedPrices: []ingest.NegotiatedPrice{
							{NegotiatedType: "negotiated", NegotiatedRate: 200.00, ExpirationDate: "9999-12-31"},
						},
					},
				},
			},
		},
	}
}

func buildFixture() *SearchIndex {
	return Build(fixture(), []string{"Test Plan"}, "https://example.com/rates.json")
}

func TestBuild_ProviderResolution(t *testing.T) {
	idx := buildFixture()
	records := idx.ByCode["99213"]
	if len(records) == 0 {
		t.Fatal("expected records for 99213")
	}
	r := records[0]
	if r.EIN != "11-2700051" {
		t.Errorf("EIN = %q, want %q", r.EIN, "11-2700051")
	}
	if r.BusinessName != "Acme Medical" {
		t.Errorf("BusinessName = %q, want %q", r.BusinessName, "Acme Medical")
	}
	if len(r.NPIs) == 0 || r.NPIs[0] != 1902960099 {
		t.Errorf("NPIs = %v, want [1902960099]", r.NPIs)
	}
}

func TestBuild_CrossProduct(t *testing.T) {
	idx := buildFixture()
	// 99213 has 2 provider groups × 1 price each = 2 records
	if got := len(idx.ByCode["99213"]); got != 2 {
		t.Errorf("len(ByCode[99213]) = %d, want 2", got)
	}
}

func TestBuild_TotalCounts(t *testing.T) {
	idx := buildFixture()
	if idx.TotalCodes != 2 {
		t.Errorf("TotalCodes = %d, want 2", idx.TotalCodes)
	}
	// 99213: 2 records, 99214: 1 record
	if idx.TotalRecords != 3 {
		t.Errorf("TotalRecords = %d, want 3", idx.TotalRecords)
	}
}

func TestSearch_ByCodeOnly(t *testing.T) {
	idx := buildFixture()
	results, desc, codeType := idx.Search("99213", "", "", 100)
	if len(results) != 2 {
		t.Errorf("len(results) = %d, want 2", len(results))
	}
	if desc != "Office visit" {
		t.Errorf("description = %q, want %q", desc, "Office visit")
	}
	if codeType != "CPT" {
		t.Errorf("billing_code_type = %q, want %q", codeType, "CPT")
	}
}

func TestSearch_UnknownCode(t *testing.T) {
	idx := buildFixture()
	results, _, _ := idx.Search("XXXXX", "", "", 100)
	if results != nil {
		t.Errorf("expected nil for unknown code, got %v", results)
	}
}

func TestSearch_FilterByNPI(t *testing.T) {
	idx := buildFixture()
	results, _, _ := idx.Search("99213", "1902960099", "", 100)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].NPIs[0] != 1902960099 {
		t.Errorf("unexpected NPI %d", results[0].NPIs[0])
	}
}

func TestSearch_FilterByNPI_NoMatch(t *testing.T) {
	idx := buildFixture()
	results, _, _ := idx.Search("99213", "1111111111", "", 100)
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-matching NPI, got %d", len(results))
	}
}

func TestSearch_FilterByEIN_WithHyphen(t *testing.T) {
	idx := buildFixture()
	results, _, _ := idx.Search("99213", "", "11-2700051", 100)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].EIN != "11-2700051" {
		t.Errorf("EIN = %q", results[0].EIN)
	}
}

func TestSearch_FilterByEIN_WithoutHyphen(t *testing.T) {
	idx := buildFixture()
	results, _, _ := idx.Search("99213", "", "112700051", 100)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1 (hyphen normalization)", len(results))
	}
}

func TestSearch_FilterByEIN_CaseInsensitive(t *testing.T) {
	idx := buildFixture()
	results, _, _ := idx.Search("99213", "", "11-2700051", 100)
	if len(results) != 1 {
		t.Errorf("expected 1 result for case-insensitive EIN match, got %d", len(results))
	}
}

func TestSearch_Limit(t *testing.T) {
	idx := buildFixture()
	results, _, _ := idx.Search("99213", "", "", 1)
	if len(results) != 1 {
		t.Errorf("len(results) = %d, want 1 (limit applied)", len(results))
	}
}
