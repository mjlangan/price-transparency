import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import SearchForm from './SearchForm'
import { PlanInfo } from '../api'

const testPlans: PlanInfo[] = [
  {
    plan_id: 'PLAN-001',
    plan_name: 'Alpha Plan',
    plan_id_type: 'hios',
    plan_market_type: 'individual',
    issuer_name: 'Alpha Insurer',
  },
  {
    plan_id: 'PLAN-002',
    plan_name: 'Beta Plan',
    plan_id_type: 'hios',
    plan_market_type: 'individual',
    issuer_name: 'Beta Insurer',
  },
]

function setup(props?: Partial<React.ComponentProps<typeof SearchForm>>) {
  const onSearch = vi.fn()
  render(
    <SearchForm
      onSearch={onSearch}
      isLoading={false}
      disabled={false}
      plans={testPlans}
      {...props}
    />,
  )
  return { onSearch }
}

describe('SearchForm', () => {
  it('submit button is disabled when code is empty', () => {
    setup()
    expect(screen.getByRole('button', { name: /search/i })).toBeDisabled()
  })

  it('submit button is enabled once code is typed', async () => {
    setup()
    await userEvent.type(screen.getByLabelText(/billing code/i), '99213')
    expect(screen.getByRole('button', { name: /search/i })).toBeEnabled()
  })

  it('calls onSearch with trimmed code and selected plan', async () => {
    const { onSearch } = setup()
    await userEvent.type(screen.getByLabelText(/billing code/i), '  99213  ')
    await userEvent.click(screen.getByRole('button', { name: /search/i }))
    expect(onSearch).toHaveBeenCalledWith({ code: '99213', npi: '', ein: '', planId: 'PLAN-001' })
  })

  it('includes npi and ein in callback when filled', async () => {
    const { onSearch } = setup()
    await userEvent.type(screen.getByLabelText(/billing code/i), '99213')
    await userEvent.type(screen.getByLabelText(/npi/i), '1902960099')
    await userEvent.type(screen.getByLabelText(/ein/i), '11-2700051')
    await userEvent.click(screen.getByRole('button', { name: /search/i }))
    expect(onSearch).toHaveBeenCalledWith({
      code: '99213',
      npi: '1902960099',
      ein: '11-2700051',
      planId: 'PLAN-001',
    })
  })

  it('submit button is disabled when isLoading is true', async () => {
    setup({ isLoading: true })
    await userEvent.type(screen.getByLabelText(/billing code/i), '99213')
    expect(screen.getByRole('button', { name: /searching/i })).toBeDisabled()
  })

  it('submit button is disabled when disabled prop is true', async () => {
    setup({ disabled: true })
    await userEvent.type(screen.getByLabelText(/billing code/i), '99213')
    expect(screen.getByRole('button', { name: /search/i })).toBeDisabled()
  })

  it('renders all plans in the plan picker', () => {
    setup()
    const options = screen.getAllByRole('option')
    expect(options).toHaveLength(testPlans.length)
    expect(options[0]).toHaveTextContent('Alpha Plan (PLAN-001)')
    expect(options[1]).toHaveTextContent('Beta Plan (PLAN-002)')
  })

  it('defaults to the first plan in the list', () => {
    setup()
    const select = screen.getByRole('combobox', { name: /insurance plan/i })
    expect((select as HTMLSelectElement).value).toBe('PLAN-001')
  })

  it('includes the selected plan id in onSearch when plan changes', async () => {
    const { onSearch } = setup()
    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: /insurance plan/i }),
      'PLAN-002',
    )
    await userEvent.type(screen.getByLabelText(/billing code/i), '99213')
    await userEvent.click(screen.getByRole('button', { name: /search/i }))
    expect(onSearch).toHaveBeenCalledWith(
      expect.objectContaining({ planId: 'PLAN-002' }),
    )
  })
})
