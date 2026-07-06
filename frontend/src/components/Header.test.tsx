import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { Header } from './Header'

describe('Header', () => {
  it('renders standard titles and descriptions', () => {
    render(<Header />)
    expect(screen.getByText('Portify')).toBeInTheDocument()
    expect(screen.getByText('The Universal Playlist Converter')).toBeInTheDocument()
    expect(screen.getByText('Supported Platforms:')).toBeInTheDocument()
  })

  it('renders all platform icons', () => {
    render(<Header />)
    expect(screen.getByTitle('Spotify')).toBeInTheDocument()
    expect(screen.getByTitle('YouTube Music')).toBeInTheDocument()
    expect(screen.getByTitle('Tidal')).toBeInTheDocument()
  })
})
