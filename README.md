# Price Transparency Explorer

## Overview

A simple full-stack application that fetches and parses in-network rates for specific billing codes using in-network files listed in a Transparency in Coverage index file. This information can be searched using the frontend by selecting an insurance plan and providing a billing code. The code was written with the assistance of Claude Code.

## Running the App

### Prerequisites

- Go 1.22+
- Node.js 18+

### Backend

```bash
cd backend
go run ./main.go
# Listens on :8080
# Optional: INDEX_URL=<url> go run ./main.go
```

### Frontend

```bash
cd frontend
npm install
npm run dev
# Opens on http://localhost:5173
```

### Tests

```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm run test
```

---

## Architecture & Data Flow

The application consists of a Golang backend providing API endpoints and a React/TypeScript frontend providing the user interface. On startup, the backend retrieves a hardcoded index file. Plans found in the index file are collected and matched to their associated rate files, which are then fetched and parsed. The data is kept in-memory using maps. When all the data is parsed, the backend signals to the frontend that it is ready via the /api/status endpoint.

The user selects a plan from a dropdown in the frontend, and then provides a billing code (and optionally an NPI and/or EIN). Matching data is then presented in a paginated table. The table can also be sorted by Business Name or Rate, so users can easily scan through for either a particular provider in the area or for the best rate for their plan.

## Assumptions & Design Decisions

<!-- Walk through the key choices you made and why:
     - Which in-network file you picked and why
     - In-memory vs database
     - How you handle provider reference resolution
     - Any CMS format quirks you encountered
     - Anything else a reviewer should know -->
- In-memory storage is used to keep things simple.
- Initially I went with the first rate file in the index for simplicity, then added the plan picker & modified the backend so it could fetch all available rate data and match the results to the selected plan.
- Provider references are resolved by storing a map of provider group ID to provider reference as the data store is being built.
- When building the data store, deduplication based on code, provider_group, rate details happens to avoid returning multiple rows with the same provider & price.
- I do not handle `provider_references` that use location URLs instead of inline `provider_groups`. A warning is printed on the backend's console if this situation is encountered.
- Plans with duplicate names but different IDs are kept separate in the picker because I am assuming they are represented multiple times in the index file for a reason.
- There is a hard limit of 1000 results fetched by the frontend.
- The plan selector is basically deciding which rate file's data gets searched.

## Tradeoffs & Future Improvements

<!-- What you gave up in the interest of time, and what the production version would look like -->
- A production version would use a more robust data store, such as a SQL datbase for storing the data. This also avoids having to fetch and parse the data every time the backend starts up.
- Better management for index files, possibly allowing multiple index files to be consumed.
- It would be deployed onto cloud infrastructure instead of just running locally.
- It could be containerized for easy deployment, as well.
- The plan picker can be a little confusing with multiple plans of the same name but with different IDs. Users would probably appreciate a better experience for selecting the correct plan.
- Ability to handle an unlimited number of search results instead of a hard cap of 1000.
- Infinite scroll UX instead of pagination.
- Proper handling of location URLs instead of inline `provider_groups`.
- Geographic data from another source could be matched against the business names, allowing users to filter businesses to within a radius of an address or zip code.
