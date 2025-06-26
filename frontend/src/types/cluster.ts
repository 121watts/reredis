/**
 * Cluster-specific type definitions
 * Provides better type safety and documentation for cluster-related data structures
 */

export interface ClusterMetrics {
  activeNodes: number
  totalNodes: number
  totalKeys: number
  avgKeysPerNode: number
  maxKeys: number
  minKeys: number
  slotCoverage: number
  loadBalance: number
  overallHealth: number
}

export interface ClusterNodeHelpers {
  getSlotPercentage: (node: ClusterNode) => number
  getNodeStatus: (node: ClusterNode) => NodeStatus
  getStatusVariant: (node: ClusterNode) => BadgeVariant
  getHealthColor: (score: number) => string
  formatNodeId: (id: string, length?: number) => string
  isNodeActive: (node: ClusterNode) => boolean
}

export type NodeStatus = 'active' | 'waiting' | 'current'
export type BadgeVariant = 'success' | 'warning' | 'info' | 'danger' | 'default' | 'outline'
export type ViewMode = 'grid' | 'topology'

// Re-export from messages for convenience
export type { ClusterNode, ClusterInfo } from '@/types/messages'