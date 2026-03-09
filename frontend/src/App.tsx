import { useState, useEffect } from 'react'
import { ArrowRightLeft, Music, Youtube, Loader2, CheckCircle, AlertCircle } from 'lucide-react'
import { apiClient } from './api'
import type { ProviderInfo } from './gen/converter/v1/service_pb'
import './App.css'

function App() {
  const [sources, setSources] = useState<ProviderInfo[]>([])
  const [destinations, setDestinations] = useState<ProviderInfo[]>([])
  const [selectedSource, setSelectedSource] = useState('spotify')
  const [selectedDest, setSelectedDest] = useState('youtube')
  
  const [sourcePlaylistId, setSourcePlaylistId] = useState('')
  
  const [isConverting, setIsConverting] = useState(false)
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
        sourcePlaylistId: sourcePlaylistId
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
            
            <div className="provider-role">Destination</div>
          </div>
        </div>

        <div className="input-group">
          <input 
            type="text" 
            placeholder="Paste Playlist ID or URL here..." 
            className="playlist-input"
            value={sourcePlaylistId}
            onChange={e => setSourcePlaylistId(e.target.value)}
          />
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
          disabled={isConverting || !sourcePlaylistId}
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
