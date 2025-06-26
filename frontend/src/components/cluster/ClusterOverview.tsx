import React from 'react'
import { Metric } from '@/components/ui'
import { useClusterMetrics } from '@/hooks/useClusterMetrics'
import { formatBytes } from '@/utils/formatBytes'
import type { ClusterNode } from '@/types/messages'

interface ClusterOverviewProps {
  nodes: ClusterNode[]
  totalSlots: number
  clusterSize: number
  isConnected: boolean
}

/**
 * Displays high-level cluster metrics in a clean grid layout
 * Shows total nodes, active nodes, total keys, and connection status
 */
export const ClusterOverview: React.FC<ClusterOverviewProps> = ({
  nodes,
  totalSlots,
  clusterSize,
  isConnected,
}) => {
  const metrics = useClusterMetrics(nodes, totalSlots, isConnected)

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6">
      <Metric
        label="Total Nodes"
        value={clusterSize}
        icon="🏢"
        variant="info"
      />

      <Metric
        label="Active Nodes"
        value={metrics.activeNodes}
        icon="✅"
        variant="success"
      />

      <Metric
        label="Total Keys"
        value={metrics.totalKeys}
        icon="🔑"
        variant="default"
      />

      <Metric
        label="Storage Used"
        value={formatBytes(metrics.totalBytes)}
        icon="💾"
        variant="default"
      />
    </div>
  )
}
