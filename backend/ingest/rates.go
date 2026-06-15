package ingest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// FetchRates downloads and parses the in-network rate file at the given URL.
func FetchRates(url string) (*RateFile, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching rates: HTTP %d", resp.StatusCode)
	}

	var rf RateFile
	if err := json.NewDecoder(resp.Body).Decode(&rf); err != nil {
		return nil, fmt.Errorf("parsing rates: %w", err)
	}

	warnProviderLocations(&rf)

	return &rf, nil
}

// warnProviderLocations logs a warning for any provider_references entry that uses
// a location URL instead of inline provider_groups. These are not currently resolved.
func warnProviderLocations(rf *RateFile) {
	for _, pr := range rf.ProviderReferences {
		if pr.Location != "" {
			log.Printf("WARNING: provider_group_id %d uses an external location URL (%s) — provider details will be missing for this group",
				pr.ProviderGroupID, pr.Location)
		}
	}
}
