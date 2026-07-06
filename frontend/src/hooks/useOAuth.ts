import { useState, useEffect, useRef } from 'react'
import { apiClient } from '../api'

export function useOAuth() {
  const [tokens, setTokensState] = useState<Record<string, string>>(() => {
    const saved = sessionStorage.getItem('portifyAuthTokens')
    return saved ? JSON.parse(saved) : {}
  })

  const [isAuthLoading, setIsAuthLoading] = useState(false)
  const authLatch = useRef(false)

  const setTokens = (value: Record<string, string> | ((prev: Record<string, string>) => Record<string, string>)) => {
    setTokensState(prev => {
      const next = typeof value === 'function' ? value(prev) : value
      sessionStorage.setItem('portifyAuthTokens', JSON.stringify(next))
      return next
    })
  }

  useEffect(() => {
    const handleCallback = async () => {
      const params = new URLSearchParams(window.location.search)
      const code = params.get('code')
      const state = params.get('state') // We pass providerId as state

      if (window.location.href.includes('code=')) {
        // alert("Debug Callback! URL: " + window.location.href)
      }

      if (code && state && !authLatch.current) {
        authLatch.current = true
        console.log("Starting code exchange for", state)
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

  const handleLogin = async (providerId: string) => {
    try {
      const res = await apiClient.getAuthURL({ providerId })
      if (res.authUrl) {
         const finalUrl = new URL(res.authUrl, window.location.href)
         finalUrl.searchParams.set('state', providerId)
         window.location.href = finalUrl.toString()
      }
    } catch (err) {
      console.error("Failed to get auth URL:", err)
      alert("Failed to initiate login flow.")
    }
  }

  return { tokens, setTokens, isAuthLoading, handleLogin }
}
