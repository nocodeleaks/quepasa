import * as signalR from '@microsoft/signalr'

type MessageHandler = (message: any) => void

class SignalRService {
  private connection: signalR.HubConnection | null = null
  private messageHandlers: MessageHandler[] = []
  private token: string = ''
  private isConnecting: boolean = false
  private reconnectAttempts: number = 0
  private maxReconnectAttempts: number = 5

  /**
   * Connect to SignalR hub for a specific server token
   */
  async connect(serverToken: string): Promise<void> {
    if (this.isConnecting) {
      console.log('SignalR: Already connecting...')
      return
    }

    // If already connected to the same token, do nothing
    if (this.connection && this.token === serverToken && 
        this.connection.state === signalR.HubConnectionState.Connected) {
      console.log('SignalR: Already connected to', serverToken)
      return
    }

    // Disconnect from previous connection if exists
    await this.disconnect()

    this.isConnecting = true
    this.token = serverToken

    try {
      const url = `${window.location.origin}/signalr`
      
      this.connection = new signalR.HubConnectionBuilder()
        .withUrl(url)
        .withAutomaticReconnect({
          nextRetryDelayInMilliseconds: (retryContext) => {
            // Exponential backoff: 1s, 2s, 4s, 8s, 16s
            const delay = Math.min(1000 * Math.pow(2, retryContext.previousRetryCount), 16000)
            return delay
          }
        })
        .configureLogging(signalR.LogLevel.Information)
        .build()

      // Set up event handlers
      this.connection.on('message', (payload: any) => {
        console.log('SignalR: Received message', payload)
        this.messageHandlers.forEach(handler => {
          try {
            handler(payload)
          } catch (e) {
            console.error('SignalR: Error in message handler', e)
          }
        })
      })

      this.connection.onclose((error) => {
        console.log('SignalR: Connection closed', error)
      })

      this.connection.onreconnecting((error) => {
        console.log('SignalR: Reconnecting...', error)
      })

      this.connection.onreconnected((connectionId) => {
        console.log('SignalR: Reconnected with ID:', connectionId)
        // Re-register token after reconnection
        this.registerToken()
      })

      // Start connection
      await this.connection.start()
      console.log('SignalR: Connected successfully')

      // Register the token with the hub
      await this.registerToken()

      this.reconnectAttempts = 0
    } catch (error) {
      console.error('SignalR: Connection error', error)
      this.reconnectAttempts++
      
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        // Retry connection after delay
        const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 16000)
        console.log(`SignalR: Retrying in ${delay}ms...`)
        setTimeout(() => this.connect(serverToken), delay)
      }
    } finally {
      this.isConnecting = false
    }
  }

  /**
   * Register the server token with the SignalR hub
   */
  private async registerToken(): Promise<void> {
    if (this.connection && this.connection.state === signalR.HubConnectionState.Connected && this.token) {
      try {
        await this.connection.invoke('Token', this.token)
        console.log('SignalR: Token registered:', this.token)
      } catch (error) {
        console.error('SignalR: Error registering token', error)
      }
    }
  }

  /**
   * Disconnect from SignalR hub
   */
  async disconnect(): Promise<void> {
    if (this.connection) {
      try {
        await this.connection.stop()
        console.log('SignalR: Disconnected')
      } catch (error) {
        console.error('SignalR: Error disconnecting', error)
      }
      this.connection = null
    }
    this.token = ''
    this.reconnectAttempts = 0
  }

  /**
   * Add a message handler
   */
  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.push(handler)
    
    // Return unsubscribe function
    return () => {
      const index = this.messageHandlers.indexOf(handler)
      if (index > -1) {
        this.messageHandlers.splice(index, 1)
      }
    }
  }

  /**
   * Remove all message handlers
   */
  clearHandlers(): void {
    this.messageHandlers = []
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.connection?.state === signalR.HubConnectionState.Connected
  }

  /**
   * Get current connection state
   */
  getState(): signalR.HubConnectionState | null {
    return this.connection?.state ?? null
  }
}

// Singleton instance
const signalRService = new SignalRService()
export default signalRService
