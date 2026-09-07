import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { Footer } from './Footer'

describe('Footer', () => {
  it('renders YouTube API Services attribution and legal links', () => {
    const onNavigate = vi.fn()
    render(<Footer onNavigate={onNavigate} />)

    expect(screen.getByText(/YouTube API Services/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /privacy policy/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /terms of service/i })).toBeInTheDocument()

    const googleLink = screen.getByRole('link', { name: /google privacy policy/i })
    expect(googleLink).toHaveAttribute('href', 'https://www.google.com/policies/privacy')

    const ytLink = screen.getByRole('link', { name: /youtube terms of service/i })
    expect(ytLink).toHaveAttribute('href', 'https://www.youtube.com/t/terms')
  })

  it('calls onNavigate when internal legal buttons are clicked', () => {
    const onNavigate = vi.fn()
    render(<Footer onNavigate={onNavigate} />)

    fireEvent.click(screen.getByRole('button', { name: /privacy policy/i }))
    expect(onNavigate).toHaveBeenCalledWith('/privacy')

    fireEvent.click(screen.getByRole('button', { name: /terms of service/i }))
    expect(onNavigate).toHaveBeenCalledWith('/terms')
  })
})
