import { useState } from 'react'
import { DataTable } from '@/components/DataTable'
import { PaginatedDataTable } from '@/components/PaginatedDataTable'

interface DataViewsProps {
  data: { [key: string]: string }
  onDelete: (key: string) => void
  isConnected: boolean
}

type ViewMode = 'live' | 'browse'

export const DataViews = ({ data, onDelete, isConnected }: DataViewsProps) => {
  const [activeView, setActiveView] = useState<ViewMode>('live')

  return (
    <div className="bg-white/5 backdrop-blur-xl rounded-3xl shadow-2xl border border-white/10 overflow-hidden">
      {/* Tab Headers */}
      <div className="flex border-b border-white/10">
        <button
          onClick={() => setActiveView('live')}
          className={`flex-1 px-6 py-4 text-center font-medium transition-all ${
            activeView === 'live'
              ? 'bg-purple-600/20 text-purple-300 border-b-2 border-purple-400'
              : 'text-slate-400 hover:text-white hover:bg-white/5'
          }`}
        >
          <div className="flex items-center justify-center gap-2">
            <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
            Live View
          </div>
          <p className="text-xs mt-1 opacity-75">
            Real-time updates via WebSocket
          </p>
        </button>
        
        <button
          onClick={() => setActiveView('browse')}
          className={`flex-1 px-6 py-4 text-center font-medium transition-all ${
            activeView === 'browse'
              ? 'bg-purple-600/20 text-purple-300 border-b-2 border-purple-400'
              : 'text-slate-400 hover:text-white hover:bg-white/5'
          }`}
        >
          <div className="flex items-center justify-center gap-2">
            📄 Browse All
          </div>
          <p className="text-xs mt-1 opacity-75">
            Paginated view of all keys
          </p>
        </button>
      </div>

      {/* Tab Content */}
      <div className="p-6">
        {activeView === 'live' ? (
          <DataTable
            data={data}
            onDelete={onDelete}
            isConnected={isConnected}
          />
        ) : (
          <PaginatedDataTable />
        )}
      </div>
    </div>
  )
}