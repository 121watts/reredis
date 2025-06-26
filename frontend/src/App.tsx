import { Layout } from '@/components/Layout'
import { Header } from '@/components/Header'
import { CommandForm } from '@/components/CommandForm'
import { DataViews } from '@/components/DataViews'
import { ClusterDashboard, ClusterHealthMetrics, ClusterNodeManager } from '@/components/cluster'
import { ClusterEvents } from '@/components/ClusterEvents'
import { ToastContainer } from '@/components/ToastContainer'
import { MockDataLoader } from '@/components/MockDataLoader'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useToast } from '@/hooks/useToast'
import { useURLState } from '@/hooks/useURLState'
import type { TabType } from '@/types/ui'

function App() {
  const [activeTab, setActiveTab] = useURLState<TabType>('tab', 'data', ['data', 'cluster'])
  const { data, isConnected, clusterInfo, sendCommand } = useWebSocket(
    'ws://localhost:9080/ws'
  )
  const { toasts, removeToast, success, error, warning } = useToast()

  const handleSendCommand = (
    action: 'set' | 'del',
    key: string,
    value?: string,
    silent = false
  ) => {
    if (!isConnected) {
      error('WebSocket is not connected')
      return
    }

    if (!key.trim()) {
      warning('Key cannot be empty')
      return
    }

    sendCommand(action, key, value)

    if (!silent) {
      if (action === 'set') {
        success(`Key "${key}" has been set`)
      } else {
        success(`Key "${key}" has been deleted`)
      }
    }
  }

  const handleDelete = (key: string) => {
    if (!isConnected) {
      error('WebSocket is not connected')
      return
    }

    sendCommand('del', key)
    success(`Key "${key}" has been deleted`)
  }

  const handleAddNode = (host: string, port: string) => {
    if (!isConnected) {
      error('WebSocket is not connected')
      return
    }

    // Send CLUSTER MEET command to add the node
    // Note: This would require extending the WebSocket protocol to support CLUSTER commands
    // For now, we'll just show a success message
    success(`Node ${host}:${port} addition requested`)
    warning('Note: CLUSTER MEET functionality requires server-side implementation')
  }

  return (
    <Layout>
      <Header isConnected={isConnected} />

      {/* Enhanced Tab Navigation */}
      <div className="bg-white/5 backdrop-blur-xl rounded-3xl shadow-2xl border border-white/10 p-2 mb-8">
        <nav className="flex space-x-2">
          <button
            onClick={() => setActiveTab('data')}
            className={`flex-1 py-3 px-6 rounded-2xl font-medium text-sm transition-all duration-200 ${
              activeTab === 'data'
                ? 'bg-blue-500/80 text-white shadow-lg shadow-blue-500/20 backdrop-blur-sm'
                : 'text-slate-300 hover:text-white hover:bg-white/10'
            }`}
          >
            <div className="flex items-center justify-center gap-2">
              <span className="text-lg">📊</span>
              <span>Data Management</span>
            </div>
          </button>
          <button
            onClick={() => setActiveTab('cluster')}
            className={`flex-1 py-3 px-6 rounded-2xl font-medium text-sm transition-all duration-200 ${
              activeTab === 'cluster'
                ? 'bg-blue-500/80 text-white shadow-lg shadow-blue-500/20 backdrop-blur-sm'
                : 'text-slate-300 hover:text-white hover:bg-white/10'
            }`}
          >
            <div className="flex items-center justify-center gap-2">
              <span className="text-lg">🔗</span>
              <span>Cluster Dashboard</span>
              {clusterInfo && (
                <span className="ml-2 px-2 py-1 bg-white/20 rounded-full text-xs">
                  {clusterInfo.nodes.filter(n => n.slotStart !== -1).length}/{clusterInfo.clusterSize}
                </span>
              )}
            </div>
          </button>
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'data' && (
        <div className="grid gap-8">
          <CommandForm
            onSendCommand={handleSendCommand}
            isConnected={isConnected}
          />
          <MockDataLoader
            onLoadData={handleSendCommand}
            isConnected={isConnected}
          />
          <DataViews
            data={data}
            onDelete={handleDelete}
            isConnected={isConnected}
          />
        </div>
      )}

      {activeTab === 'cluster' && (
        <div className="space-y-8">
          {clusterInfo ? (
            <>
              {/* Main cluster dashboard */}
              <ClusterDashboard
                nodes={clusterInfo.nodes}
                currentNodeId={clusterInfo.currentNodeId}
                totalSlots={clusterInfo.totalSlots}
                clusterSize={clusterInfo.clusterSize}
                isConnected={isConnected}
              />
              
              {/* Secondary dashboard components */}
              <div className="grid gap-8 lg:grid-cols-3">
                <div className="lg:col-span-1">
                  <ClusterHealthMetrics
                    nodes={clusterInfo.nodes}
                    totalSlots={clusterInfo.totalSlots}
                    isConnected={isConnected}
                  />
                </div>
                <div className="lg:col-span-1">
                  <ClusterNodeManager
                    isConnected={isConnected}
                    onAddNode={handleAddNode}
                  />
                </div>
                <div className="lg:col-span-1">
                  <ClusterEvents />
                </div>
              </div>
            </>
          ) : (
            <div className="grid gap-8 lg:grid-cols-2">
              <div className="bg-white/5 backdrop-blur-xl rounded-3xl shadow-2xl border border-white/10 hover:border-white/20 transition-all duration-300 p-8 text-center">
                <div className="text-6xl mb-4">🔄</div>
                <div className="text-xl font-semibold text-white mb-2">Loading Cluster Information</div>
                <div className="text-slate-400 mb-6">Connecting to cluster and retrieving node topology...</div>
                <button
                  onClick={() => sendCommand('cluster_info')}
                  className="px-6 py-3 bg-purple-500 text-white rounded-lg hover:bg-purple-600 transition-all shadow-md hover:shadow-lg font-medium"
                >
                  🔄 Refresh Cluster Info
                </button>
              </div>
              
              <div>
                <ClusterNodeManager
                  isConnected={isConnected}
                  onAddNode={handleAddNode}
                />
              </div>
            </div>
          )}
        </div>
      )}

      <ToastContainer toasts={toasts} onRemoveToast={removeToast} />
    </Layout>
  )
}

export default App
