import { useState, useEffect } from 'react'
import { ArrowRightLeft, Loader2, CheckCircle, AlertCircle } from 'lucide-react'
import { apiClient } from './api'
import { useOAuth } from './hooks/useOAuth'
import { ProviderBox } from './components/ProviderBox'
import { SpotifyIcon, YouTubeMusicIcon, TidalIcon } from './components/Icons'
import type { ProviderInfo } from './gen/converter/v1/service_pb'
import type { CanonicalPlaylist } from './gen/converter/v1/model_pb'
import './App.css'



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

  const [isConverting, setIsConverting] = useState(false)
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

  const { tokens, setTokens, isAuthLoading, handleLogin } = useOAuth()

  // Fetch Playlists when Source Token or Provider changes
  useEffect(() => {
    const fetchPlaylists = async () => {
      const token = tokens[selectedSource]
      console.log("DEBUG START FETCH SOURCE:", selectedSource, token)
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
        console.log("DEBUG SUCCESS:", activePlaylists)
        setSourcePlaylists(activePlaylists)
        if (activePlaylists.length > 0 && !sourcePlaylistId) {
           setSourcePlaylistId(activePlaylists[0].id)
        }
      } catch(err: any) {
        console.error("DEBUG FETCH ERROR SOURCE:", err)
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
        <div className="supported-providers-banner">
          <span className="banner-text">Supported Platforms:</span>
          <div className="banner-icons">
            <SpotifyIcon className="banner-icon spotify" title="Spotify" />
            <YouTubeMusicIcon className="banner-icon youtube" title="YouTube Music" />
            <TidalIcon className="banner-icon tidal" title="Tidal" />
          </div>
        </div>
      </header>

      <main className="converter-card">
        <div className="provider-section">
          <ProviderBox
            type="source"
            providerId={selectedSource}
            providers={sources}
            onProviderChange={(id) => {
              setSelectedSource(id)
              setSourcePlaylistId("")
            }}
            playlistId={sourcePlaylistId}
            onPlaylistChange={setSourcePlaylistId}
            playlists={sourcePlaylists}
            isFetching={isFetchingSource}
            token={tokens[selectedSource]}
            onRefresh={() => setRefreshSource(r => r + 1)}
            onLogin={handleLogin}
          />

          {/* Exchange Icon */}
          <ArrowRightLeft className="exchange-icon" />

          <ProviderBox
            type="destination"
            providerId={selectedDest}
            providers={destinations}
            onProviderChange={(id) => {
              setSelectedDest(id)
              setDestPlaylistId("")
            }}
            playlistId={destPlaylistId}
            onPlaylistChange={setDestPlaylistId}
            playlists={destPlaylists}
            isFetching={isFetchingDest}
            token={tokens[selectedDest]}
            onRefresh={() => setRefreshDest(r => r + 1)}
            onLogin={handleLogin}
          />
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
