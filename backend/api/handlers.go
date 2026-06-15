package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"price-transparency/backend/search"
)

const defaultLimit = 100

type StatusResponse struct {
	Status             string `json:"status"`
	Message            string `json:"message"`
	Error              string `json:"error,omitempty"`
	FileURL            string `json:"file_url,omitempty"`
	LoadedAt           string `json:"loaded_at,omitempty"`
	BillingCodesLoaded int    `json:"billing_codes_loaded,omitempty"`
	RateRecordsLoaded  int    `json:"rate_records_loaded,omitempty"`
}

// AppState is the shared ingestion state. Implemented by main.go to avoid a circular import.
type AppState interface {
	GetStatus() (status, message, errMsg string)
	GetIndex() *search.SearchIndex
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

		if idx := state.GetIndex(); idx != nil {
			resp.FileURL = idx.FileURL
			resp.LoadedAt = idx.LoadedAt.UTC().Format("2006-01-02T15:04:05Z")
			resp.BillingCodesLoaded = idx.TotalCodes
			resp.RateRecordsLoaded = idx.TotalRecords
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func SearchHandler(state AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idx := state.GetIndex()
		if idx == nil {
			status, _, _ := state.GetStatus()
			writeJSON(w, http.StatusServiceUnavailable, search.SearchResponse{
				Error: "data is still loading",
			})
			_ = status
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			writeJSON(w, http.StatusBadRequest, search.SearchResponse{
				Error: "query parameter 'code' is required",
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
