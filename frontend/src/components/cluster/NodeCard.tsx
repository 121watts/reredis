import React from 'react'
import { Card, Badge, Progress } from '@/components/ui'
import { useClusterNodes } from '@/hooks/useClusterNodes'
import { formatBytes } from '@/utils/formatBytes'
import type { ClusterNode } from '@/types/messages'

interface NodeCardProps {
  node: ClusterNode
  totalSlots: number
  isSelected?: boolean
  onSelect?: () => void
}

/**
 * Individual node card displaying node status, slots, and keys
 * Interactive card that can be selected for detailed view
 */
export const NodeCard: React.FC<NodeCardProps> = ({
  node,
  totalSlots,
  isSelected = false,
  onSelect
}) => {
  const { 
    getSlotPercentage, 
    getNodeStatus, 
    getStatusVariant, 
    formatNodeId,
    isNodeActive 
  } = useClusterNodes(totalSlots)

  const slotPercentage = getSlotPercentage(node)
  const status = getNodeStatus(node)
  const variant = getStatusVariant(node)

  const getCardVariant = () => {
    if (isSelected) return 'glass'
    if (node.isCurrentNode) return 'gradient'
    return 'default'
  }

  const getBorderClass = () => {
    if (isSelected) return 'border-blue-400 shadow-lg shadow-blue-500/20'
    if (node.isCurrentNode) return 'border-emerald-400 shadow-lg shadow-emerald-500/20'
    if (isNodeActive(node)) return 'border-green-400 shadow-lg shadow-green-500/10'
    return 'border-white/20'
  }

  return (
    <Card
      variant={getCardVariant()}
      className={`cursor-pointer transition-all duration-200 ${getBorderClass()}`}
      onClick={onSelect}
    >
      <div className="p-6">
        {/* Node Header */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className={`w-4 h-4 rounded-full ${
              node.isCurrentNode
                ? 'bg-emerald-400 shadow-lg shadow-emerald-400/50'
                : isNodeActive(node)
                ? 'bg-green-400 shadow-lg shadow-green-400/50'
                : 'bg-slate-400'
            }`} />
            <h3 className="font-bold text-lg text-white">
              Node {node.host}:{node.port}
            </h3>
          </div>
          <Badge variant={variant} size="sm">
            {status === 'current' ? 'Current' : status === 'active' ? 'Active' : 'Waiting'}
          </Badge>
        </div>

        {/* Node Details */}
        <div className="space-y-3">
          <div>
            <div className="text-xs text-slate-400 uppercase tracking-wide font-semibold mb-1">
              Node ID
            </div>
            <div className="font-mono text-sm bg-white/10 px-2 py-1 rounded border border-white/20 text-slate-200">
              {formatNodeId(node.id, 16)}
            </div>
          </div>
          
          <div>
            <div className="text-xs text-slate-400 uppercase tracking-wide font-semibold mb-1">
              Address
            </div>
            <div className="font-mono text-sm text-slate-200">
              {node.host}:{node.port}
            </div>
          </div>

          {isNodeActive(node) ? (
            <>
              <div>
                <div className="text-xs text-slate-400 uppercase tracking-wide font-semibold mb-1">
                  Slot Range
                </div>
                <div className="font-mono text-sm text-slate-200">
                  {node.slotStart.toLocaleString()} - {node.slotEnd.toLocaleString()}
                </div>
              </div>
              
              <div>
                <div className="text-xs text-slate-400 uppercase tracking-wide font-semibold mb-1">
                  Cluster Share
                </div>
                <Progress 
                  value={slotPercentage}
                  variant={node.isCurrentNode ? 'success' : 'default'}
                  showLabel
                  size="sm"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-xs text-slate-400 uppercase tracking-wide font-semibold mb-1">
                    Keys Stored
                  </div>
                  <div className="text-lg font-bold text-slate-200">
                    {(node.keyCount || 0).toLocaleString()}
                  </div>
                </div>
                <div>
                  <div className="text-xs text-slate-400 uppercase tracking-wide font-semibold mb-1">
                    Storage Used
                  </div>
                  <div className="text-lg font-bold text-slate-200">
                    {formatBytes(node.byteSize || 0)}
                  </div>
                </div>
              </div>
            </>
          ) : (
            <div className="text-center py-4 text-amber-400">
              <div className="text-2xl mb-2">⏳</div>
              <div className="text-sm font-medium">Awaiting slot assignment</div>
            </div>
          )}
        </div>

        {/* Expanded Details */}
        {isSelected && isNodeActive(node) && (
          <div className="mt-4 pt-4 border-t border-white/20 space-y-2">
            <div className="text-xs text-slate-400 uppercase tracking-wide font-semibold">
              Additional Details
            </div>
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div>
                <span className="text-slate-400">Slot Count:</span>
                <span className="ml-1 font-semibold text-slate-200">
                  {(node.slotEnd - node.slotStart + 1).toLocaleString()}
                </span>
              </div>
              <div>
                <span className="text-slate-400">Avg Keys/Slot:</span>
                <span className="ml-1 font-semibold text-slate-200">
                  {Math.round((node.keyCount || 0) / (node.slotEnd - node.slotStart + 1)) || 0}
                </span>
              </div>
              <div>
                <span className="text-slate-400">Avg Bytes/Key:</span>
                <span className="ml-1 font-semibold text-slate-200">
                  {(node.keyCount || 0) > 0 ? formatBytes((node.byteSize || 0) / (node.keyCount || 1)) : '0 B'}
                </span>
              </div>
              <div>
                <span className="text-slate-400">Total Storage:</span>
                <span className="ml-1 font-semibold text-slate-200">
                  {formatBytes(node.byteSize || 0)}
                </span>
              </div>
            </div>
          </div>
        )}
      </div>
    </Card>
  )
}