import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import SearchForm from './SearchForm'

function setup(props?: Partial<React.ComponentProps<typeof SearchForm>>) {
  const onSearch = vi.fn()
  render(
    <SearchForm
      onSearch={onSearch}
      isLoading={false}
      disabled={false}
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

  it('calls onSearch with trimmed code', async () => {
    const { onSearch } = setup()
    await userEvent.type(screen.getByLabelText(/billing code/i), '  99213  ')
    await userEvent.click(screen.getByRole('button', { name: /search/i }))
    expect(onSearch).toHaveBeenCalledWith({ code: '99213', npi: '', ein: '' })
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
})
