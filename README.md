# Price Transparency Explorer

## Overview

<!-- Brief description of what this app does and how it works end-to-end -->

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

<!-- Describe how data moves through the system:
     index file → rate file → in-memory index → API → UI -->

## Assumptions & Design Decisions

<!-- Walk through the key choices you made and why:
     - Which in-network file you picked and why
     - In-memory vs database
     - How you handle provider reference resolution
     - Any CMS format quirks you encountered
     - Anything else a reviewer should know -->

## Tradeoffs

<!-- What you gave up in the interest of time, and what the production version would look like -->

## What I'd Improve With More Time

<!-- Honest list of what's missing or rough -->
