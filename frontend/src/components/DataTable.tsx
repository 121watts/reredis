interface DataTableProps {
  data: Record<string, string>;
  onDelete: (key: string) => void;
  isConnected: boolean;
}

export const DataTable = ({ data, onDelete, isConnected }: DataTableProps) => {
  const entries = Object.entries(data);
  
  return (
    <div className="bg-white/5 backdrop-blur-xl rounded-3xl shadow-2xl border border-white/10 overflow-hidden">
      <div className="p-6 border-b border-white/10">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-semibold text-white">Stored Data</h2>
            <p className="text-slate-400 text-sm mt-1">
              {entries.length} {entries.length === 1 ? 'entry' : 'entries'} in cache
            </p>
          </div>
          <div className="flex items-center gap-2 px-3 py-1 bg-purple-500/20 rounded-full">
            <div className="w-2 h-2 bg-purple-400 rounded-full animate-pulse"></div>
            <span className="text-xs text-purple-300 font-medium">LIVE</span>
          </div>
        </div>
      </div>

      <div className="overflow-x-auto">
        {entries.length > 0 ? (
          <table className="w-full">
            <thead>
              <tr className="border-b border-white/5 bg-white/5">
                <th className="p-6 text-left text-sm font-semibold text-slate-300 uppercase tracking-wider">Key</th>
                <th className="p-6 text-left text-sm font-semibold text-slate-300 uppercase tracking-wider">Value</th>
                <th className="p-6 text-right text-sm font-semibold text-slate-300 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {entries.map(([key, value], index) => (
                <tr 
                  key={key} 
                  className="group hover:bg-white/5 transition-all duration-200 animate-in slide-in-from-left"
                  style={{ animationDelay: `${index * 50}ms` }}
                >
                  <td className="p-6">
                    <div className="flex items-center gap-3">
                      <div className="w-2 h-2 bg-emerald-400 rounded-full"></div>
                      <span className="font-mono text-emerald-400 font-medium">{key}</span>
                    </div>
                  </td>
                  <td className="p-6">
                    <span className="font-mono text-slate-300 bg-slate-800/50 px-3 py-1 rounded-lg">
                      {value}
                    </span>
                  </td>
                  <td className="p-6 text-right">
                    <button
                      onClick={() => onDelete(key)}
                      disabled={!isConnected}
                      className="opacity-0 group-hover:opacity-100 bg-red-500/20 hover:bg-red-500/30 disabled:bg-slate-600/20 disabled:cursor-not-allowed rounded-xl px-4 py-2 text-sm font-medium text-red-300 hover:text-red-200 transition-all duration-200 transform hover:scale-105 active:scale-95 disabled:transform-none border border-red-500/30"
                    >
                      <span className="flex items-center gap-2">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                        DELETE
                      </span>
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div className="p-12 text-center">
            <div className="mb-4">
              <svg className="w-16 h-16 text-slate-600 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-slate-400 mb-2">
              {isConnected ? 'No data stored yet' : 'Connection lost'}
            </h3>
            <p className="text-slate-500 text-sm">
              {isConnected 
                ? 'Add your first key-value pair using the form above' 
                : 'Reconnecting to server...'
              }
            </p>
          </div>
        )}
      </div>
    </div>
  );
};