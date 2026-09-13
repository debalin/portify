import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { PrivacyPolicy } from './PrivacyPolicy'

describe('PrivacyPolicy', () => {
  it('renders YouTube API Services disclosures and mandatory compliance links', () => {
    const onBack = vi.fn()
    render(<PrivacyPolicy onBack={onBack} />)

    expect(screen.getByRole('heading', { level: 1, name: /privacy policy/i })).toBeInTheDocument()
    expect(screen.getByText(/YouTube API Services Compliance & Disclosure/i)).toBeInTheDocument()

    // Mandatory Google Privacy Policy link
    const googleLink = screen.getByRole('link', { name: /https:\/\/www\.google\.com\/policies\/privacy/i })
    expect(googleLink).toHaveAttribute('href', 'https://www.google.com/policies/privacy')

    // Mandatory YouTube Terms of Service link
    const ytLink = screen.getByRole('link', { name: /https:\/\/www\.youtube\.com\/t\/terms/i })
    expect(ytLink).toHaveAttribute('href', 'https://www.youtube.com/t/terms')

    // Mandatory revocation link
    const revokeLink = screen.getByRole('link', { name: /https:\/\/myaccount\.google\.com\/permissions/i })
    expect(revokeLink).toHaveAttribute('href', 'https://myaccount.google.com/permissions')

    // Back button
    fireEvent.click(screen.getByRole('button', { name: /back to converter/i }))
    expect(onBack).toHaveBeenCalled()
  })

  it('renders required sections for Policies III.A.2e, III.A.2g, and III.E.4a-g', () => {
    render(<PrivacyPolicy onBack={vi.fn()} />)

    // Policy III.A.2e: Sharing disclosure
    expect(screen.getByText(/4\. How We Use, Process, and Share Your Information/i)).toBeInTheDocument()
    expect(screen.getByText(/do not sell, rent, trade, license, or share/i)).toBeInTheDocument()

    // Policy III.A.2g: Device storage & cookies
    expect(screen.getByText(/5\. Device Storage, Cookies, and Browser Technologies/i)).toBeInTheDocument()
    expect(screen.getByText(/portifyAuthTokens/i)).toBeInTheDocument()

    // Policy III.E.4a-g: Data refresh/deletion schedule
    expect(screen.getByText(/6\. API Data Storage, Refresh, and Deletion Lifecycle/i)).toBeInTheDocument()
    expect(screen.getByText(/Zero Persistent Storage \(No Caching\)/i)).toBeInTheDocument()
  })
})
