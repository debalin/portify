import { useState, useEffect } from 'react'
import { ArrowRightLeft, Music, Youtube, Loader2, CheckCircle, AlertCircle, LogIn } from 'lucide-react'
import { apiClient } from './api'
import type { ProviderInfo } from './gen/converter/v1/service_pb'
import type { CanonicalPlaylist } from './gen/converter/v1/model_pb'
import './App.css'

function App() {
  const [sources, setSources] = useState<ProviderInfo[]>([])
  const [destinations, setDestinations] = useState<ProviderInfo[]>([])
  const [selectedSource, setSelectedSource] = useState('spotify')
  const [selectedDest, setSelectedDest] = useState('youtube')
  
  const [sourcePlaylistId, setSourcePlaylistId] = useState('')
  const [sourcePlaylists, setSourcePlaylists] = useState<CanonicalPlaylist[]>([])
  
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
  const [result, setResult] = useState<{success: boolean, message: string, url?: string} | null>(null)

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

  // Handle OAuth Redirect Callback
  useEffect(() => {
    const handleCallback = async () => {
      const params = new URLSearchParams(window.location.search)
      const code = params.get('code')
      const state = params.get('state') // We will pass providerId as state

      if (code && state) {
        // Synchronously remove the code from the URL so React Strict Mode (which double-fires useEffects)
        // doesn't try to exchange the same code twice.
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
        } catch (err) {
          console.error("Code exchange error:", err)
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

      try {
        const res = await apiClient.listUserPlaylists({
          providerId: selectedSource,
          accessToken: token
        })
        setSourcePlaylists(res.playlists)
        if (res.playlists.length > 0) {
           setSourcePlaylistId(res.playlists[0].id)
        }
      } catch(err) {
        console.error("Failed to fetch playlists", err)
      }
    }
    fetchPlaylists()
  }, [selectedSource, tokens])

  const handleLogin = async (providerId: string) => {
    try {
      const res = await apiClient.getAuthURL({ providerId })
      if (res.authUrl) {
         // Optionally append state if the backend doesn't do it automatically
         const finalUrl = new URL(res.authUrl)
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
             url: res.destinationPlaylistUrl
           })
        }
        
        // STATUS_ERROR = 4
        if (res.status === 4) {
           setResult({
             success: false,
             message: res.message
           })
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
          <div className="provider-box source">
            <Music className="provider-icon spotify" />
            
            <select 
              className="provider-select"
              value={selectedSource} 
              onChange={e => setSelectedSource(e.target.value)}
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
                 <select value={sourcePlaylistId} onChange={e => setSourcePlaylistId(e.target.value)} className="provider-select inner">
                    <option value="" disabled>Select a playlist...</option>
                    {sourcePlaylists.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                 </select>
              </div>
            )}
            
            <div className="provider-role">Source</div>
          </div>

          {/* Exchange Icon */}
          <ArrowRightLeft className="exchange-icon" />

          {/* Destination Box */}
          <div className="provider-box destination">
            <Youtube className="provider-icon youtube" />
            
            <select 
              className="provider-select"
              value={selectedDest} 
              onChange={e => setSelectedDest(e.target.value)}
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
              <div className="logged-in-badge"><CheckCircle size={16} className="success-icon" /> Connected</div>
            )}
            
            <div className="provider-role">Destination</div>
          </div>
        </div>


        {progress && progress.status > 0 && progress.status < 3 && (
          <div className="progress-container">
            <div className="progress-header">
              <span className="progress-status">{progress.message}</span>
              {progress.total > 0 && (
                <span className="progress-count">{progress.converted} / {progress.total}</span>
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
                <a href={result.url} target="_blank" rel="noreferrer" className="result-link">
                  Open New Playlist &rarr;
                </a>
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
    </div>
  )
}

export default App
