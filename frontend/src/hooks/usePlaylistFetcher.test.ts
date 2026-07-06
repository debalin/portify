import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { usePlaylistFetcher } from './usePlaylistFetcher'

vi.mock('../api', () => ({
  apiClient: {
    listUserPlaylists: vi.fn()
  }
}))

import { apiClient } from '../api'

describe('usePlaylistFetcher', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('does not fetch playlists if token is not available', () => {
    const setTokens = vi.fn()
    const setSourcePlaylistId = vi.fn()

    const { result } = renderHook(() => usePlaylistFetcher({
      selectedSource: 'spotify',
      selectedDest: 'youtube',
      tokens: {},
      setTokens,
      sourcePlaylistId: '',
      setSourcePlaylistId
    }))

    expect(result.current.sourcePlaylists).toEqual([])
    expect(result.current.destPlaylists).toEqual([])
    expect(apiClient.listUserPlaylists).not.toHaveBeenCalled()
  })

  it('fetches source playlists and sets default playlist id', async () => {
    const mockPlaylists = [
      { id: 'pl-1', name: 'My Favorites', description: 'Favorites' }
    ]
    vi.mocked(apiClient.listUserPlaylists).mockResolvedValue({ playlists: mockPlaylists } as any)

    const setTokens = vi.fn()
    const setSourcePlaylistId = vi.fn()

    const { result } = renderHook(() => usePlaylistFetcher({
      selectedSource: 'spotify',
      selectedDest: 'youtube',
      tokens: { spotify: 'tok-spotify' },
      setTokens,
      sourcePlaylistId: '',
      setSourcePlaylistId
    }))

    await waitFor(() => {
      expect(result.current.sourcePlaylists).toEqual(mockPlaylists)
    })
    expect(setSourcePlaylistId).toHaveBeenCalledWith('pl-1')
  })

  it('handles 401 unauthenticated errors by clearing tokens', async () => {
    vi.mocked(apiClient.listUserPlaylists).mockRejectedValue(new Error('401 Unauthorized'))

    const setTokens = vi.fn()
    const setSourcePlaylistId = vi.fn()

    renderHook(() => usePlaylistFetcher({
      selectedSource: 'spotify',
      selectedDest: 'youtube',
      tokens: { spotify: 'tok-spotify' },
      setTokens,
      sourcePlaylistId: '',
      setSourcePlaylistId
    }))

    await waitFor(() => {
      expect(setTokens).toHaveBeenCalled()
    })
  })

  it('handles destination playlist fetching and triggers refreshes', async () => {
    const mockPlaylists = [
      { id: 'pl-dest', name: 'Destination Favorites', description: 'Desc' }
    ]
    vi.mocked(apiClient.listUserPlaylists).mockResolvedValue({ playlists: mockPlaylists } as any)

    const setTokens = vi.fn()
    const setSourcePlaylistId = vi.fn()

    const { result } = renderHook(() => usePlaylistFetcher({
      selectedSource: 'spotify',
      selectedDest: 'youtube',
      tokens: { youtube: 'tok-yt' },
      setTokens,
      sourcePlaylistId: '',
      setSourcePlaylistId
    }))

    await waitFor(() => {
      expect(result.current.destPlaylists).toEqual(mockPlaylists)
    })

    act(() => {
      result.current.triggerRefreshSource()
      result.current.triggerRefreshDest()
    })

    expect(result.current.isFetchingSource).toBe(false)
  })
})
