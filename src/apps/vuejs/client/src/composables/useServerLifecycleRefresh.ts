import { useCableSubscription } from '@/composables/useCableSubscription'

type UseServerLifecycleRefreshOptions = {
  token?: string
  onRefresh: () => Promise<void> | void
  onDeleted?: () => void
  onConnectError?: (error: unknown) => void
}

const serverLifecycleEvents = [
  'server.connected',
  'server.disconnected',
  'server.stopped',
  'server.logged_out',
] as const

export function useServerLifecycleRefresh(options: UseServerLifecycleRefreshOptions) {
  const matchesToken = (payload: any) => !options.token || payload?.token === options.token

  useCableSubscription(
    [
      ...serverLifecycleEvents.map((event) => ({
        event,
        handler: async (payload: any) => {
          if (!matchesToken(payload)) {
            return
          }

          try {
            await options.onRefresh()
          } catch {
            // Keep the current UI state if the refresh fails.
          }
        },
      })),
      {
        event: 'server.deleted',
        handler: async (payload: any) => {
          if (!matchesToken(payload)) {
            return
          }

          if (options.onDeleted) {
            options.onDeleted()
            return
          }

          try {
            await options.onRefresh()
          } catch {
            // Keep the current UI state if the refresh fails.
          }
        },
      },
    ],
    {
      onConnectError: options.onConnectError,
    },
  )
}
