import { FormEvent, useEffect, useState } from 'react'
import { PlanInfo } from '../api'

export interface SearchParams {
  code: string
  npi: string
  ein: string
  planId: string
}

interface Props {
  onSearch: (params: SearchParams) => void
  isLoading: boolean
  disabled: boolean
  plans: PlanInfo[]
}

export default function SearchForm({ onSearch, isLoading, disabled, plans }: Props) {
  const [code, setCode] = useState('')
  const [npi, setNpi] = useState('')
  const [ein, setEin] = useState('')
  const [planId, setPlanId] = useState('')

  useEffect(() => {
    if (plans.length > 0 && planId === '') {
      setPlanId(plans[0].plan_id)
    }
  }, [plans])

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const trimmedCode = code.trim()
    if (!trimmedCode || !planId) return
    onSearch({ code: trimmedCode, npi: npi.trim(), ein: ein.trim(), planId })
  }

  const canSubmit = code.trim().length > 0 && planId.length > 0 && !isLoading && !disabled

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label htmlFor="plan" className="block text-sm font-medium text-gray-700">
          Insurance Plan
        </label>
        <select
          id="plan"
          value={planId}
          onChange={e => setPlanId(e.target.value)}
          disabled={plans.length === 0 || disabled}
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
        >
          {plans.map(p => (
            <option key={p.plan_id} value={p.plan_id}>
              {p.plan_name} ({p.plan_id})
            </option>
          ))}
        </select>
      </div>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div>
          <label htmlFor="code" className="block text-sm font-medium text-gray-700">
            Billing Code <span className="text-red-500">*</span>
          </label>
          <input
            id="code"
            type="text"
            value={code}
            onChange={e => setCode(e.target.value)}
            placeholder="e.g. 99213"
            required
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
          />
        </div>
        <div>
          <label htmlFor="npi" className="block text-sm font-medium text-gray-700">
            NPI{' '}
            <span className="text-gray-400 font-normal">(optional)</span>
          </label>
          <input
            id="npi"
            type="text"
            value={npi}
            onChange={e => setNpi(e.target.value)}
            placeholder="10-digit NPI"
            maxLength={10}
            pattern="\d*"
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
          />
        </div>
        <div>
          <label htmlFor="ein" className="block text-sm font-medium text-gray-700">
            EIN{' '}
            <span className="text-gray-400 font-normal">(optional)</span>
          </label>
          <input
            id="ein"
            type="text"
            value={ein}
            onChange={e => setEin(e.target.value)}
            placeholder="e.g. 11-2700051"
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
          />
        </div>
      </div>
      <button
        type="submit"
        disabled={!canSubmit}
        className="inline-flex items-center gap-2 rounded-md bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {isLoading ? 'Searching...' : 'Search'}
      </button>
    </form>
  )
}
