import { describe, it, expect, vi, beforeEach } from 'vitest'
import { fetchStatus, fetchPlans, searchRates } from './api'

function mockFetch(body: unknown, status = 200) {
  return vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
  })
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('fetchStatus', () => {
  it('returns parsed StatusResponse', async () => {
    const payload = {
      status: 'ready',
      message: 'Loaded 1470 billing codes',
      billing_codes_loaded: 1470,
      rate_records_loaded: 173375,
    }
    vi.stubGlobal('fetch', mockFetch(payload))

    const result = await fetchStatus()
    expect(result.status).toBe('ready')
    expect(result.billing_codes_loaded).toBe(1470)
    expect(result.rate_records_loaded).toBe(173375)
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', mockFetch({}, 500))
    await expect(fetchStatus()).rejects.toThrow('Status fetch failed: 500')
  })
})

describe('fetchPlans', () => {
  it('returns parsed PlansResponse', async () => {
    const payload = {
      plans: [
        { plan_id: 'PLAN-001', plan_name: 'Alpha Plan', plan_id_type: 'hios', plan_market_type: 'individual', issuer_name: 'Alpha' },
      ],
    }
    vi.stubGlobal('fetch', mockFetch(payload))

    const result = await fetchPlans()
    expect(result.plans).toHaveLength(1)
    expect(result.plans[0].plan_id).toBe('PLAN-001')
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', mockFetch({}, 503))
    await expect(fetchPlans()).rejects.toThrow('Plans fetch failed: 503')
  })
})

describe('searchRates', () => {
  it('builds query string with required code and plan_id', async () => {
    vi.stubGlobal('fetch', mockFetch({ billing_code: '99213', results: [], result_count: 0 }))
    await searchRates('99213', 'PLAN-001')
    const url = (fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string
    expect(url).toContain('code=99213')
    expect(url).toContain('plan_id=PLAN-001')
  })

  it('includes npi when provided', async () => {
    vi.stubGlobal('fetch', mockFetch({ billing_code: '99213', results: [], result_count: 0 }))
    await searchRates('99213', 'PLAN-001', '1902960099')
    const url = (fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string
    expect(url).toContain('npi=1902960099')
  })

  it('includes ein when provided', async () => {
    vi.stubGlobal('fetch', mockFetch({ billing_code: '99213', results: [], result_count: 0 }))
    await searchRates('99213', 'PLAN-001', undefined, '11-2700051')
    const url = (fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string
    expect(url).toContain('ein=11-2700051')
  })

  it('omits npi and ein when not provided', async () => {
    vi.stubGlobal('fetch', mockFetch({ billing_code: '99213', results: [], result_count: 0 }))
    await searchRates('99213', 'PLAN-001')
    const url = (fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string
    expect(url).not.toContain('npi=')
    expect(url).not.toContain('ein=')
  })

  it('throws with error message on non-ok response', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 503,
        json: () => Promise.resolve({ error: 'data is still loading' }),
      }),
    )
    await expect(searchRates('99213', 'PLAN-001')).rejects.toThrow('data is still loading')
  })
})
