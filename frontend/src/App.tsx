import React, { useState, useEffect } from 'react'
import { ArrowRightLeft, Loader2, CheckCircle, AlertCircle, LogIn, RefreshCcw } from 'lucide-react'
import { apiClient } from './api'
import type { ProviderInfo } from './gen/converter/v1/service_pb'
import type { CanonicalPlaylist } from './gen/converter/v1/model_pb'
import './App.css'

const DefaultMusicIcon = ({ className }: { className?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}><path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/></svg>
)

const YouTubeMusicIcon = ({ className }: { className?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M12 0C5.376 0 0 5.376 0 12s5.376 12 12 12 12-5.376 12-12S18.624 0 12 0zm0 19.104c-3.924 0-7.104-3.18-7.104-7.104S8.076 4.896 12 4.896s7.104 3.18 7.104 7.104-3.18 7.104-7.104 7.104zm0-13.332c-3.432 0-6.228 2.796-6.228 6.228S8.568 18.228 12 18.228s6.228-2.796 6.228-6.228S15.432 5.772 12 5.772zM9.684 15.54V8.46L15.816 12l-6.132 3.54z"/>
  </svg>
)

const SpotifyIcon = ({ className }: { className?: string }) => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="1em" height="1em" 
    viewBox="0 0 24 24" 
    fill="currentColor" 
    className={className}
  >
    <path d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.6 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.54.659.3 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.6.18-1.2.72-1.38 4.2-1.26 11.28-1.02 15.72 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z" />
  </svg>
)

const TidalIcon = ({ className }: { className?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M12.01 5.334L8.683 2 5.341 5.334l3.327 3.322 3.342-3.322zM8.683 8.656L5.341 12l3.342 3.333 3.327-3.333-3.327-3.344zM12.01 12l-3.327 3.333L12.01 18.667l3.342-3.334L12.01 12zM15.353 8.656L12.01 12l3.342 3.333L18.678 12l-3.325-3.344zM18.678 5.334L15.353 2 12.01 5.334l3.343 3.322 3.325-3.322z"/>
  </svg>
)

const DynamicProviderIcon = ({ providerId, className }: { providerId: string, className?: string }) => {
  if (providerId === 'spotify') return <SpotifyIcon className={className} />
  if (providerId === 'youtube') return <YouTubeMusicIcon className={className} />
  if (providerId === 'tidal') return <TidalIcon className={className} />
  return <DefaultMusicIcon className={className} />
}

function App() {
  const [sources, setSources] = useState<ProviderInfo[]>([])
  const [destinations, setDestinations] = useState<ProviderInfo[]>([])
  const [selectedSource, setSelectedSourceState] = useState(() => sessionStorage.getItem('portifySource') || 'spotify')
  const [selectedDest, setSelectedDestState] = useState(() => sessionStorage.getItem('portifyDest') || 'youtube')
  
  const [sourcePlaylistId, setSourcePlaylistIdState] = useState(() => sessionStorage.getItem('portifyPlaylistId') || '')
  const [sourcePlaylists, setSourcePlaylists] = useState<CanonicalPlaylist[]>([])
  
  const [isFetchingSource, setIsFetchingSource] = useState(false)
  const [refreshSource, setRefreshSource] = useState(0)

  // Storage wrappers
  const setSelectedSource = (val: string) => { sessionStorage.setItem('portifySource', val); setSelectedSourceState(val) }
  const setSelectedDest = (val: string) => { sessionStorage.setItem('portifyDest', val); setSelectedDestState(val) }
  const setSourcePlaylistId = (val: string) => { sessionStorage.setItem('portifyPlaylistId', val); setSourcePlaylistIdState(val) }

  const [destPlaylistId, setDestPlaylistIdState] = useState(() => sessionStorage.getItem('portifyDestPlaylistId') || '')
  const [destPlaylists, setDestPlaylists] = useState<CanonicalPlaylist[]>([])
  
  const [isFetchingDest, setIsFetchingDest] = useState(false)
  const [refreshDest, setRefreshDest] = useState(0)

  // Storage wrappers for dest
  const setDestPlaylistId = (val: string) => { sessionStorage.setItem('portifyDestPlaylistId', val); setDestPlaylistIdState(val) }

  // Auth state: providerId -> accessToken (Persisted in sessionStorage to survive OAuth redirects)
  const [tokens, setTokensState] = useState<Record<string, string>>(() => {
    const saved = sessionStorage.getItem('portifyAuthTokens')
    return saved ? JSON.parse(saved) : {}
  })

  // Wrapper for setTokens to automatically save to sessionStorage
  const setTokens = (value: Record<string, string> | ((prev: Record<string, string>) => Record<string, string>)) => {
    setTokensState(prev => {
      const next = typeof value === 'function' ? value(prev) : value
      sessionStorage.setItem('portifyAuthTokens', JSON.stringify(next))
      return next
    })
  }
  
  const [isConverting, setIsConverting] = useState(false)
  const [isAuthLoading, setIsAuthLoading] = useState(false)
  const [result, setResult] = useState<{success: boolean, message: string, url?: string, failedTracks?: Array<any>} | null>(null)

  // Fetch registered providers on load
  useEffect(() => {
    async function loadProviders() {
      try {
        const res = await apiClient.listProviders({})
        setSources(res.sources)
        setDestinations(res.destinations)
      } catch (err) {
        console.error("Failed to load providers:", err)
      }
    }
    loadProviders()
  }, [])

  const authLatch = React.useRef(false)

  // Handle OAuth Redirect Callback
  useEffect(() => {
    const handleCallback = async () => {
      const params = new URLSearchParams(window.location.search)
      const code = params.get('code')
      const state = params.get('state') // We will pass providerId as state

      if (window.location.href.includes('code=')) {
        alert("Debug Callback! URL: " + window.location.href)
      }

      if (code && state && !authLatch.current) {
        authLatch.current = true
        console.log("Starting code exchange for", state)
        // Ensure browser address bar stops carrying the 1-time-use code
        window.history.replaceState({}, document.title, window.location.pathname)
        setIsAuthLoading(true)
        try {
          const res = await apiClient.exchangeAuthCode({
            providerId: state,
            code: code,
          })
          if (res.success) {
            setTokens(prev => ({ ...prev, [state]: res.accessToken }))
          } else {
            console.error("Auth failed:", res.errorMessage)
            alert("Authentication failed: " + res.errorMessage)
          }
        } catch (err: any) {
          console.error("Code exchange error:", err)
          alert("CRITICAL: Network Error during Code Exchange! " + (err.message || err))
        } finally {
          setIsAuthLoading(false)
        }
      }
    }
    handleCallback()
  }, [])

  // Fetch Playlists when Source Token or Provider changes
  useEffect(() => {
    const fetchPlaylists = async () => {
      const token = tokens[selectedSource]
      if (!token) {
        setSourcePlaylists([])
        return
      }
      setIsFetchingSource(true)
      try {
        const res = await apiClient.listUserPlaylists({
          providerId: selectedSource,
          accessToken: token
        })
        const activePlaylists = res.playlists || []
        setSourcePlaylists(activePlaylists)
        if (activePlaylists.length > 0 && !sourcePlaylistId) {
           setSourcePlaylistId(activePlaylists[0].id)
        }
      } catch(err: any) {
        console.error("Failed to fetch playlists", err)
        const msgLower = (err.message || err).toString().toLowerCase()
        if (msgLower.includes('401') || msgLower.includes('expired') || msgLower.includes('credential') || msgLower.includes('unauthorized')) {
            setTokens(prev => { const n = {...prev}; delete n[selectedSource]; return n })
        } else {
            alert("Failed to fetch playlists: " + (err.message || err))
        }
      } finally {
        setIsFetchingSource(false)
      }
    }
    fetchPlaylists()
  }, [selectedSource, tokens, refreshSource])

  // Fetch Playlists when Dest Token or Provider changes
  useEffect(() => {
    const fetchDestPlaylists = async () => {
      const token = tokens[selectedDest]
      if (!token) {
        setDestPlaylists([])
        return
      }
      setIsFetchingDest(true)
      try {
        const res = await apiClient.listUserPlaylists({
          providerId: selectedDest,
          accessToken: token
        })
        const activePlaylists = res.playlists || []
        setDestPlaylists(activePlaylists)
      } catch(err: any) {
        console.error("Failed to fetch dest playlists", err)
        const msgLower = (err.message || err).toString().toLowerCase()
        if (msgLower.includes('401') || msgLower.includes('expired') || msgLower.includes('credential') || msgLower.includes('unauthorized')) {
            setTokens(prev => { const n = {...prev}; delete n[selectedDest]; return n })
        } else {
            alert("Failed to fetch destination playlists: " + (err.message || err))
        }
      } finally {
        setIsFetchingDest(false)
      }
    }
    fetchDestPlaylists()
  }, [selectedDest, tokens, refreshDest])

  const handleLogin = async (providerId: string) => {
    try {
      const res = await apiClient.getAuthURL({ providerId })
      if (res.authUrl) {
         // Optionally append state if the backend doesn't do it automatically
         const finalUrl = new URL(res.authUrl, window.location.href)
         finalUrl.searchParams.set('state', providerId)
         window.location.href = finalUrl.toString()
      }
    } catch(err) {
      console.error("Failed to get auth URL", err)
      alert("Failed to initiate login.")
    }
  }

  const [progress, setProgress] = useState<{status: number, message: string, converted: number, total: number} | null>(null)

  const handleConvert = async () => {
    if (!sourcePlaylistId) {
      alert("Please enter a playlist ID")
      return
    }

    setIsConverting(true)
    setResult(null)
    setProgress(null)
    
    try {
      const stream = apiClient.convertPlaylist({
        sourceProvider: selectedSource,
        destinationProvider: selectedDest,
        sourcePlaylistId: sourcePlaylistId,
        destinationPlaylistId: destPlaylistId,
        sourceAuthToken: tokens[selectedSource],
        destinationAuthToken: tokens[selectedDest]
      })
      
      for await (const res of stream) {
        setProgress({
          status: res.status,
          message: res.message,
          converted: res.tracksConverted,
          total: res.tracksTotal
        })
        
        // STATUS_DONE = 3
        if (res.status === 3) {
           setResult({
             success: true,
             message: res.message,
             url: res.destinationPlaylistUrl,
             failedTracks: res.failedTracks
           })
        }
        
        // STATUS_ERROR = 4
        if (res.status === 4) {
           setResult({
             success: false,
             message: res.message
           })
           
           const msgLower = res.message.toLowerCase()
           const isAuthError = msgLower.includes('401') || msgLower.includes('expired') || msgLower.includes('invalid credential') || msgLower.includes('unauthorized') || msgLower.includes('autherror')
           
           if (isAuthError) {
              if (msgLower.includes('source playlist')) {
                 setTokens(prev => { const n = {...prev}; delete n[selectedSource]; return n })
              } else if (msgLower.includes('destination')) {
                 setTokens(prev => { const n = {...prev}; delete n[selectedDest]; return n })
              } else {
                 setTokens(prev => { const n = {...prev}; delete n[selectedSource]; delete n[selectedDest]; return n })
              }
           }
        }
      }
      
    } catch (err: any) {
      setResult({
        success: false,
        message: err.message || "An unknown error occurred"
      })
    } finally {
      setIsConverting(false)
    }
  }

  return (
    <div className="app-container">
      <header className="header">
        <h1 className="title">Portify</h1>
        <p className="subtitle">The Universal Playlist Converter</p>
      </header>

      <main className="converter-card">
        <div className="provider-section">
          {/* Source Box */}
          <div className={`provider-box source ${selectedSource}`}>
            {tokens[selectedSource] && (
              <button 
                onClick={(e) => { e.stopPropagation(); setRefreshSource(r => r+1) }}
                className="refresh-corner-btn"
                disabled={isFetchingSource}
                title="Refresh Playlists"
              >
                <RefreshCcw size={16} className={isFetchingSource ? "spinner" : ""} />
              </button>
            )}
            <DynamicProviderIcon providerId={selectedSource} className={`provider-icon ${selectedSource}`} />
            
            <select 
              className="provider-select"
              value={selectedSource} 
              onChange={e => {
                setSelectedSource(e.target.value);
                setSourcePlaylistId(""); // Clear selected playlist when provider changes
              }}
            >
              {sources.length > 0 ? sources.map(s => (
                <option key={s.id} value={s.id}>{s.name}</option>
              )) : <option value="spotify">Spotify</option>}
            </select>

            {!tokens[selectedSource] ? (
              <button className="login-btn" onClick={() => handleLogin(selectedSource)}>
                 <LogIn size={16} className="login-icon" /> Login to Connect
              </button>
            ) : (
              <div className="playlist-picker">
                 <select value={sourcePlaylistId} onChange={e => setSourcePlaylistId(e.target.value)} className="provider-select inner" disabled={isFetchingSource}>
                    {isFetchingSource ? <option value="">Fetching Playlists...</option> : <option value="" disabled>Select a playlist...</option>}
                    {sourcePlaylists.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                 </select>
              </div>
            )}
            
            <div className="provider-role">Source</div>
          </div>

          {/* Exchange Icon */}
          <ArrowRightLeft className="exchange-icon" />

          {/* Destination Box */}
          <div className={`provider-box destination ${selectedDest}`}>
            {tokens[selectedDest] && (
              <button 
                onClick={(e) => { e.stopPropagation(); setRefreshDest(r => r+1) }}
                className="refresh-corner-btn"
                disabled={isFetchingDest}
                title="Refresh Playlists"
              >
                <RefreshCcw size={16} className={isFetchingDest ? "spinner" : ""} />
              </button>
            )}
            <DynamicProviderIcon providerId={selectedDest} className={`provider-icon ${selectedDest}`} />
            
            <select 
              className="provider-select"
              value={selectedDest} 
              onChange={e => {
                setSelectedDest(e.target.value);
                setDestPlaylistId(""); // Clear selected playlist when provider changes
              }}
            >
              {destinations.length > 0 ? destinations.map(d => (
                <option key={d.id} value={d.id}>{d.name}</option>
              )) : <option value="youtube">YouTube Music</option>}
            </select>

            {!tokens[selectedDest] ? (
              <button className="login-btn youtube-login" onClick={() => handleLogin(selectedDest)}>
                 <LogIn size={16} className="login-icon" /> Login to Connect
              </button>
            ) : (
              <div className="playlist-picker">
                 <select value={destPlaylistId} onChange={e => setDestPlaylistId(e.target.value)} className="provider-select inner" disabled={isFetchingDest}>
                    {isFetchingDest ? <option value="">Fetching Playlists...</option> : <option value="">✨ Create New Playlist</option>}
                    {destPlaylists.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                 </select>
              </div>
            )}
            
            <div className="provider-role">Destination</div>
          </div>
        </div>


        {progress && progress.status > 0 && progress.status < 3 && (
          <div className="progress-container">
            <div className="progress-header">
              <span className="progress-status">{progress.message}</span>
              {progress.total > 0 && (
                <span className="progress-count">{Math.round((progress.converted / progress.total) * 100)}%</span>
              )}
            </div>
            {progress.total > 0 && (
              <div className="progress-bar-bg">
                <div 
                  className="progress-bar-fill" 
                  style={{ width: `${Math.round((progress.converted / progress.total) * 100)}%` }}
                ></div>
              </div>
            )}
          </div>
        )}

        {result && (
          <div className={`result-card ${result.success ? 'success' : 'error'}`}>
            {result.success ? <CheckCircle className="result-icon" /> : <AlertCircle className="result-icon error" />}
            <div>
              <p className="result-msg">{result.message}</p>
              {result.success && progress && progress.total > 0 && (
                  <p className="result-stats">Converted {progress.converted} of {progress.total} tracks.</p>
              )}
              {result.url && (
                <a href={result.url} target="portify_dest" rel="noreferrer" className="result-link">
                  {destPlaylistId ? 'Open Playlist' : 'Open New Playlist'} &rarr;
                </a>
              )}
              {result.failedTracks && result.failedTracks.length > 0 && (
                <details style={{marginTop: '12px', fontSize: '13px', background: 'rgba(255,100,100,0.1)', borderRadius: '8px', padding: '8px', border: '1px solid rgba(255,100,100,0.2)'}}>
                  <summary style={{cursor: 'pointer', color: '#ff6b6b', fontWeight: 600}}>
                     Failed to compile {result.failedTracks.length} tracks
                  </summary>
                  <ul style={{marginTop: '8px', paddingLeft: '20px', maxHeight: '150px', overflowY: 'auto', color: '#ffaaaa'}}>
                     {result.failedTracks.map((t, idx) => (
                       <li key={idx} style={{marginBottom: '4px'}}>{t.title} - {t.artist}</li>
                     ))}
                  </ul>
                </details>
              )}
            </div>
          </div>
        )}

        <button 
          className={`convert-btn ${isConverting ? 'loading' : ''}`} 
          onClick={handleConvert}
          disabled={isConverting || !sourcePlaylistId || !tokens[selectedSource] || !tokens[selectedDest] || isAuthLoading}
        >
          {isConverting ? (
            <span className="flex-center"><Loader2 className="spinner" /> Converting...</span>
          ) : 'Start Conversion'}
        </button>
      </main>

      {import.meta.env.VITE_SHOW_DEBUG_PANEL === 'true' && (
        <div style={{position: 'fixed', bottom: 0, left: 0, right: 0, background: 'darkred', color: 'white', padding: '10px', fontSize: '12px', zIndex: 9999}}>
          <strong>DEBUG PANEL:</strong> <br />
          Current tokens state: {JSON.stringify(tokens)} <br />
          Session Storage tokens: {sessionStorage.getItem('portifyAuthTokens')} <br />
          Selected Source: {selectedSource} | Selected Dest: {selectedDest} <br />
          URL params: {window.location.search}
        </div>
      )}
    </div>
  )
}

export default App
