import { useState, useEffect } from 'react'
import { ArrowRightLeft, Loader2 } from 'lucide-react'
import { apiClient } from './api'
import { useOAuth } from './hooks/useOAuth'
import { ProviderBox } from './components/ProviderBox'
import { Header } from './components/Header'
import { ConversionProgress } from './components/ConversionProgress'
import { ResultCard } from './components/ResultCard'
import { Footer } from './components/Footer'
import { PrivacyPolicy } from './components/PrivacyPolicy'
import { TermsOfService } from './components/TermsOfService'
import { usePlaylistFetcher } from './hooks/usePlaylistFetcher'
import { usePlaylistConverter } from './hooks/usePlaylistConverter'
import type { ProviderInfo } from './gen/converter/v1/service_pb'
import './App.css'

function App() {
  const [sources, setSources] = useState<ProviderInfo[]>([])
  const [destinations, setDestinations] = useState<ProviderInfo[]>([])
  const [selectedSource, setSelectedSourceState] = useState(() => sessionStorage.getItem('portifySource') || 'spotify')
  const [selectedDest, setSelectedDestState] = useState(() => sessionStorage.getItem('portifyDest') || 'youtube')
  
  const [sourcePlaylistId, setSourcePlaylistIdState] = useState(() => sessionStorage.getItem('portifyPlaylistId') || '')
  const [destPlaylistId, setDestPlaylistIdState] = useState(() => sessionStorage.getItem('portifyDestPlaylistId') || '')

  // Lightweight route state for /privacy, /terms, and /
  const [currentPath, setCurrentPath] = useState(() => {
    const p = window.location.pathname.toLowerCase()
    if (p.includes('/privacy') || window.location.hash.toLowerCase().includes('/privacy')) return '/privacy'
    if (p.includes('/terms') || window.location.hash.toLowerCase().includes('/terms')) return '/terms'
    return '/'
  })

  useEffect(() => {
    const handlePopState = () => {
      const p = window.location.pathname.toLowerCase()
      if (p.includes('/privacy') || window.location.hash.toLowerCase().includes('/privacy')) setCurrentPath('/privacy')
      else if (p.includes('/terms') || window.location.hash.toLowerCase().includes('/terms')) setCurrentPath('/terms')
      else setCurrentPath('/')
    }
    window.addEventListener('popstate', handlePopState)
    window.addEventListener('hashchange', handlePopState)
    return () => {
      window.removeEventListener('popstate', handlePopState)
      window.removeEventListener('hashchange', handlePopState)
    }
  }, [])

  const navigate = (path: string) => {
    window.history.pushState({}, '', path)
    setCurrentPath(path)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  // Storage wrappers
  const setSelectedSource = (val: string) => { sessionStorage.setItem('portifySource', val); setSelectedSourceState(val) }
  const setSelectedDest = (val: string) => { sessionStorage.setItem('portifyDest', val); setSelectedDestState(val) }
  const setSourcePlaylistId = (val: string) => { sessionStorage.setItem('portifyPlaylistId', val); setSourcePlaylistIdState(val) }
  const setDestPlaylistId = (val: string) => { sessionStorage.setItem('portifyDestPlaylistId', val); setDestPlaylistIdState(val) }

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

  // Use Playlist Fetcher Hook
  const {
    sourcePlaylists,
    isFetchingSource,
    triggerRefreshSource,
    destPlaylists,
    isFetchingDest,
    triggerRefreshDest
  } = usePlaylistFetcher({
    selectedSource,
    selectedDest,
    tokens,
    setTokens,
    sourcePlaylistId,
    setSourcePlaylistId
  })

  // Use Playlist Converter Hook
  const {
    isConverting,
    progress,
    result,
    handleConvert
  } = usePlaylistConverter({
    selectedSource,
    selectedDest,
    sourcePlaylistId,
    destPlaylistId,
    tokens,
    setTokens
  })

  return (
    <div className="app-container">
      <Header />

      {currentPath === '/privacy' ? (
        <PrivacyPolicy onBack={() => navigate('/')} />
      ) : currentPath === '/terms' ? (
        <TermsOfService onBack={() => navigate('/')} />
      ) : (
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
              onRefresh={triggerRefreshSource}
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
              onRefresh={triggerRefreshDest}
              onLogin={handleLogin}
            />
          </div>

          <ConversionProgress progress={progress} />

          <ResultCard result={result} progress={progress} destPlaylistId={destPlaylistId} />

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
      )}

      <Footer onNavigate={navigate} />

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
