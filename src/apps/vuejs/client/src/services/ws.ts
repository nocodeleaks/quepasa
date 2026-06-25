type Handler = (payload: any) => void

class WSService {
  socket: WebSocket | null = null
  handlers: Record<string, Handler[]> = {}

  connect(path: string) {
    if (this.socket) this.disconnect()
    const url = (window.location.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host + path
    this.socket = new WebSocket(url)

    this.socket.onopen = () => console.log('WS connected to', url)
    this.socket.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data)
        this.emit('message', data)
      } catch (err) {
        this.emit('message', e.data)
      }
    }

    this.socket.onclose = () => console.log('WS closed')
    this.socket.onerror = (e) => console.error('WS error', e)
  }

  disconnect() {
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }
  }

  on(event: string, handler: Handler) {
    if (!this.handlers[event]) this.handlers[event] = []
    this.handlers[event].push(handler)
  }

  emit(event: string, payload: any) {
    const hs = this.handlers[event] || []
    hs.forEach(h => h(payload))
  }
}

const instance = new WSService()
export default instance
