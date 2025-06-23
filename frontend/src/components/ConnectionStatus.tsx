interface ConnectionStatusProps {
  isConnected: boolean;
}

export const ConnectionStatus = ({ isConnected }: ConnectionStatusProps) => (
  <div className="flex items-center gap-3 px-4 py-2 rounded-2xl bg-white/10 backdrop-blur-md border border-white/20 shadow-lg">
    <div className="relative">
      <span className={`relative flex h-3 w-3`}>
        <span className={`animate-ping absolute inline-flex h-full w-full rounded-full ${isConnected ? 'bg-emerald-400' : 'bg-red-400'} opacity-75`}></span>
        <span className={`relative inline-flex rounded-full h-3 w-3 ${isConnected ? 'bg-emerald-500' : 'bg-red-500'} shadow-lg`}></span>
      </span>
    </div>
    <div className="flex flex-col">
      <span className={`text-sm font-medium ${isConnected ? 'text-emerald-300' : 'text-red-300'}`}>
        {isConnected ? 'Connected' : 'Disconnected'}
      </span>
      <span className="text-xs text-slate-400">
        {isConnected ? 'Real-time sync active' : 'Attempting to reconnect...'}
      </span>
    </div>
  </div>
);