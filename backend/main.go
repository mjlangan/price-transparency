package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"

	"price-transparency/backend/api"
	"price-transparency/backend/ingest"
	"price-transparency/backend/search"
)

const defaultIndexURL = "https://www.centene.com/content/dam/centene/Centene%20Corporate/json/DOCUMENT/2026-04-28_fidelis_index.json"

type appState struct {
	mu          sync.RWMutex
	status      string
	message     string
	errMsg      string
	plans       []search.PlanInfo              // sorted alphabetically by plan name
	planIndexes map[string]*search.SearchIndex // plan_id → index (multiple plans share one index)
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

func (s *appState) setReady(plans []search.PlanInfo, planIndexes map[string]*search.SearchIndex) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.plans = plans
	s.planIndexes = planIndexes
	s.status = "ready"
	codes, records := aggregateStats(planIndexes)
	s.message = fmt.Sprintf("Loaded %d plans, %d billing codes, %d rate records", len(plans), codes, records)
	log.Printf("[ready] %s", s.message)
}

// GetStatus implements api.AppState.
func (s *appState) GetStatus() (status, message, errMsg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status, s.message, s.errMsg
}

// GetPlans implements api.AppState.
func (s *appState) GetPlans() []search.PlanInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.plans
}

// GetIndexForPlan implements api.AppState.
func (s *appState) GetIndexForPlan(planID string) *search.SearchIndex {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.planIndexes == nil {
		return nil
	}
	return s.planIndexes[planID]
}

// AggregateStats implements api.AppState.
func (s *appState) AggregateStats() (totalCodes, totalRecords int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return aggregateStats(s.planIndexes)
}

func aggregateStats(planIndexes map[string]*search.SearchIndex) (totalCodes, totalRecords int) {
	seen := make(map[*search.SearchIndex]bool)
	for _, idx := range planIndexes {
		if !seen[idx] {
			seen[idx] = true
			totalCodes += idx.TotalCodes
			totalRecords += idx.TotalRecords
		}
	}
	return
}

func runIngestion(state *appState, indexURL string) {
	state.setStatus("fetching_index", "Fetching index file...")
	idx, _, err := ingest.FetchIndex(indexURL)
	if err != nil {
		state.setError(fmt.Errorf("index: %w", err))
		return
	}

	mappings := ingest.PlanFileMappings(idx)
	if len(mappings) == 0 {
		state.setError(fmt.Errorf("no in-network files found in index"))
		return
	}

	planIndexes := make(map[string]*search.SearchIndex)
	var allPlans []search.PlanInfo

	for i, m := range mappings {
		state.setStatus("fetching_rates", fmt.Sprintf("Fetching rate file %d of %d...", i+1, len(mappings)))

		rf, err := ingest.FetchRates(m.FileURL)
		if err != nil {
			state.setError(fmt.Errorf("rates (%s): %w", m.FileURL, err))
			return
		}

		planNames := make([]string, len(m.Plans))
		for j, p := range m.Plans {
			planNames[j] = p.PlanName
		}

		state.setStatus("building_index", fmt.Sprintf("Building index %d of %d...", i+1, len(mappings)))
		searchIdx := search.Build(rf, planNames, m.FileURL)

		for _, p := range m.Plans {
			allPlans = append(allPlans, search.PlanInfo{
				PlanID:         p.PlanID,
				PlanIDType:     p.PlanIDType,
				PlanName:       p.PlanName,
				PlanMarketType: p.PlanMarketType,
				IssuerName:     p.IssuerName,
			})
			planIndexes[p.PlanID] = searchIdx
		}
	}

	sort.Slice(allPlans, func(i, j int) bool {
		if allPlans[i].PlanName != allPlans[j].PlanName {
			return allPlans[i].PlanName < allPlans[j].PlanName
		}
		return allPlans[i].PlanID < allPlans[j].PlanID
	})

	state.setReady(allPlans, planIndexes)
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
	mux.HandleFunc("/api/plans", api.PlansHandler(state))
	mux.HandleFunc("/api/search", api.SearchHandler(state))

	addr := ":" + port
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, api.CORSMiddleware(mux)); err != nil {
		log.Fatalf("server: %v", err)
	}
}
