import { useMemo } from 'react'
import type { ClusterNode } from '@/types/messages'

interface ClusterNodeHelpers {
  getSlotPercentage: (node: ClusterNode) => number
  getNodeStatus: (node: ClusterNode) => 'active' | 'waiting' | 'current'
  getStatusVariant: (node: ClusterNode) => 'success' | 'warning' | 'info'
  getHealthColor: (score: number) => string
  formatNodeId: (id: string, length?: number) => string
  isNodeActive: (node: ClusterNode) => boolean
}

/**
 * Custom hook providing utilities for working with cluster nodes
 * Encapsulates common node operations and formatting
 */
export const useClusterNodes = (totalSlots: number): ClusterNodeHelpers => {
  return useMemo(() => ({
    getSlotPercentage: (node: ClusterNode) => {
      if (node.slotStart === -1 || node.slotEnd === -1) return 0
      return ((node.slotEnd - node.slotStart + 1) / totalSlots) * 100
    },

    getNodeStatus: (node: ClusterNode) => {
      if (node.isCurrentNode) return 'current'
      if (node.slotStart === -1) return 'waiting'
      return 'active'
    },

    getStatusVariant: (node: ClusterNode) => {
      if (node.isCurrentNode) return 'info'
      if (node.slotStart === -1) return 'warning'
      return 'success'
    },

    getHealthColor: (score: number) => {
      if (score >= 80) return 'text-emerald-500'
      if (score >= 60) return 'text-amber-500'
      return 'text-red-500'
    },

    formatNodeId: (id: string, length = 12) => {
      return id.length > length ? `${id.slice(0, length)}...` : id
    },

    isNodeActive: (node: ClusterNode) => {
      return node.slotStart !== -1
    }
  }), [totalSlots])
}