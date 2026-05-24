import { useState } from 'react'
import { apiClient } from '../api'

interface UsePlaylistConverterProps {
  selectedSource: string
  selectedDest: string
  sourcePlaylistId: string
  destPlaylistId: string
  tokens: Record<string, string>
  setTokens: React.Dispatch<React.SetStateAction<Record<string, string>>>
}

export function usePlaylistConverter({
  selectedSource,
  selectedDest,
  sourcePlaylistId,
  destPlaylistId,
  tokens,
  setTokens
}: UsePlaylistConverterProps) {
  const [isConverting, setIsConverting] = useState(false)
  const [progress, setProgress] = useState<{ status: number; message: string; converted: number; total: number } | null>(null)
  const [result, setResult] = useState<{ success: boolean; message: string; url?: string; failedTracks?: any[] } | null>(null)

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
              setTokens(prev => { const n = { ...prev }; delete n[selectedSource]; return n })
            } else if (msgLower.includes('destination')) {
              setTokens(prev => { const n = { ...prev }; delete n[selectedDest]; return n })
            } else {
              setTokens(prev => { const n = { ...prev }; delete n[selectedSource]; delete n[selectedDest]; return n })
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

  return {
    isConverting,
    progress,
    result,
    handleConvert
  }
}
