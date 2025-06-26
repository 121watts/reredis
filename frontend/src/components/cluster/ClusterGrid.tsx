import React from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui'
import { NodeCard } from './NodeCard'
import type { ClusterNode } from '@/types/messages'

interface ClusterGridProps {
  nodes: ClusterNode[]
  totalSlots: number
  selectedNodeId?: string | null
  onNodeSelect?: (nodeId: string | null) => void
}

/**
 * Grid layout displaying all cluster nodes
 * Provides node selection and detailed view capabilities
 */
export const ClusterGrid: React.FC<ClusterGridProps> = ({
  nodes,
  totalSlots,
  selectedNodeId,
  onNodeSelect
}) => {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Cluster Nodes</CardTitle>
          <span className="text-sm text-slate-400">
            {nodes.length} nodes total
          </span>
        </div>
      </CardHeader>
      
      <CardContent>
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
          {nodes.map((node) => (
            <NodeCard
              key={node.id}
              node={node}
              totalSlots={totalSlots}
              isSelected={selectedNodeId === node.id}
              onSelect={() => onNodeSelect?.(
                selectedNodeId === node.id ? null : node.id
              )}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  )
}