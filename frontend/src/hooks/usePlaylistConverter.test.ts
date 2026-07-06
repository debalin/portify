import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { usePlaylistConverter } from './usePlaylistConverter'

vi.mock('../api', () => ({
  apiClient: {
    convertPlaylist: vi.fn()
  }
}))

import { apiClient } from '../api'

describe('usePlaylistConverter', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('alerts if playlist ID is missing', async () => {
    const originalAlert = window.alert
    window.alert = vi.fn()

    const setTokens = vi.fn()
    const { result } = renderHook(() => usePlaylistConverter({
      selectedSource: 'spotify',
      selectedDest: 'youtube',
      sourcePlaylistId: '',
      destPlaylistId: '',
      tokens: {},
      setTokens
    }))

    await act(async () => {
      await result.current.handleConvert()
    })

    expect(window.alert).toHaveBeenCalledWith("Please enter a playlist ID")
    expect(result.current.isConverting).toBe(false)
    window.alert = originalAlert
  })

  it('runs the full conversion stream and sets success result', async () => {
    vi.mocked(apiClient.convertPlaylist).mockImplementation(function* () {
      yield { status: 1, message: 'Fetching...', tracksConverted: 0, tracksTotal: 5 }
      yield { status: 2, message: 'Converting...', tracksConverted: 3, tracksTotal: 5 }
      yield {
        status: 3,
        message: 'Successfully converted.',
        tracksConverted: 5,
        tracksTotal: 5,
        destinationPlaylistUrl: 'https://playlist-url',
        failedTracks: []
      }
    } as any)

    const setTokens = vi.fn()
    const { result } = renderHook(() => usePlaylistConverter({
      selectedSource: 'spotify',
      selectedDest: 'youtube',
      sourcePlaylistId: 'pl-123',
      destPlaylistId: 'pl-dest-123',
      tokens: { spotify: 'tok-spotify', youtube: 'tok-yt' },
      setTokens
    }))

    await act(async () => {
      await result.current.handleConvert()
    })

    await waitFor(() => {
      expect(result.current.isConverting).toBe(false)
      expect(result.current.progress).toEqual({
        status: 3,
        message: 'Successfully converted.',
        converted: 5,
        total: 5
      })
      expect(result.current.result).toEqual({
        success: true,
        message: 'Successfully converted.',
        url: 'https://playlist-url',
        failedTracks: []
      })
    })
  })

  it('handles auth error in source playlist and clears source token', async () => {
    vi.mocked(apiClient.convertPlaylist).mockImplementation(function* () {
      yield { status: 4, message: '401 Unauthorized for source playlist' }
    } as any)

    const setTokens = vi.fn()
    const { result } = renderHook(() => usePlaylistConverter({
      selectedSource: 'spotify',
      selectedDest: 'youtube',
      sourcePlaylistId: 'pl-123',
      destPlaylistId: '',
      tokens: { spotify: 'tok-spotify', youtube: 'tok-yt' },
      setTokens
    }))

    await act(async () => {
      await result.current.handleConvert()
    })

    await waitFor(() => {
      expect(result.current.result?.success).toBe(false)
      expect(setTokens).toHaveBeenCalled()
    })
  })

  it('handles auth error in destination playlist and clears destination token', async () => {
    vi.mocked(apiClient.convertPlaylist).mockImplementation(function* () {
      yield { status: 4, message: '401 Unauthorized for destination' }
    } as any)

    const setTokens = vi.fn()
    const { result } = renderHook(() => usePlaylistConverter({
      selectedSource: 'spotify',
      selectedDest: 'youtube',
      sourcePlaylistId: 'pl-123',
      destPlaylistId: '',
      tokens: { spotify: 'tok-spotify', youtube: 'tok-yt' },
      setTokens
    }))

    await act(async () => {
      await result.current.handleConvert()
    })

    await waitFor(() => {
      expect(result.current.result?.success).toBe(false)
      expect(setTokens).toHaveBeenCalled()
    })
  })

  it('handles generic conversion error and preserves state', async () => {
    vi.mocked(apiClient.convertPlaylist).mockImplementation(function* () {
      yield { status: 4, message: 'Some generic network error' }
    } as any)

    const setTokens = vi.fn()
    const { result } = renderHook(() => usePlaylistConverter({
      selectedSource: 'spotify',
      selectedDest: 'youtube',
      sourcePlaylistId: 'pl-123',
      destPlaylistId: '',
      tokens: { spotify: 'tok-spotify', youtube: 'tok-yt' },
      setTokens
    }))

    await act(async () => {
      await result.current.handleConvert()
    })

    await waitFor(() => {
      expect(result.current.result).toEqual({
        success: false,
        message: 'Some generic network error'
      })
      expect(setTokens).not.toHaveBeenCalled()
    })
  })
})
