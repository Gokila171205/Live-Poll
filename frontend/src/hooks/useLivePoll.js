import { useState, useEffect, useRef, useCallback } from 'react'

function getWsBase() {
  if (import.meta.env.VITE_WS_BASE_URL) {
    return import.meta.env.VITE_WS_BASE_URL.replace(/\/+$/, '')
  }
  if (typeof window !== 'undefined' && (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')) {
    return 'ws://localhost:8080/api'
  }
  return 'wss://live-poll-v9go.onrender.com/api'
}

const WS_BASE = getWsBase()

/**
 * Custom hook for live WebSocket subscription to a specific poll.
 * Zero polling - strictly driven by real-time WebSocket updates from Go + Redis Pub/Sub.
 */
export function useLivePoll(pollId) {
  const [liveData, setLiveData] = useState(null)
  const [status, setStatus] = useState('connecting') // 'connecting' | 'connected' | 'disconnected' | 'error'
  const [lastUpdated, setLastUpdated] = useState(null)
  const [logs, setLogs] = useState([])

  const wsRef = useRef(null)
  const reconnectTimeoutRef = useRef(null)
  const retryCountRef = useRef(0)

  const addLog = useCallback((msg) => {
    const time = new Date().toLocaleTimeString()
    setLogs((prev) => [{ id: Math.random(), time, msg }, ...prev.slice(0, 19)])
  }, [])

  useEffect(() => {
    if (!pollId) {
      setStatus('disconnected')
      return
    }

    let isSubscribed = true

    function connect() {
      if (!isSubscribed) return

      // Clear any pending reconnect
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }

      setStatus('connecting')
      let effectiveWsBase = WS_BASE
      if (typeof window !== 'undefined' && window.location.protocol === 'https:' && effectiveWsBase.startsWith('ws://')) {
        effectiveWsBase = effectiveWsBase.replace(/^ws:\/\//, 'wss://')
      }
      const wsUrl = `${effectiveWsBase}/polls/${pollId}/live`
      
      try {
        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          if (!isSubscribed) return
          setStatus('connected')
          retryCountRef.current = 0
          addLog(`Connected to live stream for poll: ${pollId}`)
        }

        ws.onmessage = (event) => {
          if (!isSubscribed) return
          try {
            const data = JSON.parse(event.data)
            if (data.type === 'poll_results_updated') {
              setLiveData(data)
              setLastUpdated(new Date())
              addLog(`Real-time update: Total Votes = ${data.totalVotes}`)
            }
          } catch (err) {
            console.error('Failed to parse WebSocket frame:', err)
          }
        }

        ws.onerror = (err) => {
          if (!isSubscribed) return
          console.warn('[WebSocket Error]:', err)
          setStatus('error')
        }

        ws.onclose = (e) => {
          if (!isSubscribed) return
          setStatus('disconnected')

          // Stop retrying if clean close or max retries exceeded
          if (retryCountRef.current >= 15) {
            setStatus('error')
            addLog(`Max reconnection attempts reached. Please refresh.`)
            return
          }

          addLog(`Disconnected (code: ${e.code}). Reconnecting...`)

          // Exponential backoff capped at 8s
          const delay = Math.min(1000 * Math.pow(1.5, retryCountRef.current), 8000)
          retryCountRef.current += 1
          reconnectTimeoutRef.current = setTimeout(connect, delay)
        }
      } catch (err) {
        setStatus('error')
        addLog(`Connection attempt failed: ${err.message}`)
      }
    }

    connect()

    return () => {
      isSubscribed = false
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        // Disconnect cleanly
        wsRef.current.close(1000, 'Component unmounted')
        wsRef.current = null
      }
    }
  }, [pollId, addLog])

  return {
    liveData,
    status,
    lastUpdated,
    logs,
  }
}
