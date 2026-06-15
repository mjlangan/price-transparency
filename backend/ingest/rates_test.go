package ingest

import (
	"testing"
)

func TestFetchRates_ProviderReferences(t *testing.T) {
	fixture := RateFile{
		ProviderReferences: []ProviderReference{
			{
				ProviderGroupID: 42,
				ProviderGroups: []ProviderGroup{
					{
						NPI: []int64{1902960099},
						TIN: TIN{Type: "ein", Value: "11-2700051", BusinessName: "Morton Zinberg MD"},
					},
				},
			},
		},
	}
	srv := serveJSON(t, fixture)
	defer srv.Close()

	rf, err := FetchRates(srv.URL)
	if err != nil {
		t.Fatalf("FetchRates: %v", err)
	}
	if len(rf.ProviderReferences) != 1 {
		t.Fatalf("len(ProviderReferences) = %d, want 1", len(rf.ProviderReferences))
	}
	pr := rf.ProviderReferences[0]
	if pr.ProviderGroupID != 42 {
		t.Errorf("ProviderGroupID = %d, want 42", pr.ProviderGroupID)
	}
	if len(pr.ProviderGroups) == 0 {
		t.Fatal("expected at least one ProviderGroup")
	}
	pg := pr.ProviderGroups[0]
	if len(pg.NPI) == 0 || pg.NPI[0] != 1902960099 {
		t.Errorf("NPI = %v, want [1902960099]", pg.NPI)
	}
	if pg.TIN.Value != "11-2700051" {
		t.Errorf("TIN.Value = %q, want %q", pg.TIN.Value, "11-2700051")
	}
	if pg.TIN.BusinessName != "Morton Zinberg MD" {
		t.Errorf("TIN.BusinessName = %q, want %q", pg.TIN.BusinessName, "Morton Zinberg MD")
	}
}

func TestFetchRates_InNetwork(t *testing.T) {
	fixture := RateFile{
		InNetwork: []InNetworkItem{
			{
				BillingCode:     "99213",
				BillingCodeType: "CPT",
				Description:     "Office visit",
				NegotiatedRates: []NegotiatedRate{
					{
						ProviderReferences: []int{1},
						NegotiatedPrices: []NegotiatedPrice{
							{NegotiatedType: "negotiated", NegotiatedRate: 139.39},
						},
					},
				},
			},
		},
	}
	srv := serveJSON(t, fixture)
	defer srv.Close()

	rf, err := FetchRates(srv.URL)
	if err != nil {
		t.Fatalf("FetchRates: %v", err)
	}
	if len(rf.InNetwork) != 1 {
		t.Fatalf("len(InNetwork) = %d, want 1", len(rf.InNetwork))
	}
	item := rf.InNetwork[0]
	if item.BillingCode != "99213" {
		t.Errorf("BillingCode = %q, want %q", item.BillingCode, "99213")
	}
	if len(item.NegotiatedRates) != 1 {
		t.Fatalf("len(NegotiatedRates) = %d, want 1", len(item.NegotiatedRates))
	}
	if item.NegotiatedRates[0].NegotiatedPrices[0].NegotiatedRate != 139.39 {
		t.Errorf("NegotiatedRate = %f, want 139.39", item.NegotiatedRates[0].NegotiatedPrices[0].NegotiatedRate)
	}
}

func TestFetchRates_NPIInt64(t *testing.T) {
	// NPI values like 1902960099 exceed int32 max (2147483647) — must parse as int64
	fixture := RateFile{
		ProviderReferences: []ProviderReference{
			{ProviderGroupID: 1, ProviderGroups: []ProviderGroup{
				{NPI: []int64{1902960099}},
			}},
		},
	}
	srv := serveJSON(t, fixture)
	defer srv.Close()

	rf, err := FetchRates(srv.URL)
	if err != nil {
		t.Fatalf("FetchRates: %v", err)
	}
	npi := rf.ProviderReferences[0].ProviderGroups[0].NPI[0]
	if npi != 1902960099 {
		t.Errorf("NPI = %d, want 1902960099 (int64 overflow check)", npi)
	}
}
