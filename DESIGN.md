# Design: Price Transparency Explorer

## Overview

Full-stack application that ingests a CMS Transparency in Coverage index file, fetches the referenced in-network rate file, exposes a search API, and renders results in a responsive UI.

**Stack**: Go backend (stdlib only) + React + TypeScript + Tailwind CSS frontend (Vite)

---

## Project Structure

```
price-transparency/
├── README.md
├── DESIGN.md
├── backend/
│   ├── go.mod                  # module price-transparency/backend, go 1.22, zero deps
│   ├── main.go                 # wires everything, AppState, HTTP server
│   ├── ingest/
│   │   ├── types.go            # all CMS JSON structs
│   │   ├── index.go            # fetch + parse TOC index file
│   │   └── rates.go            # fetch + parse in-network rate file
│   ├── search/
│   │   ├── types.go            # SearchRecord, SearchIndex, SearchResponse
│   │   └── index.go            # build in-memory index, Search method
│   └── api/
│       ├── middleware.go       # CORS
│       └── handlers.go         # /api/status, /api/search
└── frontend/
    ├── index.html
    ├── package.json
    ├── vite.config.ts          # proxy /api → localhost:8080
    └── src/
        ├── main.tsx
        ├── App.tsx             # polling, layout, shared state
        ├── api.ts              # typed fetch wrappers + interfaces
        └── components/
            ├── StatusBanner.tsx
            ├── SearchForm.tsx
            └── ResultsTable.tsx
```

---

## Data Flow

```
Startup
  main.go → go ingest.Run(state, INDEX_URL)
    1. HTTP GET index file (~185KB JSON)
       Confirmed structure: reporting_structure[0].in_network_files[0].location
       → https://...2026-04-28_centene-management-company-llc_fidelis-ex_in-network.json
    2. HTTP GET rate file (~86MB JSON, ~30s on slow connection)
    3. json.Unmarshal into RateFile
    4. search.Build() → denormalized SearchIndex
       If a provider_references entry contains a `location` URL instead of
       inline provider_groups, log a warning and skip that group.
    5. state.Status = "ready"

  http.ListenAndServe(":8080") [immediately, returns 503 until ready]

Search Request
  GET /api/search?code=99213&npi=1234567890&ein=11-2700051
    1. index.ByCode["99213"] → O(1) slice lookup
    2. Linear filter on NPI / EIN if provided
    3. Return JSON (limit 100 by default)

Frontend
  - Polls /api/status every 2s → shows progress banner
  - When ready: shows SearchForm
  - On submit: calls /api/search → renders ResultsTable
```

---

## CMS File Format

### Index File (table of contents)

```json
{
  "reporting_entity_name": "Fidelis Care",
  "reporting_structure": [{
    "reporting_plans": [{"plan_name": "...", "plan_id": "...", "issuer_name": "..."}],
    "in_network_files": [{"description": "...", "location": "https://..."}]
  }]
}
```

### In-Network Rate File

```json
{
  "provider_references": [{
    "provider_group_id": 1,
    "provider_groups": [{
      "npi": [1902960099],
      "tin": {"type": "ein", "value": "11-2700051", "business_name": "Morton Zinberg MD"}
    }]
  }],
  "in_network": [{
    "billing_code": "99213",
    "billing_code_type": "CPT",
    "description": "Office visit, established patient",
    "negotiated_rates": [{
      "provider_references": [1],
      "negotiated_prices": [{
        "negotiated_type": "negotiated",
        "negotiated_rate": 139.39,
        "expiration_date": "9999-12-31",
        "billing_class": "professional",
        "service_code": ["11"]
      }]
    }]
  }]
}
```

**Key linking**: `negotiated_rates[].provider_references` are `provider_group_id` values referencing `provider_references[].provider_group_id`.

---

## Go Data Structures

### CMS Types (`ingest/types.go`)

```go
type IndexFile struct {
    ReportingEntityName string               `json:"reporting_entity_name"`
    ReportingStructure  []ReportingStructure `json:"reporting_structure"`
    LastUpdatedOn       string               `json:"last_updated_on"`
    Version             string               `json:"version"`
}
type ReportingStructure struct {
    ReportingPlans []ReportingPlan `json:"reporting_plans"`
    InNetworkFiles []FileRef       `json:"in_network_files"`
}
type ReportingPlan struct {
    PlanName   string `json:"plan_name"`
    PlanID     string `json:"plan_id"`
    IssuerName string `json:"issuer_name"`
}
type FileRef struct {
    Description string `json:"description"`
    Location    string `json:"location"`
}

type RateFile struct {
    ProviderReferences []ProviderReference `json:"provider_references"`
    InNetwork          []InNetworkItem     `json:"in_network"`
    LastUpdatedOn      string              `json:"last_updated_on"`
}
type ProviderReference struct {
    ProviderGroupID int             `json:"provider_group_id"`
    ProviderGroups  []ProviderGroup `json:"provider_groups"`
}
type ProviderGroup struct {
    NPI []int64 `json:"npi"` // int64: NPI values exceed int32 max
    TIN TIN     `json:"tin"`
}
type TIN struct {
    Type         string `json:"type"`
    Value        string `json:"value"`
    BusinessName string `json:"business_name"`
}
type InNetworkItem struct {
    BillingCode     string           `json:"billing_code"`
    BillingCodeType string           `json:"billing_code_type"`
    Description     string           `json:"description"`
    Name            string           `json:"name"`
    NegotiatedRates []NegotiatedRate `json:"negotiated_rates"`
}
type NegotiatedRate struct {
    ProviderReferences []int             `json:"provider_references"`
    NegotiatedPrices   []NegotiatedPrice `json:"negotiated_prices"`
}
type NegotiatedPrice struct {
    NegotiatedType      string   `json:"negotiated_type"`
    NegotiatedRate      float64  `json:"negotiated_rate"`
    ExpirationDate      string   `json:"expiration_date"`
    BillingClass        string   `json:"billing_class"`
    ServiceCode         []string `json:"service_code"`
    BillingCodeModifier []string `json:"billing_code_modifier"`
    Setting             string   `json:"setting"`
}
```

### In-Memory Index (`search/types.go`)

```go
type SearchRecord struct {
    BillingCode     string   `json:"billing_code"`
    BillingCodeType string   `json:"billing_code_type"`
    Description     string   `json:"description"`
    ProviderGroupID int      `json:"provider_group_id"`
    NPIs            []int64  `json:"npis"`
    EIN             string   `json:"ein"`
    EINType         string   `json:"ein_type"`
    BusinessName    string   `json:"business_name"`
    NegotiatedRate  float64  `json:"negotiated_rate"`
    NegotiatedType  string   `json:"negotiated_type"`
    BillingClass    string   `json:"billing_class"`
    Setting         string   `json:"setting"`
    ServiceCodes    []string `json:"service_codes"`
    Modifiers       []string `json:"modifiers"`
    ExpirationDate  string   `json:"expiration_date"`
}

type SearchIndex struct {
    ByCode    map[string][]SearchRecord
    PlanNames []string
    FileURL   string
    LoadedAt  time.Time
    TotalCodes int
    TotalRecords int
}
```

**Build algorithm**:
1. Build `providerRefMap: map[int]ProviderReference` from `provider_group_id`
2. For each `InNetworkItem` → each `NegotiatedRate` → cross-product provider IDs × prices
3. Denormalize all provider info into each `SearchRecord`, append to `ByCode[billingCode]`

**Query**:
- `ByCode[code]` — O(1) map lookup
- Linear filter for NPI (parse string → int64, scan `record.NPIs`) and EIN (strip hyphens, case-insensitive)

---

## App State & Startup

```go
type AppState struct {
    mu      sync.RWMutex
    Status  string // "idle"|"fetching_index"|"fetching_rates"|"building_index"|"ready"|"error"
    Message string
    Error   string
    Index   *search.SearchIndex
}
```

- `INDEX_URL` env var overrides hardcoded default
- Background goroutine updates status; HTTP server starts immediately
- Returns HTTP 503 with `{"status":"..."}` until ingestion completes

---

## API

### `GET /api/status`
```json
{
  "status": "ready",
  "message": "Loaded 1470 billing codes, 173375 rate records",
  "file_url": "https://...",
  "loaded_at": "2026-06-14T10:30:00Z",
  "billing_codes_loaded": 1470,
  "rate_records_loaded": 173375
}
```

### `GET /api/search?code=99213&npi=1234567890&ein=11-2700051&limit=100`

| Param | Required | Notes |
|-------|----------|-------|
| `code` | Yes | Exact billing code match |
| `npi` | No | 10-digit NPI as string (avoids JS precision loss) |
| `ein` | No | With or without hyphen |
| `limit` | No | Default 100 |

**Response**:
```json
{
  "billing_code": "99213",
  "billing_code_type": "CPT",
  "description": "Office visit, established patient",
  "result_count": 47,
  "results": [...]
}
```

**Error responses**: 400 (missing code), 503 (still loading)

---

## Frontend Components

| Component | Responsibility |
|-----------|---------------|
| `App.tsx` | Polls `/api/status` every 2s via `useEffect`; holds search results in `useState`; layout |
| `StatusBanner.tsx` | Spinner + progress during load; success/error banner when settled |
| `SearchForm.tsx` | Billing code (required) + NPI + EIN inputs; calls `onSearch` prop callback |
| `ResultsTable.tsx` | `$` rate formatting; "No expiration" for `9999-12-31`; client-side pagination (50/page) |

**Vite proxy** (`vite.config.ts`): `/api → http://localhost:8080` — no CORS needed in dev.

---

## Design Decisions & Tradeoffs

| Decision | Rationale | Production Alternative |
|----------|-----------|------------------------|
| `json.Unmarshal` in one pass | Actual files are 86–177MB, not 50GB+; pragmatic for timebox | Streaming parser (jstream or hand-rolled token decoder) |
| Load only first in-network file (`fidelis-ex`) | `reporting_structure[0]` covers 88 plans; `fidelis-es` (6 plans) skipped | Loop over all `in_network_files`, merge indexes |
| Denormalized flat records | O(1) query path; ~50MB RAM tradeoff | Store nested CMS structs, join at query time |
| Exact billing code match | Assignment implies users know their code | Prefix search or description FTS with SQLite FTS5 |
| NPI as string in API | Avoids JavaScript integer precision loss for 10-digit values | Same |
| EIN strip-hyphen normalization | Data uses `11-2700051`; users may omit hyphen | Same |
| In-memory only | Acceptable per assignment spec | SQLite or Postgres with appropriate indexes |
| Poll for status (not WebSocket) | One-time load; polling is simpler | Server-Sent Events for streaming progress |
| `9999-12-31` → "No expiration" | CMS convention for perpetual rates | Same |
| Results unsorted | Avoids arbitrary default; insertion order is stable | Add `sort` query param (e.g. `rate_asc`, `rate_desc`) — `SearchIndex.Search` already returns a slice ready for `sort.Slice` |
| `provider_location` URL variant: warn + skip | Not present in Fidelis data; full support adds a nested fetch per provider group | Fetch each `location` URL, merge provider_groups into the ref map |
| Tailwind CSS | Clean utility-first styling with minimal setup overhead | Plain scoped CSS |

---

## Run Instructions

```bash
# Backend
cd backend
go run ./main.go
# Listens on :8080; INDEX_URL env var overrides default

# Frontend
cd frontend
npm install
npm run dev
# Opens on http://localhost:5173
```

---

## Unit Testing

### Backend (`go test ./...`)

**`search/index_test.go`** — highest value tests; exercises the core logic with fixture data:

| Test | What it covers |
|------|---------------|
| `TestBuild_ProviderResolution` | `Build()` correctly dereferences `provider_group_id` → NPI/EIN into flat records |
| `TestBuild_CrossProduct` | Multiple provider refs × multiple negotiated prices each produce a separate `SearchRecord` |
| `TestSearch_ByCodeOnly` | Returns all records for a billing code; unknown code returns empty slice |
| `TestSearch_FilterByNPI` | NPI filter matches; non-matching NPI returns empty |
| `TestSearch_FilterByEIN_WithHyphen` | `11-2700051` matches stored `11-2700051` |
| `TestSearch_FilterByEIN_WithoutHyphen` | `112700051` also matches (hyphen normalization) |
| `TestSearch_FilterByEIN_CaseInsensitive` | Mixed-case EIN matches |
| `TestSearch_Limit` | Limit param caps returned records |
| `TestSearch_UnknownCode` | Returns empty result, not error |

Use a small hand-crafted `RateFile` fixture (2–3 billing codes, 2–3 provider groups) defined inline in the test file — no file I/O.

**`ingest/index_test.go`** — parsing correctness:

| Test | What it covers |
|------|---------------|
| `TestParseIndex_ExtractsFirstInNetworkURL` | Unmarshals index JSON, returns first `in_network_files[0].location` |
| `TestParseIndex_MultipleReportingStructures` | Returns URL from first non-empty `in_network_files` across structures |
| `TestParseIndex_Empty` | Returns error when no `in_network_files` found |

Test via `http.NewRequest` + `httptest.NewServer` serving a fixture JSON string — no real network calls.

**`ingest/rates_test.go`** — parsing correctness:

| Test | What it covers |
|------|---------------|
| `TestParseRateFile_ProviderReferences` | Correct `provider_group_id`, NPI slice, TIN fields |
| `TestParseRateFile_InNetwork` | Correct `billing_code`, `billing_code_type`, nested rates/prices |
| `TestParseRateFile_NPIInt64` | Large NPI value (e.g. `1902960099`) parses without overflow |

**`api/handlers_test.go`** — HTTP contract using `httptest`:

| Test | What it covers |
|------|---------------|
| `TestStatusHandler_Loading` | Returns 200 with `"status":"fetching_rates"` when not ready |
| `TestStatusHandler_Ready` | Returns 200 with `billing_codes_loaded` and `rate_records_loaded` fields |
| `TestStatusHandler_Error` | Returns 200 with `"status":"error"` and error message |
| `TestSearchHandler_MissingCode` | Returns 400 |
| `TestSearchHandler_NotReady` | Returns 503 with status body |
| `TestSearchHandler_HappyPath` | Returns 200 with results for known code |
| `TestSearchHandler_EmptyResults` | Returns 200 with `result_count: 0` for unknown code |

### Frontend (`npm run test` via Vitest + React Testing Library)

**`api.test.ts`** — mock `globalThis.fetch`:

| Test | What it covers |
|------|---------------|
| `fetchStatus returns parsed StatusResponse` | Correct shape, numeric fields typed |
| `searchRates builds correct query string` | NPI/EIN params included only when provided |
| `searchRates omits undefined params` | No `?npi=undefined` in URL |

**`ResultsTable.test.tsx`** — most display logic lives here:

| Test | What it covers |
|------|---------------|
| `renders one row per result` | Row count matches `results.length` |
| `formats rate as $X.XX` | `139.39` → `$139.39` |
| `shows "No expiration" for 9999-12-31` | Display label substitution |
| `shows empty state when results is empty` | Empty state element visible |
| `paginates at 50 records` | Page 1 shows 50 rows; next button appears |

**`SearchForm.test.tsx`**:

| Test | What it covers |
|------|---------------|
| `submit button disabled when code is empty` | Validation gate |
| `calls onSearch with trimmed code` | Whitespace stripped before callback |
| `omits npi and ein from callback when blank` | Optional fields absent when inputs empty |

### Running Tests

```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm run test
```

---

## Verification Checklist

1. `curl http://localhost:8080/api/status` → transitions through states to `"ready"`
2. `curl "http://localhost:8080/api/search?code=99213"` → JSON with results
3. Add `&npi=<npi from results>` → fewer results
4. Add `&ein=11-2700051` → narrowed by EIN
5. Frontend: status banner shows progress, disappears when ready
6. UI search for `99213` → table renders with formatted rates
7. NPI/EIN filter inputs narrow results live on re-submit
8. Empty result state shows gracefully (e.g., code `XXXXX`)
9. Restart backend while frontend open → frontend shows loading state again
