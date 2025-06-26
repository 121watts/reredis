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

export interface ClusterNode {
  id: string
  host: string
  port: string
  slotStart: number
  slotEnd: number
  keyCount?: number
  byteSize?: number
  isCurrentNode?: boolean
}

export interface ClusterInfoMessage {
  action: 'cluster_info'
  nodes: ClusterNode[]
  currentNodeId: string
  totalSlots: number
  clusterSize: number
}

export interface ClusterEventMessage {
  action: 'cluster_event'
  event: 'node_added' | 'node_removed' | 'slots_redistributed'
  nodeId?: string
  message: string
}

export interface ClusterStatsMessage {
  action: 'cluster_stats'
  nodes: ClusterNode[]
  currentNodeId: string
  totalSlots: number
  clusterSize: number
  totalKeys: number
}

export type ServerMessage = SetMessage | DelMessage | SyncMessage | ClusterInfoMessage | ClusterEventMessage | ClusterStatsMessage

export interface CommandMessage {
  action: 'set' | 'del' | 'get_all' | 'cluster_info'
  key?: string
  value?: string
}
