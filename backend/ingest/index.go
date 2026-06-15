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

// PlanFileMapping links a set of reporting plans to a single in-network rate file.
type PlanFileMapping struct {
	FileURL string
	Plans   []ReportingPlan
}

// PlanFileMappings returns one entry per unique in-network file URL, collecting all
// reporting plans that reference it. If the same URL appears in multiple
// reporting_structure entries, their plans are merged into a single entry.
func PlanFileMappings(idx *IndexFile) []PlanFileMapping {
	seen := make(map[string]int) // url → index in result
	var result []PlanFileMapping
	for _, rs := range idx.ReportingStructure {
		for _, f := range rs.InNetworkFiles {
			if f.Location == "" {
				continue
			}
			if i, ok := seen[f.Location]; ok {
				result[i].Plans = append(result[i].Plans, rs.ReportingPlans...)
			} else {
				seen[f.Location] = len(result)
				result = append(result, PlanFileMapping{
					FileURL: f.Location,
					Plans:   append([]ReportingPlan{}, rs.ReportingPlans...),
				})
			}
		}
	}
	return result
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
