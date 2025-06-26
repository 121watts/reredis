import React from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui'
import { useClusterNodes } from '@/hooks/useClusterNodes'
import { formatBytes } from '@/utils/formatBytes'
import type { ClusterNode } from '@/types/messages'

interface ClusterTopologyProps {
  nodes: ClusterNode[]
  currentNodeId: string
  totalSlots: number
  selectedNodeId?: string | null
  onNodeSelect?: (nodeId: string | null) => void
}

/**
 * Network topology visualization showing cluster nodes in a circular layout
 * Displays nodes as connected circles with size based on slot distribution
 */
export const ClusterTopology: React.FC<ClusterTopologyProps> = ({
  nodes,
  currentNodeId,
  totalSlots,
  selectedNodeId,
  onNodeSelect
}) => {
  const { isNodeActive } = useClusterNodes(totalSlots)

  const getNodeColor = (node: ClusterNode) => {
    if (node.isCurrentNode) return 'bg-emerald-500 shadow-lg shadow-emerald-500/50'
    if (isNodeActive(node)) return 'bg-blue-500 shadow-lg shadow-blue-500/50'
    return 'bg-amber-500 shadow-lg shadow-amber-500/50'
  }

  const getNodeSize = (node: ClusterNode) => {
    if (!isNodeActive(node)) return 'w-12 h-12'
    const percentage = ((node.slotEnd - node.slotStart + 1) / totalSlots) * 100
    if (percentage > 40) return 'w-20 h-20'
    if (percentage > 20) return 'w-16 h-16'
    return 'w-12 h-12'
  }

  const getNodeIcon = (node: ClusterNode) => {
    if (node.isCurrentNode) return '🏠'
    if (isNodeActive(node)) return '💾'
    return '⏳'
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Network Topology</CardTitle>
          <div className="flex items-center gap-4 text-sm">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 bg-emerald-500 rounded-full shadow-lg shadow-emerald-500/50"></div>
              <span className="text-slate-300">Current Node</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 bg-blue-500 rounded-full shadow-lg shadow-blue-500/50"></div>
              <span className="text-slate-300">Active Node</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 bg-amber-500 rounded-full shadow-lg shadow-amber-500/50"></div>
              <span className="text-slate-300">Waiting Node</span>
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent>
        <div className="relative min-h-96 bg-gradient-to-br from-white/5 to-blue-500/10 rounded-lg p-8 backdrop-blur-sm border border-white/10">
          {/* Connection lines (drawn first, behind nodes) */}
          <svg className="absolute inset-0 w-full h-full pointer-events-none">
            {nodes.map((node, index) => {
              const angle = (index * 2 * Math.PI) / nodes.length
              const radius = 140
              const x = Math.cos(angle) * radius
              const y = Math.sin(angle) * radius
              
              // Calculate actual positions (center of container + offset)
              const centerX = 192; // half of min-h-96 (384px / 2)
              const centerY = 192;
              const nodeX = centerX + x;
              const nodeY = centerY + y;
              
              return (
                <line
                  key={node.id}
                  x1={centerX}
                  y1={centerY}
                  x2={nodeX}
                  y2={nodeY}
                  stroke="rgba(255, 255, 255, 0.2)"
                  strokeWidth="2"
                  strokeDasharray={!isNodeActive(node) ? "5,5" : "none"}
                />
              )
            })}
          </svg>

          {/* Center connection hub */}
          <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-10">
            <div className="w-16 h-16 bg-white/20 rounded-full flex items-center justify-center shadow-lg backdrop-blur-sm border border-white/20">
              <span className="text-2xl">🔗</span>
            </div>
            <div className="text-center mt-2 text-sm font-medium text-slate-300">
              Redis Cluster
            </div>
          </div>

          {/* Nodes arranged in circle */}
          {nodes.map((node, index) => {
            const angle = (index * 2 * Math.PI) / nodes.length
            const radius = 140
            const x = Math.cos(angle) * radius
            const y = Math.sin(angle) * radius
            
            return (
              <div
                key={node.id}
                className="group absolute transform -translate-x-1/2 -translate-y-1/2 cursor-pointer"
                style={{
                  left: `calc(50% + ${x}px)`,
                  top: `calc(50% + ${y}px)`
                }}
                onClick={() => onNodeSelect?.(
                  selectedNodeId === node.id ? null : node.id
                )}
              >
                {/* Node circle */}
                <div
                  className={`${getNodeSize(node)} ${getNodeColor(node)} rounded-full flex items-center justify-center transition-all duration-200 hover:scale-110 ${
                    selectedNodeId === node.id ? 'ring-4 ring-blue-400/50' : ''
                  }`}
                >
                  <span className="text-white font-bold text-xs">
                    {getNodeIcon(node)}
                  </span>
                </div>

                {/* Node info popup - only visible on hover */}
                <div className="absolute top-full mt-2 left-1/2 transform -translate-x-1/2 bg-white/10 backdrop-blur-xl rounded-lg shadow-lg p-3 min-w-48 border border-white/20 z-10 opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none">
                  <div className="text-xs font-semibold text-white mb-2">
                    {node.isCurrentNode ? 'Current Node' : `Node ${index + 1}`}
                  </div>
                  <div className="space-y-1 text-xs text-slate-300">
                    <div>ID: {node.id.slice(0, 12)}...</div>
                    <div>Address: {node.host}:{node.port}</div>
                    {isNodeActive(node) ? (
                      <>
                        <div>Slots: {node.slotStart.toLocaleString()} - {node.slotEnd.toLocaleString()}</div>
                        <div>Keys: {(node.keyCount || 0).toLocaleString()}</div>
                        <div>Storage: {formatBytes(node.byteSize || 0)}</div>
                        <div>Share: {(((node.slotEnd - node.slotStart + 1) / totalSlots) * 100).toFixed(1)}%</div>
                      </>
                    ) : (
                      <div className="text-amber-300 font-medium">Awaiting slot assignment</div>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}