import { usePagination } from '@/hooks/usePagination'

export const PaginatedDataTable = () => {
  const { keys, loading, hasMore, loadMore, reset } = usePagination(
    'http://localhost:8080/api/v1/keys',
    20
  )

  return (
    <div className="bg-white/5 backdrop-blur-xl p-6 rounded-3xl shadow-2xl border border-white/10">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-xl font-semibold text-white mb-2">All Keys</h2>
          <p className="text-slate-400 text-sm">
            Browse all keys with pagination ({keys.length} loaded)
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={reset}
            className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors"
          >
            Reset
          </button>
          <button
            onClick={loadMore}
            disabled={loading || !hasMore}
            className="px-4 py-2 bg-purple-600 hover:bg-purple-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
          >
            {loading ? 'Loading...' : hasMore ? 'Load More' : 'All Loaded'}
          </button>
        </div>
      </div>

      {keys.length === 0 ? (
        <div className="text-center py-8 text-slate-400">
          <p>No keys loaded. Click "Load More" to start browsing.</p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-slate-700">
                <th className="text-left py-3 px-4 text-slate-300 font-medium">
                  Key
                </th>
              </tr>
            </thead>
            <tbody>
              {keys.map((key, i) => (
                <tr
                  key={i}
                  className="border-b border-slate-800 hover:bg-white/5"
                >
                  <td className="py-3 px-4 text-slate-200 font-mono text-sm">
                    {key}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
