import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { ResultCard } from './ResultCard'

describe('ResultCard', () => {
  it('returns null if result is null', () => {
    const { container } = render(<ResultCard result={null} progress={null} destPlaylistId="" />)
    expect(container.firstChild).toBeNull()
  })

  it('renders success card with redirect links and statistics', () => {
    render(
      <ResultCard 
        result={{ success: true, message: 'Done conversion', url: 'https://youtube.com/playlist' }}
        progress={{ converted: 5, total: 5 }}
        destPlaylistId=""
      />
    )
    expect(screen.getByText('Done conversion')).toBeInTheDocument()
    expect(screen.getByText('Converted 5 of 5 tracks.')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /open new playlist/i })).toHaveAttribute('href', 'https://youtube.com/playlist')
  })

  it('renders error state and collapsed failed tracks correctly', () => {
    render(
      <ResultCard 
        result={{ 
          success: false, 
          message: 'Error during matching',
          failedTracks: [{ title: 'Track A', artist: 'Artist A' }]
        }}
        progress={null}
        destPlaylistId="existing-pl"
      />
    )
    expect(screen.getByText('Error during matching')).toBeInTheDocument()
    expect(screen.getByText('Failed to compile 1 tracks')).toBeInTheDocument()
    expect(screen.getByText('Track A - Artist A')).toBeInTheDocument()
  })

  it('renders Open Liked Music text when destPlaylistId is LIKED_SONGS', () => {
    render(
      <ResultCard 
        result={{ success: true, message: 'Done conversion', url: 'https://music.youtube.com/playlist?list=LM' }}
        progress={{ converted: 5, total: 5 }}
        destPlaylistId="LIKED_SONGS"
      />
    )
    expect(screen.getByRole('link', { name: /open liked music/i })).toHaveAttribute('href', 'https://music.youtube.com/playlist?list=LM')
  })
})
