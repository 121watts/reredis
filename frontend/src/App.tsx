import { Layout } from '@/components/Layout'
import { Header } from '@/components/Header'
import { CommandForm } from '@/components/CommandForm'
import { DataTable } from '@/components/DataTable'
import { ToastContainer } from '@/components/ToastContainer'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useToast } from '@/hooks/useToast'

function App() {
  const { data, isConnected, sendCommand } = useWebSocket(
    'ws://localhost:8080/ws'
  )
  const { toasts, removeToast, success, error, warning } = useToast()

  const handleSendCommand = (
    action: 'set' | 'del',
    key: string,
    value?: string
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

    if (action === 'set') {
      success(`Key "${key}" has been set`)
    } else {
      success(`Key "${key}" has been deleted`)
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

  return (
    <Layout>
      <Header isConnected={isConnected} />

      <div className="grid gap-8">
        <CommandForm
          onSendCommand={handleSendCommand}
          isConnected={isConnected}
        />
        <DataTable
          data={data}
          onDelete={handleDelete}
          isConnected={isConnected}
        />
      </div>

      <ToastContainer toasts={toasts} onRemoveToast={removeToast} />
    </Layout>
  )
}

export default App
