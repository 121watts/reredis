import React from 'react'
import { Card, CardHeader, CardTitle, CardContent, Badge, Progress } from '@/components/ui'
import { useClusterMetrics } from '@/hooks/useClusterMetrics'
import type { ClusterNode } from '@/types/messages'

interface ClusterHealthMetricsProps {
  nodes: ClusterNode[]
  totalSlots: number
  isConnected: boolean
}

interface HealthIndicatorProps {
  label: string
  value: number
  unit?: string
  description?: string
}

const HealthIndicator: React.FC<HealthIndicatorProps> = ({ 
  label, 
  value, 
  unit = '%', 
  description 
}) => (
  <div>
    <div className="flex justify-between text-sm mb-2">
      <span className="font-medium text-slate-300">{label}</span>
      <span className="text-slate-300">
        {Math.round(value)}{unit} {description && `(${description})`}
      </span>
    </div>
    <Progress value={value} showLabel={false} />
  </div>
)

const HealthRecommendations: React.FC<{
  metrics: ReturnType<typeof useClusterMetrics>
  isConnected: boolean
}> = ({ metrics, isConnected }) => {
  const recommendations = []
  
  if (!isConnected) {
    recommendations.push('Restore WebSocket connection to the cluster')
  }
  if (metrics.activeNodes < 3) {
    recommendations.push('Add more nodes to reach minimum cluster size (3 nodes)')
  }
  if (metrics.slotCoverage < 100) {
    recommendations.push('Ensure all hash slots are properly assigned')
  }
  if (metrics.loadBalance < 70) {
    recommendations.push('Consider rebalancing data distribution across nodes')
  }

  if (recommendations.length === 0) {
    return null
  }

  return (
    <div className="mt-6 p-4 bg-yellow-500/20 border border-yellow-400/30 rounded-lg backdrop-blur-sm">
      <div className="flex items-center gap-2 text-yellow-200 mb-2">
        <span>⚠️</span>
        <span className="font-semibold">Health Recommendations</span>
      </div>
      <ul className="text-yellow-300 text-sm space-y-1">
        {recommendations.map((rec, index) => (
          <li key={index}>• {rec}</li>
        ))}
      </ul>
    </div>
  )
}

/**
 * Displays comprehensive cluster health metrics and recommendations
 * Shows connection status, cluster formation, slot distribution, and load balance
 */
export const ClusterHealthMetrics: React.FC<ClusterHealthMetricsProps> = ({
  nodes,
  totalSlots,
  isConnected
}) => {
  const metrics = useClusterMetrics(nodes, totalSlots, isConnected)

  const getHealthVariant = (score: number) => {
    if (score >= 80) return 'success'
    if (score >= 60) return 'warning'
    return 'danger'
  }

  const overviewStats = [
    { icon: '🟢', label: 'Active Nodes', value: `${metrics.activeNodes}/${metrics.totalNodes}` },
    { icon: '🔑', label: 'Total Keys', value: metrics.totalKeys.toLocaleString() },
    { icon: '📊', label: 'Slot Coverage', value: `${Math.round(metrics.slotCoverage)}%` },
    { icon: '⚖️', label: 'Load Balance', value: `${Math.round(metrics.loadBalance)}%` }
  ]

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Cluster Health</CardTitle>
          <Badge 
            variant={getHealthVariant(metrics.overallHealth)}
            size="md"
          >
            {Math.round(metrics.overallHealth)}% Healthy
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-6">
        {/* Health Overview */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {overviewStats.map((stat, index) => (
            <div key={index} className="text-center">
              <div className="text-2xl mb-2">{stat.icon}</div>
              <div className="text-2xl font-bold text-white">{stat.value}</div>
              <div className="text-sm text-slate-400">{stat.label}</div>
            </div>
          ))}
        </div>

        {/* Detailed Health Metrics */}
        <div className="space-y-4">
          <HealthIndicator
            label="Connection Status"
            value={isConnected ? 100 : 0}
            description={isConnected ? 'Connected' : 'Disconnected'}
          />
          
          <HealthIndicator
            label="Cluster Formation"
            value={metrics.activeNodes >= 3 ? 100 : (metrics.activeNodes / 3) * 100}
            description={metrics.activeNodes >= 3 ? 'Complete' : `${metrics.activeNodes}/3 nodes`}
          />
          
          <HealthIndicator
            label="Slot Distribution"
            value={metrics.slotCoverage}
            description="coverage"
          />
          
          <HealthIndicator
            label="Load Balance"
            value={metrics.loadBalance}
            description="balanced"
          />
        </div>

        {/* Key Distribution Details */}
        {metrics.activeNodes > 0 && (
          <div className="mt-6 pt-6 border-t border-white/20">
            <h4 className="font-semibold text-white mb-3">Key Distribution</h4>
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <div className="text-lg font-bold text-blue-400">
                  {metrics.avgKeysPerNode.toLocaleString()}
                </div>
                <div className="text-xs text-slate-400">Avg per Node</div>
              </div>
              <div>
                <div className="text-lg font-bold text-emerald-400">
                  {metrics.maxKeys.toLocaleString()}
                </div>
                <div className="text-xs text-slate-400">Max on Node</div>
              </div>
              <div>
                <div className="text-lg font-bold text-amber-400">
                  {metrics.minKeys.toLocaleString()}
                </div>
                <div className="text-xs text-slate-400">Min on Node</div>
              </div>
            </div>
          </div>
        )}

        {/* Health Recommendations */}
        <HealthRecommendations 
          metrics={metrics} 
          isConnected={isConnected} 
        />
      </CardContent>
    </Card>
  )
}