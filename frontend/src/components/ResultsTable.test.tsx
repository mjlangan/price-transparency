import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ResultsTable from './ResultsTable'
import { SearchResponse, SearchRecord } from '../api'

function makeRecord(overrides: Partial<SearchRecord> = {}): SearchRecord {
  return {
    billing_code: '99213',
    billing_code_type: 'CPT',
    description: 'Office visit',
    provider_group_id: 1,
    npis: [1902960099],
    ein: '11-2700051',
    ein_type: 'ein',
    business_name: 'Acme Medical',
    negotiated_rate: 139.39,
    negotiated_type: 'negotiated',
    billing_class: 'professional',
    setting: 'outpatient',
    service_codes: ['11'],
    modifiers: [],
    expiration_date: '9999-12-31',
    ...overrides,
  }
}

function makeResponse(results: SearchRecord[]): SearchResponse {
  return {
    billing_code: '99213',
    billing_code_type: 'CPT',
    description: 'Office visit',
    result_count: results.length,
    results,
  }
}

describe('ResultsTable', () => {
  it('renders one row per result', () => {
    render(<ResultsTable response={makeResponse([makeRecord(), makeRecord()])} />)
    const rows = screen.getAllByRole('row')
    // 1 header + 2 data rows
    expect(rows).toHaveLength(3)
  })

  it('formats rate as $X.XX', () => {
    render(<ResultsTable response={makeResponse([makeRecord({ negotiated_rate: 139.39 })])} />)
    expect(screen.getByText('$139.39')).toBeInTheDocument()
  })

  it('shows "No expiration" for 9999-12-31', () => {
    render(<ResultsTable response={makeResponse([makeRecord({ expiration_date: '9999-12-31' })])} />)
    expect(screen.getByText('No expiration')).toBeInTheDocument()
  })

  it('shows real date for non-perpetual expiration', () => {
    render(<ResultsTable response={makeResponse([makeRecord({ expiration_date: '2027-06-30' })])} />)
    expect(screen.getByText('2027-06-30')).toBeInTheDocument()
  })

  it('shows empty state when results is empty', () => {
    render(<ResultsTable response={makeResponse([])} />)
    expect(screen.getByText(/no results found/i)).toBeInTheDocument()
  })

  it('shows result count in header', () => {
    render(<ResultsTable response={makeResponse([makeRecord(), makeRecord()])} />)
    expect(screen.getByText(/2 results/i)).toBeInTheDocument()
  })

  it('paginates at 50 records — page 1 shows 50 rows and next button appears', async () => {
    const records = Array.from({ length: 55 }, (_, i) =>
      makeRecord({ business_name: `Provider ${i}` }),
    )
    render(<ResultsTable response={makeResponse(records)} />)
    const rows = screen.getAllByRole('row')
    // header + 50 data rows
    expect(rows).toHaveLength(51)
    expect(screen.getByRole('button', { name: /next/i })).toBeInTheDocument()
  })

  it('navigates to next page on Next click', async () => {
    const records = Array.from({ length: 55 }, (_, i) =>
      makeRecord({ business_name: `Provider ${i}` }),
    )
    render(<ResultsTable response={makeResponse(records)} />)
    await userEvent.click(screen.getByRole('button', { name: /next/i }))
    // page 2 has 5 rows
    const rows = screen.getAllByRole('row')
    expect(rows).toHaveLength(6) // header + 5
  })

  it('previous button is disabled on first page', () => {
    const records = Array.from({ length: 55 }, () => makeRecord())
    render(<ResultsTable response={makeResponse(records)} />)
    expect(screen.getByRole('button', { name: /previous/i })).toBeDisabled()
  })
})
