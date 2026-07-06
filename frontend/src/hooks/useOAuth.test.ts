import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useOAuth } from './useOAuth'
import { apiClient } from '../api'

vi.mock('../api', () => ({
  apiClient: {
    exchangeAuthCode: vi.fn(),
    getAuthURL: vi.fn(),
  }
}))

describe('useOAuth', () => {
  const originalLocation = window.location

  beforeEach(() => {
    sessionStorage.clear()
    vi.clearAllMocks()
    window.history.replaceState({}, document.title, '/')
  })

  it('initializes tokens from sessionStorage', () => {
    sessionStorage.setItem('portifyAuthTokens', JSON.stringify({ spotify: 'token123' }))
    const { result } = renderHook(() => useOAuth())
    expect(result.current.tokens).toEqual({ spotify: 'token123' })
  })

  it('updates tokens and sessionStorage', () => {
    const { result } = renderHook(() => useOAuth())
    act(() => {
      result.current.setTokens({ youtube: 'token456' })
    })
    expect(result.current.tokens).toEqual({ youtube: 'token456' })
    expect(sessionStorage.getItem('portifyAuthTokens')).toBe(JSON.stringify({ youtube: 'token456' }))
  })

  it('initiates login flow via getAuthURL', async () => {
    vi.mocked(apiClient.getAuthURL).mockResolvedValue({ authUrl: 'https://auth.com/login' } as any)
    
    // Mock window.location.href assignment safely
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { ...originalLocation, href: 'http://localhost/' }
    })

    const { result } = renderHook(() => useOAuth())
    await act(async () => {
      await result.current.handleLogin('spotify')
    })

    expect(apiClient.getAuthURL).toHaveBeenCalledWith({ providerId: 'spotify' })
    expect(window.location.href).toContain('state=spotify')

    // Restore window.location
    Object.defineProperty(window, 'location', {
      writable: true,
      value: originalLocation
    })
  })

  it('handles successful oauth callback from URL params', async () => {
    window.history.replaceState({}, document.title, '/?code=auth_code_123&state=spotify')
    vi.mocked(apiClient.exchangeAuthCode).mockResolvedValue({ success: true, accessToken: 'new_token' } as any)
    
    const { result } = renderHook(() => useOAuth())
    
    await waitFor(() => {
      expect(apiClient.exchangeAuthCode).toHaveBeenCalledWith({ providerId: 'spotify', code: 'auth_code_123' })
      expect(result.current.tokens).toEqual({ spotify: 'new_token' })
    })
  })

  it('handles failed oauth callback gracefully', async () => {
    window.history.replaceState({}, document.title, '/?code=bad_code&state=spotify')
    vi.mocked(apiClient.exchangeAuthCode).mockResolvedValue({ success: false, errorMessage: 'invalid grant' } as any)
    
    // Mock alert
    const originalAlert = window.alert
    window.alert = vi.fn()

    const { result } = renderHook(() => useOAuth())
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(apiClient.exchangeAuthCode).toHaveBeenCalled()
    expect(result.current.tokens).toEqual({})
    expect(window.alert).toHaveBeenCalledWith(expect.stringContaining('Authentication failed'))

    window.alert = originalAlert
  })
})
