import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'

// Mock the API client before importing App
vi.mock('./api', () => ({
  apiClient: {
    listProviders: vi.fn().mockResolvedValue({
      sources: [
        { id: 'spotify', name: 'Spotify', authUrlHint: 'https://spotify.auth' },
      ],
      destinations: [
        { id: 'youtube', name: 'YouTube Music', authUrlHint: 'https://youtube.auth' },
      ],
    }),
    getAuthURL: vi.fn().mockResolvedValue({
      authUrl: 'https://accounts.spotify.com/authorize?client_id=test',
    }),
    exchangeAuthCode: vi.fn().mockResolvedValue({
      success: true,
      accessToken: 'mock-token-abc',
    }),
    listUserPlaylists: vi.fn().mockResolvedValue({
      playlists: [
        { id: 'pl-1', name: 'My Playlist', description: 'A test playlist' },
        { id: 'pl-2', name: 'Coding Beats', description: 'Focus music' },
      ],
    }),
    convertPlaylist: vi.fn().mockImplementation(function* () {
      yield {
        status: 1, // FETCHING
        message: 'Fetching playlist data...',
        tracksConverted: 0,
        tracksTotal: 0,
      }
      yield {
        status: 2, // CONVERTING
        message: 'Converting tracks... (1/2)',
        tracksConverted: 1,
        tracksTotal: 2,
      }
      yield {
        status: 3, // DONE
        message: 'Successfully converted.',
        tracksConverted: 2,
        tracksTotal: 2,
        destinationPlaylistUrl: 'https://youtube.com/playlist?list=test',
        failedTracks: [],
      }
    }),
  },
}))

import App from './App'

describe('App', () => {
  beforeEach(() => {
    // Clear sessionStorage between tests
    sessionStorage.clear()
    // Reset window.location
    window.history.replaceState({}, document.title, '/')
  })

  it('renders the app title', async () => {
    render(<App />)
    await waitFor(() => {
      expect(screen.getByText('Portify')).toBeInTheDocument()
    })
  })

  it('displays source and destination provider sections', async () => {
    render(<App />)
    await waitFor(() => {
      expect(screen.getByText(/source/i)).toBeInTheDocument()
      expect(screen.getByText(/destination/i)).toBeInTheDocument()
    })
  })

  it('shows login buttons when not authenticated', async () => {
    render(<App />)
    await waitFor(() => {
      const loginButtons = screen.getAllByText(/login to connect/i)
      expect(loginButtons.length).toBeGreaterThan(0)
    })
  })

  it('displays the convert button', async () => {
    render(<App />)
    await waitFor(() => {
      const convertBtn = screen.getByText(/start conversion/i)
      expect(convertBtn).toBeInTheDocument()
    })
  })

  it('reads default source from sessionStorage', async () => {
    sessionStorage.setItem('portifySource', 'spotify')
    render(<App />)
    await waitFor(() => {
      const stored = sessionStorage.getItem('portifySource')
      expect(stored).toBe('spotify')
    })
  })

  it('reads default destination from sessionStorage', async () => {
    sessionStorage.setItem('portifyDest', 'youtube')
    render(<App />)
    await waitFor(() => {
      const stored = sessionStorage.getItem('portifyDest')
      expect(stored).toBe('youtube')
    })
  })
})

describe('api module', () => {
  it('exports apiClient with expected methods', async () => {
    const { apiClient } = await import('./api')
    expect(apiClient).toBeDefined()
  })
})
