import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ProviderBox } from './ProviderBox'
import type { ProviderInfo } from '../gen/converter/v1/service_pb'
import type { CanonicalPlaylist } from '../gen/converter/v1/model_pb'

describe('ProviderBox', () => {
  const mockProviders: ProviderInfo[] = [
    { id: 'spotify', name: 'Spotify', authUrlHint: '' } as ProviderInfo,
    { id: 'youtube', name: 'YouTube Music', authUrlHint: '' } as ProviderInfo,
  ]
  const mockPlaylists: CanonicalPlaylist[] = [
    { id: '1', name: 'Playlist 1', description: '' } as CanonicalPlaylist,
  ]

  const defaultProps = {
    type: 'source' as const,
    providerId: 'spotify',
    providers: mockProviders,
    onProviderChange: vi.fn(),
    playlistId: '',
    onPlaylistChange: vi.fn(),
    playlists: mockPlaylists,
    isFetching: false,
    token: undefined,
    onRefresh: vi.fn(),
    onLogin: vi.fn(),
  }

  it('renders login button when token is absent', () => {
    render(<ProviderBox {...defaultProps} />)
    const btn = screen.getByRole('button', { name: /login to connect/i })
    expect(btn).toBeDefined()
    fireEvent.click(btn)
    expect(defaultProps.onLogin).toHaveBeenCalledWith('spotify')
  })

  it('renders playlist picker and refresh button when token is present', () => {
    render(<ProviderBox {...defaultProps} token="mock-token" />)
    
    // Refresh button
    const refreshBtn = screen.getByTitle('Refresh Playlists')
    expect(refreshBtn).toBeDefined()
    fireEvent.click(refreshBtn)
    expect(defaultProps.onRefresh).toHaveBeenCalled()

    // Playlist picker
    expect(screen.getByText('Select a playlist...')).toBeDefined()
    expect(screen.getByText('Playlist 1')).toBeDefined()
  })

  it('handles provider change', () => {
    render(<ProviderBox {...defaultProps} />)
    const selects = screen.getAllByRole('combobox')
    const providerSelect = selects[0]
    fireEvent.change(providerSelect, { target: { value: 'youtube' } })
    expect(defaultProps.onProviderChange).toHaveBeenCalledWith('youtube')
  })

  it('handles playlist change', () => {
    render(<ProviderBox {...defaultProps} token="mock-token" />)
    const selects = screen.getAllByRole('combobox')
    const playlistSelect = selects[1]
    fireEvent.change(playlistSelect, { target: { value: '1' } })
    expect(defaultProps.onPlaylistChange).toHaveBeenCalledWith('1')
  })

  it('displays destination specific text', () => {
    render(<ProviderBox {...defaultProps} type="destination" token="mock-token" />)
    expect(screen.getByText('✨ Create New Playlist')).toBeDefined()
    expect(screen.getByText('Destination')).toBeDefined()
  })
})
