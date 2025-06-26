import React, { useState } from 'react'
import { Button } from '@/components/ui'
import { useURLState } from '@/hooks/useURLState'
import { ClusterOverview } from './ClusterOverview'
import { CurrentNodeSpotlight } from './CurrentNodeSpotlight'
import { ClusterGrid } from './ClusterGrid'
import { ClusterTopology } from './ClusterTopology'
import type { ClusterNode } from '@/types/messages'
import type { ViewMode } from '@/types/ui'

interface ClusterDashboardProps {
  nodes: ClusterNode[]
  currentNodeId: string
  totalSlots: number
  clusterSize: number
  isConnected: boolean
}

/**
 * Main cluster dashboard component
 * Orchestrates the display of cluster overview, current node details, and node grid
 */
export const ClusterDashboard: React.FC<ClusterDashboardProps> = ({
  nodes,
  currentNodeId,
  totalSlots,
  clusterSize,
  isConnected
}) => {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [viewMode, setViewMode] = useURLState<ViewMode>('view', 'grid', ['grid', 'topology'])
  
  const currentNode = nodes.find(node => node.id === currentNodeId)

  return (
    <div className="space-y-8">
      {/* Header with View Controls */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold text-white">Cluster Dashboard</h2>
          <p className="text-slate-400">Monitor your Redis cluster topology and performance</p>
        </div>
        
        <div className="flex items-center gap-2">
          <Button
            variant={viewMode === 'grid' ? 'primary' : 'secondary'}
            size="md"
            onClick={() => setViewMode('grid')}
          >
            📊 Grid View
          </Button>
          <Button
            variant={viewMode === 'topology' ? 'primary' : 'secondary'}
            size="md"
            onClick={() => setViewMode('topology')}
          >
            🔗 Topology
          </Button>
        </div>
      </div>

      {/* Cluster Overview Metrics */}
      <ClusterOverview
        nodes={nodes}
        totalSlots={totalSlots}
        clusterSize={clusterSize}
        isConnected={isConnected}
      />

      {/* Current Node Spotlight */}
      {currentNode && (
        <CurrentNodeSpotlight
          node={currentNode}
          totalSlots={totalSlots}
        />
      )}

      {/* Node Views */}
      {viewMode === 'grid' ? (
        <ClusterGrid
          nodes={nodes}
          totalSlots={totalSlots}
          selectedNodeId={selectedNodeId}
          onNodeSelect={setSelectedNodeId}
        />
      ) : (
        <ClusterTopology
          nodes={nodes}
          currentNodeId={currentNodeId}
          totalSlots={totalSlots}
          selectedNodeId={selectedNodeId}
          onNodeSelect={setSelectedNodeId}
        />
      )}
    </div>
  )
}