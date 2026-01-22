import type { SSEEvent } from '../types'

type EventHandler = (event: SSEEvent) => void

class SSEService {
  private eventSource: EventSource | null = null
  private handlers: EventHandler[] = []
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private baseDelay = 1000

  connect() {
    if (this.eventSource) {
      this.disconnect()
    }

    const url = `${import.meta.env.VITE_API_URL || '/api/v1'}/events`
    this.eventSource = new EventSource(url, { withCredentials: true })

    this.eventSource.onopen = () => {
      console.log('SSE connected')
      this.reconnectAttempts = 0
    }

    this.eventSource.onmessage = (event) => {
      try {
        const data: SSEEvent = JSON.parse(event.data)
        this.handlers.forEach((handler) => handler(data))
      } catch (error) {
        console.error('Failed to parse SSE event:', error)
      }
    }

    this.eventSource.onerror = () => {
      console.error('SSE error, attempting reconnect...')
      this.eventSource?.close()
      this.attemptReconnect()
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnect attempts reached')
      return
    }

    const delay = this.baseDelay * Math.pow(2, this.reconnectAttempts)
    this.reconnectAttempts++

    setTimeout(() => {
      console.log(`Reconnecting... attempt ${this.reconnectAttempts}`)
      this.connect()
    }, delay)
  }

  disconnect() {
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }
    this.reconnectAttempts = 0
  }

  subscribe(handler: EventHandler) {
    this.handlers.push(handler)
    return () => {
      this.handlers = this.handlers.filter((h) => h !== handler)
    }
  }

  isConnected() {
    return this.eventSource?.readyState === EventSource.OPEN
  }
}

export const sseService = new SSEService()
