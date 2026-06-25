import { onMounted, onUnmounted } from 'vue'
import cableService from '@/services/cable'

type CableHandler = (payload: any) => void

type CableEventBinding = {
  event: string
  handler: CableHandler
}

type UseCableSubscriptionOptions = {
  token?: string
  subscribeToken?: boolean
  onConnectError?: (error: unknown) => void
}

export function useCableSubscription(
  bindings: CableEventBinding[],
  options: UseCableSubscriptionOptions = {},
) {
  let unsubscribeHandlers: Array<() => void> = []

  onMounted(() => {
    cableService.retain()
    unsubscribeHandlers = bindings.map(({ event, handler }) => cableService.onEvent(event, handler))

    void cableService.connect()
      .then(async () => {
        if (options.subscribeToken && options.token) {
          await cableService.subscribeToken(options.token)
        }
      })
      .catch((error) => {
        options.onConnectError?.(error)
      })
  })

  onUnmounted(() => {
    for (const unsubscribe of unsubscribeHandlers) {
      unsubscribe()
    }
    unsubscribeHandlers = []

    const release = () => {
      cableService.release()
    }

    if (options.subscribeToken && options.token) {
      void cableService.unsubscribeToken(options.token)
        .catch(() => {
          // Ignore unsubscribe failures during teardown.
        })
        .finally(release)
      return
    }

    release()
  })

  return {
    isConnected: () => cableService.isConnected(),
  }
}
