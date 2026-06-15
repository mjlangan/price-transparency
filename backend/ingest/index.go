package ingest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// FetchIndex downloads and parses the CMS index (table of contents) file.
// It returns the parsed IndexFile and the URL of the first in-network rate file found.
func FetchIndex(url string) (*IndexFile, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("fetching index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fetching index: HTTP %d", resp.StatusCode)
	}

	var idx IndexFile
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		return nil, "", fmt.Errorf("parsing index: %w", err)
	}

	rateURL, err := firstInNetworkURL(&idx)
	if err != nil {
		return nil, "", err
	}

	return &idx, rateURL, nil
}

func firstInNetworkURL(idx *IndexFile) (string, error) {
	for _, rs := range idx.ReportingStructure {
		for _, f := range rs.InNetworkFiles {
			if f.Location != "" {
				return f.Location, nil
			}
		}
	}
	return "", fmt.Errorf("no in_network_files found in index")
}

// PlanNames returns a deduplicated list of plan names across all reporting structures.
func PlanNames(idx *IndexFile) []string {
	seen := make(map[string]bool)
	var names []string
	for _, rs := range idx.ReportingStructure {
		for _, p := range rs.ReportingPlans {
			if p.PlanName != "" && !seen[p.PlanName] {
				seen[p.PlanName] = true
				names = append(names, p.PlanName)
			}
		}
	}
	return names
}
