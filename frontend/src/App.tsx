import { useEffect, useRef, useState } from 'react'
import { fetchStatus, fetchPlans, searchRates, StatusResponse, SearchResponse, PlanInfo } from './api'
import StatusBanner from './components/StatusBanner'
import SearchForm, { SearchParams } from './components/SearchForm'
import ResultsTable from './components/ResultsTable'

const POLL_INTERVAL_MS = 2000

export default function App() {
  const [status, setStatus] = useState<StatusResponse | null>(null)
  const [plans, setPlans] = useState<PlanInfo[]>([])
  const [searchResult, setSearchResult] = useState<SearchResponse | null>(null)
  const [searchError, setSearchError] = useState<string>('')
  const [isSearching, setIsSearching] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    pollRef.current = setInterval(async () => {
      try {
        const s = await fetchStatus()
        setStatus(s)
        if (s.status === 'ready' || s.status === 'error') {
          clearInterval(pollRef.current!)
          if (s.status === 'ready') {
            fetchPlans().then(r => setPlans(r.plans)).catch(() => {})
          }
        }
      } catch {
        // network not yet up — keep polling
      }
    }, POLL_INTERVAL_MS)

    return () => clearInterval(pollRef.current!)
  }, [])

  async function handleSearch({ code, npi, ein, planId }: SearchParams) {
    setIsSearching(true)
    setSearchError('')
    setSearchResult(null)
    try {
      const result = await searchRates(code, planId, npi || undefined, ein || undefined)
      setSearchResult(result)
    } catch (err) {
      setSearchError(err instanceof Error ? err.message : 'Search failed')
    } finally {
      setIsSearching(false)
    }
  }

  const isReady = status?.status === 'ready'

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 shadow-sm">
        <div className="mx-auto max-w-7xl px-4 py-4 sm:px-6 lg:px-8">
          <h1 className="text-xl font-bold text-gray-900">Price Transparency Explorer</h1>
          <p className="mt-0.5 text-sm text-gray-500">
            Search in-network negotiated rates &middot; Fidelis Care
          </p>
        </div>
      </header>

      <main className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8 space-y-6">
        <StatusBanner status={status} />

        {isReady && (
          <div className="rounded-lg bg-white border border-gray-200 shadow-sm px-6 py-5">
            <h2 className="text-sm font-semibold text-gray-700 mb-4">Search Rates</h2>
            <SearchForm
              onSearch={handleSearch}
              isLoading={isSearching}
              disabled={!isReady}
              plans={plans}
            />
          </div>
        )}

        {searchError && (
          <div className="rounded-md bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">
            {searchError}
          </div>
        )}

        {searchResult && <ResultsTable response={searchResult} />}
      </main>
    </div>
  )
}
