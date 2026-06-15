import { StatusResponse } from '../api'

interface Props {
  status: StatusResponse | null
}

const STATUS_LABELS: Record<string, string> = {
  idle: 'Starting up...',
  fetching_index: 'Fetching index file...',
  fetching_rates: 'Downloading rate file (this may take up to 30 seconds)...',
  building_index: 'Building search index...',
  ready: 'Ready',
  error: 'Error',
}

export default function StatusBanner({ status }: Props) {
  if (!status) {
    return (
      <div className="flex items-center gap-2 text-sm text-gray-500">
        <Spinner />
        <span>Connecting...</span>
      </div>
    )
  }

  if (status.status === 'ready') {
    return (
      <div className="rounded-md bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-800">
        <span className="font-medium">Ready.</span>{' '}
        {status.billing_codes_loaded !== undefined && (
          <>
            {status.billing_codes_loaded.toLocaleString()} billing codes &middot;{' '}
            {status.rate_records_loaded?.toLocaleString()} rate records loaded.
          </>
        )}
      </div>
    )
  }

  if (status.status === 'error') {
    return (
      <div className="rounded-md bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">
        <span className="font-medium">Ingestion error:</span> {status.error ?? status.message}
      </div>
    )
  }

  return (
    <div className="flex items-center gap-2 text-sm text-gray-600">
      <Spinner />
      <span>{STATUS_LABELS[status.status] ?? status.message}</span>
    </div>
  )
}

function Spinner() {
  return (
    <svg
      className="h-4 w-4 animate-spin text-blue-500"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path
        className="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
      />
    </svg>
  )
}
