package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"price-transparency/backend/search"
)

const defaultLimit = 100

type StatusResponse struct {
	Status             string `json:"status"`
	Message            string `json:"message"`
	Error              string `json:"error,omitempty"`
	BillingCodesLoaded int    `json:"billing_codes_loaded,omitempty"`
	RateRecordsLoaded  int    `json:"rate_records_loaded,omitempty"`
}

// AppState is the shared ingestion state. Implemented by main.go to avoid a circular import.
type AppState interface {
	GetStatus() (status, message, errMsg string)
	GetPlans() []search.PlanInfo
	GetIndexForPlan(planID string) *search.SearchIndex
	AggregateStats() (totalCodes, totalRecords int)
}

func StatusHandler(state AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, message, errMsg := state.GetStatus()
		resp := StatusResponse{
			Status:  status,
			Message: message,
		}
		if errMsg != "" {
			resp.Error = errMsg
		}
		if status == "ready" {
			codes, records := state.AggregateStats()
			resp.BillingCodesLoaded = codes
			resp.RateRecordsLoaded = records
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func PlansHandler(state AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, _, _ := state.GetStatus()
		if status != "ready" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "data is still loading"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"plans": state.GetPlans()})
	}
}

func SearchHandler(state AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, _, _ := state.GetStatus()
		if status != "ready" {
			writeJSON(w, http.StatusServiceUnavailable, search.SearchResponse{
				Error: "data is still loading",
			})
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			writeJSON(w, http.StatusBadRequest, search.SearchResponse{
				Error: "query parameter 'code' is required",
			})
			return
		}

		planID := r.URL.Query().Get("plan_id")
		if planID == "" {
			writeJSON(w, http.StatusBadRequest, search.SearchResponse{
				Error: "query parameter 'plan_id' is required",
			})
			return
		}

		idx := state.GetIndexForPlan(planID)
		if idx == nil {
			writeJSON(w, http.StatusBadRequest, search.SearchResponse{
				Error: fmt.Sprintf("unknown plan_id: %q", planID),
			})
			return
		}

		npi := r.URL.Query().Get("npi")
		ein := r.URL.Query().Get("ein")

		limit := defaultLimit
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 {
				limit = n
			}
		}

		results, description, codeType := idx.Search(code, npi, ein, limit)
		if results == nil {
			results = []search.SearchRecord{}
		}

		writeJSON(w, http.StatusOK, search.SearchResponse{
			BillingCode:     code,
			BillingCodeType: codeType,
			Description:     description,
			ResultCount:     len(results),
			Results:         results,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
