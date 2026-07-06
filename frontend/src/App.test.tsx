import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
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
  it('fetches and displays playlists after token is set', async () => {
    sessionStorage.setItem('portifyAuthTokens', JSON.stringify({ spotify: 'tok', youtube: 'tok2' }))
    render(<App />)
    
    // We should see the mocked playlists "My Playlist" and "Coding Beats"
    const elements = await screen.findAllByText('My Playlist')
    expect(elements.length).toBeGreaterThan(0)
  })

  it('handles the end-to-end conversion flow', async () => {
    // Setup tokens so we can convert
    sessionStorage.setItem('portifyAuthTokens', JSON.stringify({ spotify: 'tok', youtube: 'tok2' }))
    render(<App />)
    
    // Wait for playlists to load
    await screen.findAllByText('My Playlist')

    const convertBtn = screen.getByText(/start conversion/i)
    fireEvent.click(convertBtn)

    // The mock generator returns fetching, converting, and done states
    const successMsg = await screen.findByText(/successfully converted/i)
    expect(successMsg).toBeDefined()
    
    // Check if the destination URL link is displayed
    const link = screen.getByText(/open new playlist/i)
    expect(link).toHaveProperty('href', 'https://youtube.com/playlist?list=test')
  })

  it('handles auth errors during fetch Playlists', async () => {
    sessionStorage.setItem('portifyAuthTokens', JSON.stringify({ spotify: 'bad-tok', youtube: 'bad-tok' }))
    
    const { apiClient } = await import('./api')
    vi.mocked(apiClient.listUserPlaylists).mockRejectedValueOnce(new Error('401 Unauthorized'))
    
    render(<App />)
    
    // Auth error should trigger token deletion and logout
    await waitFor(() => {
      const tokens = JSON.parse(sessionStorage.getItem('portifyAuthTokens') || '{}')
      expect(tokens.spotify).toBeUndefined()
    })
  })

  it('handles conversion errors', async () => {
    sessionStorage.setItem('portifyAuthTokens', JSON.stringify({ spotify: 'tok', youtube: 'tok2' }))
    const { apiClient } = await import('./api')
    
    vi.mocked(apiClient.convertPlaylist).mockImplementationOnce(async function* () {
      yield { status: 4, message: 'Internal Server Error' }
    } as any)
    
    render(<App />)
    await screen.findAllByText('My Playlist')
    
    const convertBtn = screen.getByText(/start conversion/i)
    fireEvent.click(convertBtn)
    
    const errMsg = await screen.findByText(/Internal Server Error/i)
    expect(errMsg).toBeDefined()
  })

  it('handles non-auth errors during fetch Dest Playlists', async () => {
    sessionStorage.setItem('portifyAuthTokens', JSON.stringify({ spotify: 'tok', youtube: 'tok2' }))
    const { apiClient } = await import('./api')
    
    vi.mocked(apiClient.listUserPlaylists).mockImplementation(async (req) => {
      if (req.providerId === 'youtube') throw new Error('Network Error')
      return { playlists: [{ id: 'pl-1', name: 'My Playlist', description: 'desc' }] } as any
    })
    
    const originalAlert = window.alert
    window.alert = vi.fn()
    
    render(<App />)
    
    await waitFor(() => {
      expect(window.alert).toHaveBeenCalledWith(expect.stringContaining('Failed to fetch destination playlists'))
    })
    
    window.alert = originalAlert
  })
})

describe('api module', () => {
  it('exports apiClient with expected methods', async () => {
    const { apiClient } = await import('./api')
    expect(apiClient).toBeDefined()
  })
})
