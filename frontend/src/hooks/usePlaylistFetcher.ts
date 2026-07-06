import { useState, useEffect } from 'react'
import { apiClient } from '../api'
import type { CanonicalPlaylist } from '../gen/converter/v1/model_pb'

interface UsePlaylistFetcherProps {
  selectedSource: string
  selectedDest: string
  tokens: Record<string, string>
  setTokens: React.Dispatch<React.SetStateAction<Record<string, string>>>
  sourcePlaylistId: string
  setSourcePlaylistId: (val: string) => void
}

export function usePlaylistFetcher({
  selectedSource,
  selectedDest,
  tokens,
  setTokens,
  sourcePlaylistId,
  setSourcePlaylistId
}: UsePlaylistFetcherProps) {
  const [sourcePlaylists, setSourcePlaylists] = useState<CanonicalPlaylist[]>([])
  const [isFetchingSource, setIsFetchingSource] = useState(false)
  const [refreshSource, setRefreshSource] = useState(0)

  const [destPlaylists, setDestPlaylists] = useState<CanonicalPlaylist[]>([])
  const [isFetchingDest, setIsFetchingDest] = useState(false)
  const [refreshDest, setRefreshDest] = useState(0)

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
      } catch (err: any) {
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
  }, [selectedSource, tokens[selectedSource], refreshSource])

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
      } catch (err: any) {
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
  }, [selectedDest, tokens[selectedDest], refreshDest])

  const triggerRefreshSource = () => setRefreshSource(r => r + 1)
  const triggerRefreshDest = () => setRefreshDest(r => r + 1)

  return {
    sourcePlaylists,
    isFetchingSource,
    triggerRefreshSource,
    destPlaylists,
    isFetchingDest,
    triggerRefreshDest
  }
}
