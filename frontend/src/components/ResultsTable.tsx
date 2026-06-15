import { useState, useEffect } from 'react'
import { SearchResponse, SearchRecord } from '../api'

const PAGE_SIZE = 50

interface Props {
  response: SearchResponse
}

export default function ResultsTable({ response }: Props) {
  const [page, setPage] = useState(0)

  useEffect(() => {
    setPage(0)
  }, [response])

  const { results, billing_code, billing_code_type, description, result_count } = response
  const totalPages = Math.ceil(results.length / PAGE_SIZE)
  const pageResults = results.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE)

  return (
    <div className="space-y-3">
      <div className="flex items-baseline justify-between">
        <div>
          <span className="text-lg font-semibold text-gray-900">{billing_code}</span>
          {billing_code_type && (
            <span className="ml-2 text-sm text-gray-500">{billing_code_type}</span>
          )}
          {description && (
            <p className="mt-0.5 text-sm text-gray-600">{description}</p>
          )}
        </div>
        <span className="text-sm text-gray-500">
          {result_count.toLocaleString()} result{result_count !== 1 ? 's' : ''}
        </span>
      </div>

      {results.length === 0 ? (
        <div className="rounded-md border border-dashed border-gray-300 py-12 text-center text-sm text-gray-500">
          No results found for this billing code and filter combination.
        </div>
      ) : (
        <>
          <div className="overflow-x-auto rounded-lg border border-gray-200 shadow-sm">
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <Th>Business Name</Th>
                  <Th>NPI(s)</Th>
                  <Th>EIN</Th>
                  <Th>Rate</Th>
                  <Th>Type</Th>
                  <Th>Class</Th>
                  <Th>Setting</Th>
                  <Th>Service Codes</Th>
                  <Th>Modifiers</Th>
                  <Th>Expiration</Th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 bg-white">
                {pageResults.map((r, i) => (
                  <Row key={i} record={r} />
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between text-sm text-gray-600">
              <span>
                Page {page + 1} of {totalPages}
              </span>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage(p => p - 1)}
                  disabled={page === 0}
                  className="rounded border border-gray-300 px-3 py-1 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                <button
                  onClick={() => setPage(p => p + 1)}
                  disabled={page >= totalPages - 1}
                  className="rounded border border-gray-300 px-3 py-1 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}

function Th({ children }: { children: React.ReactNode }) {
  return (
    <th className="whitespace-nowrap px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-gray-500">
      {children}
    </th>
  )
}

function Row({ record }: { record: SearchRecord }) {
  const expiry =
    record.expiration_date === '9999-12-31' ? 'No expiration' : record.expiration_date

  return (
    <tr className="hover:bg-gray-50">
      <Td wrap className="break-words">{record.business_name || <span className="text-gray-400">—</span>}</Td>
      <Td wrap>
        {record.npis && record.npis.length > 0
          ? record.npis.join(', ')
          : <span className="text-gray-400">—</span>}
      </Td>
      <Td>{record.ein || <span className="text-gray-400">—</span>}</Td>
      <Td className="font-medium text-gray-900">{formatRate(record.negotiated_rate)}</Td>
      <Td>{record.negotiated_type}</Td>
      <Td>{record.billing_class || <span className="text-gray-400">—</span>}</Td>
      <Td>{record.setting || <span className="text-gray-400">—</span>}</Td>
      <Td>{record.service_codes?.join(', ') || <span className="text-gray-400">—</span>}</Td>
      <Td>{record.modifiers?.join(', ') || <span className="text-gray-400">—</span>}</Td>
      <Td className={expiry === 'No expiration' ? 'text-gray-400' : ''}>{expiry}</Td>
    </tr>
  )
}

function Td({ children, className = '', wrap = false }: { children: React.ReactNode; className?: string; wrap?: boolean }) {
  return (
    <td className={`${wrap ? 'whitespace-normal' : 'whitespace-nowrap'} px-4 py-3 text-gray-700 ${className}`}>{children}</td>
  )
}

function formatRate(rate: number): string {
  return `$${rate.toFixed(2)}`
}
