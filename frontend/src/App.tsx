import { ConnectionStatus } from '@/components/ConnectionStatus'
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
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-slate-100 font-sans relative overflow-hidden">
      {/* Background Effects */}
      <div
        className="absolute inset-0 opacity-50"
        style={{
          backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%239C92AC' fill-opacity='0.05'%3E%3Ccircle cx='30' cy='30' r='2'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
        }}
      ></div>
      <div className="absolute top-0 -left-4 w-72 h-72 bg-purple-500 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-blob"></div>
      <div className="absolute top-0 -right-4 w-72 h-72 bg-blue-500 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-blob animation-delay-2000"></div>
      <div className="absolute -bottom-8 left-20 w-72 h-72 bg-pink-500 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-blob animation-delay-4000"></div>

      <div className="relative z-10 container mx-auto p-4 md:p-8 max-w-6xl">
        <header className="flex justify-between items-center mb-12">
          <div className="flex items-center space-x-4">
            <div className="p-3 bg-gradient-to-r from-purple-500 to-blue-500 rounded-2xl shadow-lg">
              <svg
                className="w-8 h-8 text-white"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  fillRule="evenodd"
                  d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm0 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1v-2z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            <div>
              <h1 className="text-4xl font-bold bg-gradient-to-r from-white via-purple-200 to-purple-400 bg-clip-text text-transparent">
                Reredis
              </h1>
              <p className="text-slate-400 text-sm">Live Key-Value Store</p>
            </div>
          </div>
          <ConnectionStatus isConnected={isConnected} />
        </header>

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
      </div>

      <ToastContainer toasts={toasts} onRemoveToast={removeToast} />
    </div>
  )
}

export default App
