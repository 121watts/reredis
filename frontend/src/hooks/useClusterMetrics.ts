import { useMemo } from 'react'
import type { ClusterNode } from '@/types/messages'

interface ClusterMetrics {
  activeNodes: number
  totalNodes: number
  totalKeys: number
  totalBytes: number
  avgKeysPerNode: number
  avgBytesPerNode: number
  maxKeys: number
  minKeys: number
  slotCoverage: number
  loadBalance: number
  overallHealth: number
}

/**
 * Custom hook to calculate cluster health metrics
 * Provides derived state from cluster nodes data
 */
export const useClusterMetrics = (
  nodes: ClusterNode[], 
  totalSlots: number, 
  isConnected: boolean
): ClusterMetrics => {
  return useMemo(() => {
    const activeNodes = nodes.filter(n => n.slotStart !== -1)
    const totalKeys = nodes.reduce((sum, node) => sum + (node.keyCount || 0), 0)
    const totalBytes = nodes.reduce((sum, node) => sum + (node.byteSize || 0), 0)
    
    // Calculate slot coverage
    const coveredSlots = activeNodes.reduce((sum, node) => 
      sum + (node.slotEnd - node.slotStart + 1), 0
    )
    const slotCoverage = totalSlots > 0 ? (coveredSlots / totalSlots) * 100 : 0
    
    // Calculate load balance
    const avgKeysPerNode = activeNodes.length > 0 ? totalKeys / activeNodes.length : 0
    const avgBytesPerNode = activeNodes.length > 0 ? totalBytes / activeNodes.length : 0
    const maxKeys = Math.max(...activeNodes.map(n => n.keyCount || 0), 0)
    const minKeys = activeNodes.length > 0 
      ? Math.min(...activeNodes.map(n => n.keyCount || 0)) 
      : 0
    const loadBalance = maxKeys > 0 ? (1 - (maxKeys - minKeys) / maxKeys) * 100 : 100
    
    // Overall health score
    const connectionHealth = isConnected ? 100 : 0
    const clusterHealth = activeNodes.length >= 3 ? 100 : (activeNodes.length / 3) * 100
    const overallHealth = (connectionHealth + clusterHealth + slotCoverage + loadBalance) / 4
    
    return {
      activeNodes: activeNodes.length,
      totalNodes: nodes.length,
      totalKeys,
      totalBytes,
      avgKeysPerNode: Math.round(avgKeysPerNode),
      avgBytesPerNode: Math.round(avgBytesPerNode),
      maxKeys,
      minKeys,
      slotCoverage,
      loadBalance,
      overallHealth
    }
  }, [nodes, totalSlots, isConnected])
}