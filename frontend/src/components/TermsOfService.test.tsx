import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { TermsOfService } from './TermsOfService'

describe('TermsOfService', () => {
  it('renders terms of service and mandatory YouTube Terms clause', () => {
    const onBack = vi.fn()
    render(<TermsOfService onBack={onBack} />)

    expect(screen.getByRole('heading', { level: 1, name: /terms of service/i })).toBeInTheDocument()
    expect(screen.getByText(/Third-Party Terms of Service/i)).toBeInTheDocument()

    // Mandatory YouTube Terms of Service link
    const ytLink = screen.getByRole('link', { name: /https:\/\/www\.youtube\.com\/t\/terms/i })
    expect(ytLink).toHaveAttribute('href', 'https://www.youtube.com/t/terms')

    // Back button
    fireEvent.click(screen.getByRole('button', { name: /back to converter/i }))
    expect(onBack).toHaveBeenCalled()
  })
})
