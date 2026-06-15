package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"price-transparency/backend/api"
	"price-transparency/backend/ingest"
	"price-transparency/backend/search"
)

const defaultIndexURL = "https://www.centene.com/content/dam/centene/Centene%20Corporate/json/DOCUMENT/2026-04-28_fidelis_index.json"

type appState struct {
	mu      sync.RWMutex
	status  string
	message string
	errMsg  string
	index   *search.SearchIndex
}

func (s *appState) setStatus(status, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
	s.message = message
	log.Printf("[%s] %s", status, message)
}

func (s *appState) setError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = "error"
	s.message = "Ingestion failed"
	s.errMsg = err.Error()
	log.Printf("[error] %v", err)
}

func (s *appState) setReady(idx *search.SearchIndex) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.index = idx
	s.status = "ready"
	s.message = fmt.Sprintf("Loaded %d billing codes, %d rate records", idx.TotalCodes, idx.TotalRecords)
	log.Printf("[ready] %s", s.message)
}

// GetStatus implements api.AppState.
func (s *appState) GetStatus() (status, message, errMsg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status, s.message, s.errMsg
}

// GetIndex implements api.AppState.
func (s *appState) GetIndex() *search.SearchIndex {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.index
}

func runIngestion(state *appState, indexURL string) {
	state.setStatus("fetching_index", "Fetching index file...")
	idx, rateURL, err := ingest.FetchIndex(indexURL)
	if err != nil {
		state.setError(fmt.Errorf("index: %w", err))
		return
	}

	planNames := ingest.PlanNames(idx)
	state.setStatus("fetching_rates", fmt.Sprintf("Fetching rate file from %s...", rateURL))

	rf, err := ingest.FetchRates(rateURL)
	if err != nil {
		state.setError(fmt.Errorf("rates: %w", err))
		return
	}

	state.setStatus("building_index", "Building search index...")
	searchIdx := search.Build(rf, planNames, rateURL)
	state.setReady(searchIdx)
}

func main() {
	indexURL := os.Getenv("INDEX_URL")
	if indexURL == "" {
		indexURL = defaultIndexURL
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	state := &appState{status: "idle", message: "Starting up"}

	go runIngestion(state, indexURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", api.StatusHandler(state))
	mux.HandleFunc("/api/search", api.SearchHandler(state))

	addr := ":" + port
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, api.CORSMiddleware(mux)); err != nil {
		log.Fatalf("server: %v", err)
	}
}
