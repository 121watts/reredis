// Define message types to match the Go backend for type safety
export interface SetMessage {
  action: 'set'
  key: string
  value: string
}

export interface DelMessage {
  action: 'del'
  key: string
}

export interface SyncMessage {
  action: 'sync'
  data: Record<string, string>
}

export type ServerMessage = SetMessage | DelMessage | SyncMessage

export interface CommandMessage {
  action: 'set' | 'del' | 'get_all'
  key?: string
  value?: string
}
