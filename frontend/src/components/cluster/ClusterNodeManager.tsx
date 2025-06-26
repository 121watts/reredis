import React, { useState } from 'react'
import { Card, CardHeader, CardTitle, CardContent, Button, Input } from '@/components/ui'

interface ClusterNodeManagerProps {
  isConnected: boolean
  onAddNode: (host: string, port: string) => void
}

const ConnectionWarning: React.FC = () => (
  <div className="bg-red-500/20 border border-red-400/30 rounded-lg p-4 backdrop-blur-sm">
    <div className="flex items-center gap-2 text-red-200">
      <span className="text-lg">⚠️</span>
      <span className="font-medium">Connection Required</span>
    </div>
    <p className="text-red-300 text-sm mt-1">
      Please ensure the WebSocket connection is active before adding nodes.
    </p>
  </div>
)

const ClusterTips: React.FC = () => (
  <div className="bg-blue-500/20 border border-blue-400/30 rounded-lg p-4 backdrop-blur-sm">
    <div className="flex items-center gap-2 text-blue-200 mb-2">
      <span className="text-lg">💡</span>
      <span className="font-medium">Cluster Formation Tips</span>
    </div>
    <ul className="text-blue-300 text-sm space-y-1">
      <li>• A minimum of 3 nodes is required for automatic slot distribution</li>
      <li>• Each node should have a unique port number</li>
      <li>• Nodes will automatically receive slot assignments when the cluster reaches 3 members</li>
      <li>• Use <code className="bg-blue-500/30 px-1 rounded text-blue-200">CLUSTER MEET</code> commands to connect nodes</li>
    </ul>
  </div>
)

/**
 * Interface for adding new nodes to the cluster
 * Provides form validation and helpful tips for cluster management
 */
export const ClusterNodeManager: React.FC<ClusterNodeManagerProps> = ({
  isConnected,
  onAddNode
}) => {
  const [host, setHost] = useState('localhost')
  const [port, setPort] = useState('')
  const [isAdding, setIsAdding] = useState(false)

  const handleAddNode = async () => {
    if (!host.trim() || !port.trim()) {
      return
    }

    setIsAdding(true)
    try {
      await onAddNode(host.trim(), port.trim())
      // Reset port after successful addition, keep host
      setPort('')
    } catch (error) {
      console.error('Failed to add node:', error)
    } finally {
      setIsAdding(false)
    }
  }

  const canAddNode = isConnected && host.trim() && port.trim() && !isAdding

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-3">
          <div className="w-3 h-3 bg-purple-500 rounded-full shadow-lg shadow-purple-500/50" />
          <CardTitle>Node Management</CardTitle>
        </div>
      </CardHeader>

      <CardContent className="space-y-6">
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-4">
            Add New Node to Cluster
          </label>
          
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <Input
              placeholder="Host (e.g., localhost)"
              value={host}
              onChange={(e) => setHost(e.target.value)}
              disabled={!isConnected || isAdding}
            />
            
            <Input
              placeholder="Port (e.g., 6380)"
              value={port}
              onChange={(e) => setPort(e.target.value)}
              disabled={!isConnected || isAdding}
            />
            
            <Button
              onClick={handleAddNode}
              disabled={!canAddNode}
              loading={isAdding}
              variant="primary"
              className="w-full"
            >
              ➕ Add Node
            </Button>
          </div>
        </div>

        {!isConnected && <ConnectionWarning />}
        
        <ClusterTips />
      </CardContent>
    </Card>
  )
}