package ingest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serveJSON(t *testing.T, v any) *httptest.Server {
	t.Helper()
	body, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
}

func TestFetchIndex_ExtractsFirstInNetworkURL(t *testing.T) {
	fixture := IndexFile{
		ReportingEntityName: "Test Issuer",
		ReportingStructure: []ReportingStructure{
			{
				ReportingPlans: []ReportingPlan{{PlanName: "Plan A"}},
				InNetworkFiles: []FileRef{{Description: "in-network", Location: "https://example.com/rates.json"}},
			},
		},
	}
	srv := serveJSON(t, fixture)
	defer srv.Close()

	_, url, err := FetchIndex(srv.URL)
	if err != nil {
		t.Fatalf("FetchIndex: %v", err)
	}
	if url != "https://example.com/rates.json" {
		t.Errorf("url = %q, want %q", url, "https://example.com/rates.json")
	}
}

func TestFetchIndex_MultipleReportingStructures(t *testing.T) {
	fixture := IndexFile{
		ReportingStructure: []ReportingStructure{
			{InNetworkFiles: []FileRef{}},
			{InNetworkFiles: []FileRef{{Location: "https://example.com/second.json"}}},
		},
	}
	srv := serveJSON(t, fixture)
	defer srv.Close()

	_, url, err := FetchIndex(srv.URL)
	if err != nil {
		t.Fatalf("FetchIndex: %v", err)
	}
	if url != "https://example.com/second.json" {
		t.Errorf("url = %q, want %q", url, "https://example.com/second.json")
	}
}

func TestFetchIndex_Empty(t *testing.T) {
	fixture := IndexFile{ReportingStructure: []ReportingStructure{}}
	srv := serveJSON(t, fixture)
	defer srv.Close()

	_, _, err := FetchIndex(srv.URL)
	if err == nil {
		t.Error("expected error for empty reporting_structure, got nil")
	}
}

func TestPlanNames_Deduplicates(t *testing.T) {
	idx := &IndexFile{
		ReportingStructure: []ReportingStructure{
			{ReportingPlans: []ReportingPlan{{PlanName: "Plan A"}, {PlanName: "Plan B"}}},
			{ReportingPlans: []ReportingPlan{{PlanName: "Plan A"}}},
		},
	}
	names := PlanNames(idx)
	if len(names) != 2 {
		t.Errorf("PlanNames = %v, want 2 unique names", names)
	}
}
