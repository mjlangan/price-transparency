package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"price-transparency/backend/search"
)

// mockState implements AppState for testing.
type mockState struct {
	status  string
	message string
	errMsg  string
	index   *search.SearchIndex
}

func (m *mockState) GetStatus() (string, string, string) {
	return m.status, m.message, m.errMsg
}

func (m *mockState) GetIndex() *search.SearchIndex {
	return m.index
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
	state := &mockState{status: "ready", message: "Loaded", index: testIndex()}
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
	if resp.FileURL == "" {
		t.Error("FileURL should be populated when ready")
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

func TestSearchHandler_MissingCode(t *testing.T) {
	state := &mockState{status: "ready", index: testIndex()}
	rec := get(SearchHandler(state), "/api/search")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestSearchHandler_NotReady(t *testing.T) {
	state := &mockState{status: "fetching_rates"}
	rec := get(SearchHandler(state), "/api/search?code=99213")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestSearchHandler_HappyPath(t *testing.T) {
	state := &mockState{status: "ready", index: testIndex()}
	rec := get(SearchHandler(state), "/api/search?code=99213")

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
	state := &mockState{status: "ready", index: testIndex()}
	rec := get(SearchHandler(state), "/api/search?code=XXXXX")

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
