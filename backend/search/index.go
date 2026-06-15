package search

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"price-transparency/backend/ingest"
)

// Build constructs a SearchIndex from a parsed rate file.
func Build(rf *ingest.RateFile, planNames []string, fileURL string) *SearchIndex {
	// Map provider_group_id → ProviderReference for O(1) lookup during build.
	providerRefMap := make(map[int]ingest.ProviderReference, len(rf.ProviderReferences))
	for _, pr := range rf.ProviderReferences {
		providerRefMap[pr.ProviderGroupID] = pr
	}

	byCode := make(map[string][]SearchRecord)
	// seen deduplicates records with identical (code, provider_group, rate details).
	// The same billing code can appear multiple times in in_network under different
	// negotiation_arrangement values, producing duplicate records for the same provider+price.
	seen := make(map[string]bool)
	totalRecords := 0

	for _, item := range rf.InNetwork {
		for _, rate := range item.NegotiatedRates {
			for _, provGroupID := range rate.ProviderReferences {
				pr, ok := providerRefMap[provGroupID]
				if !ok {
					continue
				}

				for _, pg := range pr.ProviderGroups {
					ein := pg.TIN.Value
					einType := pg.TIN.Type
					bizName := pg.TIN.BusinessName

					for _, price := range rate.NegotiatedPrices {
						key := fmt.Sprintf("%s|%d|%.2f|%s|%s|%s|%s|%s|%s",
							item.BillingCode, provGroupID,
							price.NegotiatedRate, price.NegotiatedType,
							price.BillingClass, price.Setting, price.ExpirationDate,
							strings.Join(price.ServiceCode, ","),
							strings.Join(price.BillingCodeModifier, ","),
						)
						if seen[key] {
							continue
						}
						seen[key] = true

						rec := SearchRecord{
							BillingCode:     item.BillingCode,
							BillingCodeType: item.BillingCodeType,
							Description:     item.Description,
							ProviderGroupID: provGroupID,
							NPIs:            pg.NPI,
							EIN:             ein,
							EINType:         einType,
							BusinessName:    bizName,
							NegotiatedRate:  price.NegotiatedRate,
							NegotiatedType:  price.NegotiatedType,
							BillingClass:    price.BillingClass,
							Setting:         price.Setting,
							ServiceCodes:    price.ServiceCode,
							Modifiers:       price.BillingCodeModifier,
							ExpirationDate:  price.ExpirationDate,
						}
						byCode[item.BillingCode] = append(byCode[item.BillingCode], rec)
						totalRecords++
					}
				}
			}
		}
	}

	return &SearchIndex{
		ByCode:       byCode,
		PlanNames:    planNames,
		FileURL:      fileURL,
		LoadedAt:     time.Now(),
		TotalCodes:   len(byCode),
		TotalRecords: totalRecords,
	}
}

// Search returns records matching the billing code, optionally filtered by NPI and/or EIN.
// Results are returned in insertion order (unsorted). Callers may sort the returned slice.
func (idx *SearchIndex) Search(code, npiStr, einStr string, limit int) ([]SearchRecord, string, string) {
	records, ok := idx.ByCode[code]
	if !ok {
		return nil, "", ""
	}

	// Determine description and billing_code_type from the first record.
	description := records[0].Description
	codeType := records[0].BillingCodeType

	// Parse optional filters.
	var filterNPI int64
	if npiStr != "" {
		if n, err := strconv.ParseInt(npiStr, 10, 64); err == nil {
			filterNPI = n
		}
	}
	filterEIN := normalizeEIN(einStr)

	// Apply filters.
	filtered := make([]SearchRecord, 0, len(records))
	for _, r := range records {
		if filterNPI != 0 && !containsNPI(r.NPIs, filterNPI) {
			continue
		}
		if filterEIN != "" && normalizeEIN(r.EIN) != filterEIN {
			continue
		}
		filtered = append(filtered, r)
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, description, codeType
}

func containsNPI(npis []int64, target int64) bool {
	return slices.Contains(npis, target)
}

// normalizeEIN strips hyphens and lowercases for comparison.
func normalizeEIN(ein string) string {
	return strings.ToLower(strings.ReplaceAll(ein, "-", ""))
}
