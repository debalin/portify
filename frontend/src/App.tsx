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
  
  // Auth state: providerId -> accessToken
  const [tokens, setTokens] = useState<Record<string, string>>({})
  
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
        setIsAuthLoading(true)
        try {
          const res = await apiClient.exchangeAuthCode({
            providerId: state,
            code: code,
          })
          if (res.success) {
            setTokens(prev => ({ ...prev, [state]: res.accessToken }))
            // Clean up the URL
            window.history.replaceState({}, document.title, window.location.pathname)
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

  const handleConvert = async () => {
    if (!sourcePlaylistId) {
      alert("Please enter a playlist ID")
      return
    }

    setIsConverting(true)
    setResult(null)
    
    try {
      const response = await apiClient.convertPlaylist({
        sourceProvider: selectedSource,
        destinationProvider: selectedDest,
        sourcePlaylistId: sourcePlaylistId,
        sourceAuthToken: tokens[selectedSource],
        destinationAuthToken: tokens[selectedDest]
      })
      
      setResult({
        success: response.success,
        message: response.message,
        url: response.destinationPlaylistUrl
      })
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


        {result && (
          <div className={`result-card ${result.success ? 'success' : 'error'}`}>
            {result.success ? <CheckCircle className="result-icon" /> : <AlertCircle className="result-icon error" />}
            <div>
              <p className="result-msg">{result.message}</p>
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
