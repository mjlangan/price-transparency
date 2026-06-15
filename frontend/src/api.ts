export type IngestStatus =
  | 'idle'
  | 'fetching_index'
  | 'fetching_rates'
  | 'building_index'
  | 'ready'
  | 'error'

export interface StatusResponse {
  status: IngestStatus
  message: string
  error?: string
  file_url?: string
  loaded_at?: string
  billing_codes_loaded?: number
  rate_records_loaded?: number
}

export interface SearchRecord {
  billing_code: string
  billing_code_type: string
  description: string
  provider_group_id: number
  npis: number[]
  ein: string
  ein_type: string
  business_name: string
  negotiated_rate: number
  negotiated_type: string
  billing_class: string
  setting: string
  service_codes: string[]
  modifiers: string[]
  expiration_date: string
}

export interface SearchResponse {
  billing_code: string
  billing_code_type: string
  description: string
  result_count: number
  results: SearchRecord[]
  error?: string
}

export async function fetchStatus(): Promise<StatusResponse> {
  const res = await fetch('/api/status')
  if (!res.ok) throw new Error(`Status fetch failed: ${res.status}`)
  return res.json()
}

export async function searchRates(
  code: string,
  npi?: string,
  ein?: string,
  limit = 1000,
): Promise<SearchResponse> {
  const params = new URLSearchParams({ code, limit: String(limit) })
  if (npi) params.set('npi', npi)
  if (ein) params.set('ein', ein)
  const res = await fetch(`/api/search?${params}`)
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `Search failed: ${res.status}`)
  }
  return res.json()
}
