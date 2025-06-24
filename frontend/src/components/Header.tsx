import { ConnectionStatus } from '@/components/ConnectionStatus'

interface HeaderProps {
  isConnected: boolean
}

export const Header = ({ isConnected }: HeaderProps) => {
  return (
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
  )
}
