type CableProtocolError = {
  code?: string
  message?: string
}

type CableFrame = {
  type: 'event' | 'response' | string
  id?: string
  command?: string
  event?: string
  topic?: string
  ok?: boolean
  error?: CableProtocolError
  data?: any
  timestamp?: string
}

type EventHandler = (data: any, frame: CableFrame) => void

class CableService {
  private socket: WebSocket | null = null
  private pendingConnect: Promise<void> | null = null
  private shouldReconnect = false
  private reconnectAttempts = 0
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private releaseTimer: ReturnType<typeof setTimeout> | null = null
  private consumerCount = 0
  private readonly disconnectGraceMs = 1500
  private eventHandlers = new Map<string, Set<EventHandler>>()
  private responseWaiters = new Map<string, { resolve: (frame: CableFrame) => void; reject: (error: Error) => void }>()
  private subscriptions = new Set<string>()

  retain() {
    this.consumerCount += 1
    this.shouldReconnect = true

    if (this.releaseTimer) {
      clearTimeout(this.releaseTimer)
      this.releaseTimer = null
    }
  }

  release() {
    if (this.consumerCount > 0) {
      this.consumerCount -= 1
    }

    if (this.consumerCount > 0) {
      return
    }

    this.shouldReconnect = false

    if (this.releaseTimer) {
      clearTimeout(this.releaseTimer)
    }

    this.releaseTimer = setTimeout(() => {
      this.releaseTimer = null
      if (this.consumerCount === 0) {
        void this.disconnect()
      }
    }, this.disconnectGraceMs)
  }

  async connect(): Promise<void> {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      return
    }

    if (this.pendingConnect) {
      return this.pendingConnect
    }

    this.shouldReconnect = true
    this.pendingConnect = this.openSocket()

    try {
      await this.pendingConnect
    } finally {
      this.pendingConnect = null
    }
  }

  async disconnect(): Promise<void> {
    this.shouldReconnect = false
    this.reconnectAttempts = 0

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.releaseTimer) {
      clearTimeout(this.releaseTimer)
      this.releaseTimer = null
    }

    this.rejectPending(new Error('cable disconnected'))

    if (!this.socket) {
      return
    }

    const socket = this.socket
    this.socket = null

    if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
      socket.close()
    }
  }

  isConnected(): boolean {
    return this.socket?.readyState === WebSocket.OPEN
  }

  onEvent(event: string, handler: EventHandler): () => void {
    const handlers = this.eventHandlers.get(event) ?? new Set<EventHandler>()
    handlers.add(handler)
    this.eventHandlers.set(event, handlers)

    return () => {
      const current = this.eventHandlers.get(event)
      if (!current) {
        return
      }
      current.delete(handler)
      if (current.size === 0) {
        this.eventHandlers.delete(event)
      }
    }
  }

  async subscribeToken(token: string): Promise<void> {
    const normalized = token.trim()
    if (!normalized) {
      throw new Error('missing token')
    }

    this.subscriptions.add(normalized)
    await this.connect()
    await this.sendCommand('subscribe', { token: normalized })
  }

  async unsubscribeToken(token: string): Promise<void> {
    const normalized = token.trim()
    if (!normalized) {
      return
    }

    this.subscriptions.delete(normalized)
    if (!this.isConnected()) {
      return
    }

    await this.sendCommand('unsubscribe', { token: normalized })
  }

  private async openSocket(): Promise<void> {
    const url = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/cable`

    await new Promise<void>((resolve, reject) => {
      const socket = new WebSocket(url)
      let settled = false

      const cleanup = () => {
        socket.removeEventListener('open', handleOpen)
        socket.removeEventListener('error', handleError)
      }

      const handleOpen = () => {
        cleanup()
        settled = true
        const isReconnect = this.reconnectAttempts > 0
        this.socket = socket
        this.reconnectAttempts = 0
        resolve()
        if (isReconnect) {
          void this.restoreSubscriptions()
        }
      }

      const handleError = () => {
        cleanup()
        if (settled) {
          return
        }
        settled = true
        reject(new Error('failed to connect to cable'))
      }

      socket.addEventListener('open', handleOpen)
      socket.addEventListener('error', handleError)
      socket.addEventListener('message', (event) => {
        this.handleMessage(event.data)
      })
      socket.addEventListener('close', () => {
        if (this.socket === socket) {
          this.socket = null
        }
        this.rejectPending(new Error('cable connection closed'))
        this.scheduleReconnect()
      })
    })
  }

  private scheduleReconnect() {
    if (!this.shouldReconnect || this.pendingConnect) {
      return
    }

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }

    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 10000)
    this.reconnectAttempts += 1

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      void this.connect().catch(() => {
        this.scheduleReconnect()
      })
    }, delay)
  }

  private async restoreSubscriptions(): Promise<void> {
    for (const token of this.subscriptions) {
      try {
        await this.sendCommand('subscribe', { token })
      } catch {
        // Best effort during reconnect.
      }
    }
  }

  private handleMessage(raw: string) {
    let frame: CableFrame
    try {
      frame = JSON.parse(raw) as CableFrame
    } catch {
      return
    }

    if (frame.type === 'response' && frame.id) {
      const waiter = this.responseWaiters.get(frame.id)
      if (!waiter) {
        return
      }

      this.responseWaiters.delete(frame.id)
      if (frame.ok === false) {
        waiter.reject(new Error(frame.error?.message || 'cable command failed'))
        return
      }

      waiter.resolve(frame)
      return
    }

    if (frame.type === 'event' && frame.event) {
      const handlers = this.eventHandlers.get(frame.event)
      if (!handlers) {
        return
      }

      for (const handler of handlers) {
        try {
          handler(frame.data, frame)
        } catch {
          // Ignore handler errors so one listener does not break the stream.
        }
      }
    }
  }

  private async sendCommand(command: string, data?: any): Promise<CableFrame> {
    await this.connect()

    const socket = this.socket
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      throw new Error('cable is not connected')
    }

    const id = `${command}:${Date.now()}:${Math.random().toString(36).slice(2, 10)}`
    const payload = JSON.stringify({
      id,
      command,
      data,
    })

    return new Promise<CableFrame>((resolve, reject) => {
      this.responseWaiters.set(id, { resolve, reject })
      socket.send(payload)
    })
  }

  private rejectPending(error: Error) {
    if (this.responseWaiters.size === 0) {
      return
    }

    for (const [id, waiter] of this.responseWaiters.entries()) {
      this.responseWaiters.delete(id)
      waiter.reject(error)
    }
  }
}

const cableService = new CableService()
export default cableService
