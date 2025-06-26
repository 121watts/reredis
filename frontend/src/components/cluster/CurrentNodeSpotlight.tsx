import React from 'react'
import { Card, CardContent, Badge } from '@/components/ui'
import { useClusterNodes } from '@/hooks/useClusterNodes'
import type { ClusterNode } from '@/types/messages'

interface CurrentNodeSpotlightProps {
  node: ClusterNode
  totalSlots: number
}

/**
 * Highlights the current node with detailed information
 * Shows node ID, address, slot range, and key count
 */
export const CurrentNodeSpotlight: React.FC<CurrentNodeSpotlightProps> = ({
  node,
  totalSlots
}) => {
  const { formatNodeId } = useClusterNodes(totalSlots)

  const nodeDetails = [
    {
      label: 'Node ID',
      value: formatNodeId(node.id)
    },
    {
      label: 'Network Address',
      value: `${node.host}:${node.port}`
    },
    {
      label: 'Slot Range',
      value: node.slotStart === -1 ? 'Unassigned' : `${node.slotStart.toLocaleString()}-${node.slotEnd.toLocaleString()}`
    },
    {
      label: 'Keys Stored',
      value: (node.keyCount || 0).toLocaleString()
    }
  ]

  return (
    <Card 
      variant="gradient" 
      className="bg-gradient-to-r from-emerald-500/10 via-blue-500/10 to-emerald-500/10 border-emerald-500/30"
    >
      <CardContent>
        <div className="flex items-center gap-3 mb-6">
          <div className="w-3 h-3 bg-emerald-400 rounded-full animate-pulse shadow-lg shadow-emerald-400/50" />
          <h3 className="text-xl font-bold text-white">This Node</h3>
          <Badge variant="success" size="sm">
            Current
          </Badge>
        </div>
        
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
          {nodeDetails.map((detail, index) => (
            <div key={index} className="text-center">
              <div className="text-sm text-slate-400 font-medium mb-2">
                {detail.label}
              </div>
              <div className="text-slate-200 font-mono text-sm bg-white/10 px-3 py-2 rounded-lg border border-white/20 truncate">
                {detail.value}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}