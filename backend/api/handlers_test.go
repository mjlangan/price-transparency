package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"price-transparency/backend/search"
)

const testPlanID = "TEST-PLAN-001"

// mockState implements AppState for testing.
type mockState struct {
	status      string
	message     string
	errMsg      string
	plans       []search.PlanInfo
	planIndexes map[string]*search.SearchIndex
}

func (m *mockState) GetStatus() (string, string, string) {
	return m.status, m.message, m.errMsg
}

func (m *mockState) GetPlans() []search.PlanInfo {
	return m.plans
}

func (m *mockState) GetIndexForPlan(planID string) *search.SearchIndex {
	if m.planIndexes == nil {
		return nil
	}
	return m.planIndexes[planID]
}

func (m *mockState) AggregateStats() (totalCodes, totalRecords int) {
	seen := make(map[*search.SearchIndex]bool)
	for _, idx := range m.planIndexes {
		if !seen[idx] {
			seen[idx] = true
			totalCodes += idx.TotalCodes
			totalRecords += idx.TotalRecords
		}
	}
	return
}

func testIndex() *search.SearchIndex {
	return &search.SearchIndex{
		ByCode: map[string][]search.SearchRecord{
			"99213": {
				{
					BillingCode:     "99213",
					BillingCodeType: "CPT",
					Description:     "Office visit",
					NPIs:            []int64{1902960099},
					EIN:             "11-2700051",
					BusinessName:    "Acme Medical",
					NegotiatedRate:  139.39,
					NegotiatedType:  "negotiated",
					ExpirationDate:  "9999-12-31",
				},
			},
		},
		TotalCodes:   1,
		TotalRecords: 1,
		FileURL:      "https://example.com/rates.json",
		LoadedAt:     time.Now(),
	}
}

func readyState() *mockState {
	idx := testIndex()
	return &mockState{
		status:  "ready",
		message: "Loaded",
		plans:   []search.PlanInfo{{PlanID: testPlanID, PlanName: "Test Plan", PlanIDType: "hios"}},
		planIndexes: map[string]*search.SearchIndex{
			testPlanID: idx,
		},
	}
}

func get(handler http.HandlerFunc, url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestStatusHandler_Loading(t *testing.T) {
	state := &mockState{status: "fetching_rates", message: "Downloading..."}
	rec := get(StatusHandler(state), "/api/status")

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Status != "fetching_rates" {
		t.Errorf("Status = %q, want %q", resp.Status, "fetching_rates")
	}
	if resp.BillingCodesLoaded != 0 {
		t.Errorf("BillingCodesLoaded should be 0 when not ready")
	}
}

func TestStatusHandler_Ready(t *testing.T) {
	state := readyState()
	rec := get(StatusHandler(state), "/api/status")

	var resp StatusResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Status != "ready" {
		t.Errorf("Status = %q, want ready", resp.Status)
	}
	if resp.BillingCodesLoaded != 1 {
		t.Errorf("BillingCodesLoaded = %d, want 1", resp.BillingCodesLoaded)
	}
	if resp.RateRecordsLoaded != 1 {
		t.Errorf("RateRecordsLoaded = %d, want 1", resp.RateRecordsLoaded)
	}
}

func TestStatusHandler_Error(t *testing.T) {
	state := &mockState{status: "error", message: "Ingestion failed", errMsg: "connection refused"}
	rec := get(StatusHandler(state), "/api/status")

	var resp StatusResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Status != "error" {
		t.Errorf("Status = %q, want error", resp.Status)
	}
	if resp.Error != "connection refused" {
		t.Errorf("Error = %q, want %q", resp.Error, "connection refused")
	}
}

func TestPlansHandler_Ready(t *testing.T) {
	state := readyState()
	rec := get(PlansHandler(state), "/api/plans")

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body map[string][]search.PlanInfo
	json.NewDecoder(rec.Body).Decode(&body)
	plans := body["plans"]
	if len(plans) != 1 {
		t.Fatalf("len(plans) = %d, want 1", len(plans))
	}
	if plans[0].PlanID != testPlanID {
		t.Errorf("PlanID = %q, want %q", plans[0].PlanID, testPlanID)
	}
}

func TestPlansHandler_NotReady(t *testing.T) {
	state := &mockState{status: "fetching_rates"}
	rec := get(PlansHandler(state), "/api/plans")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestSearchHandler_MissingCode(t *testing.T) {
	state := readyState()
	rec := get(SearchHandler(state), "/api/search")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestSearchHandler_MissingPlanID(t *testing.T) {
	state := readyState()
	rec := get(SearchHandler(state), "/api/search?code=99213")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	var resp search.SearchResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == "" {
		t.Error("expected error message for missing plan_id")
	}
}

func TestSearchHandler_UnknownPlanID(t *testing.T) {
	state := readyState()
	rec := get(SearchHandler(state), "/api/search?code=99213&plan_id=UNKNOWN")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	var resp search.SearchResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == "" {
		t.Error("expected error message for unknown plan_id")
	}
}

func TestSearchHandler_NotReady(t *testing.T) {
	state := &mockState{status: "fetching_rates"}
	rec := get(SearchHandler(state), "/api/search?code=99213&plan_id="+testPlanID)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestSearchHandler_HappyPath(t *testing.T) {
	state := readyState()
	rec := get(SearchHandler(state), "/api/search?code=99213&plan_id="+testPlanID)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp search.SearchResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.BillingCode != "99213" {
		t.Errorf("BillingCode = %q, want 99213", resp.BillingCode)
	}
	if resp.ResultCount != 1 {
		t.Errorf("ResultCount = %d, want 1", resp.ResultCount)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(resp.Results))
	}
}

func TestSearchHandler_EmptyResults(t *testing.T) {
	state := readyState()
	rec := get(SearchHandler(state), "/api/search?code=XXXXX&plan_id="+testPlanID)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp search.SearchResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ResultCount != 0 {
		t.Errorf("ResultCount = %d, want 0", resp.ResultCount)
	}
	if resp.Results == nil {
		t.Error("Results should be an empty slice, not nil")
	}
}
