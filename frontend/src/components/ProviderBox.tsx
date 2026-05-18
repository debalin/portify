
import { LogIn, RefreshCcw } from 'lucide-react'
import { DynamicProviderIcon } from './Icons'
import type { ProviderInfo } from '../gen/converter/v1/service_pb'
import type { CanonicalPlaylist } from '../gen/converter/v1/model_pb'

interface ProviderBoxProps {
  type: 'source' | 'destination'
  providerId: string
  providers: ProviderInfo[]
  onProviderChange: (id: string) => void
  playlistId: string
  onPlaylistChange: (id: string) => void
  playlists: CanonicalPlaylist[]
  isFetching: boolean
  token?: string
  onRefresh: () => void
  onLogin: (id: string) => void
}

export function ProviderBox({
  type,
  providerId,
  providers,
  onProviderChange,
  playlistId,
  onPlaylistChange,
  playlists,
  isFetching,
  token,
  onRefresh,
  onLogin
}: ProviderBoxProps) {
  const isSource = type === 'source'
  
  return (
    <div className={`provider-box ${type} ${providerId}`}>
      {token && (
        <button 
          onClick={(e) => { e.stopPropagation(); onRefresh() }}
          className="refresh-corner-btn"
          disabled={isFetching}
          title="Refresh Playlists"
        >
          <RefreshCcw size={16} className={isFetching ? "spinner" : ""} />
        </button>
      )}
      <DynamicProviderIcon providerId={providerId} className={`provider-icon ${providerId}`} />
      
      <select 
        className="provider-select"
        value={providerId} 
        onChange={e => onProviderChange(e.target.value)}
      >
        {providers.length > 0 ? providers.map(p => (
          <option key={p.id} value={p.id}>{p.name}</option>
        )) : <option value={providerId}>{providerId.charAt(0).toUpperCase() + providerId.slice(1)}</option>}
      </select>

      {!token ? (
        <button className={`login-btn ${!isSource ? 'youtube-login' : ''}`} onClick={() => onLogin(providerId)}>
           <LogIn size={16} className="login-icon" /> Login to Connect
        </button>
      ) : (
        <div className="playlist-picker">
           <select value={playlistId} onChange={e => onPlaylistChange(e.target.value)} className="provider-select inner" disabled={isFetching}>
              {isFetching 
                 ? <option value="">Fetching Playlists...</option> 
                 : (isSource 
                     ? <option value="" disabled>Select a playlist...</option> 
                     : <option value="">✨ Create New Playlist</option>)
              }
              {playlists.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
           </select>
        </div>
      )}
      
      <div className="provider-role">{isSource ? 'Source' : 'Destination'}</div>
    </div>
  )
}
