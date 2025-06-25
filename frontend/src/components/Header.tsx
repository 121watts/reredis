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
            viewBox="0 0 24 24"
            fill="none"
          >
            {/* Database cylinder with 3D effect */}
            <defs>
              <linearGradient id="dbGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="rgba(255,255,255,0.9)" />
                <stop offset="50%" stopColor="rgba(255,255,255,0.7)" />
                <stop offset="100%" stopColor="rgba(255,255,255,0.4)" />
              </linearGradient>
              <linearGradient id="sideGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="rgba(255,255,255,0.8)" />
                <stop offset="100%" stopColor="rgba(255,255,255,0.3)" />
              </linearGradient>
            </defs>
            
            {/* Bottom ellipse (darker) */}
            <ellipse cx="12" cy="19" rx="8" ry="2.5" fill="rgba(255,255,255,0.3)" />
            
            {/* Side walls with gradient */}
            <path 
              d="M4 7v12c0 1.4 3.6 2.5 8 2.5s8-1.1 8-2.5V7" 
              fill="url(#sideGradient)" 
              stroke="rgba(255,255,255,0.6)" 
              strokeWidth="0.5"
            />
            
            {/* Middle partition */}
            <ellipse cx="12" cy="13" rx="8" ry="2.5" fill="rgba(255,255,255,0.5)" />
            <ellipse cx="12" cy="13" rx="8" ry="2.5" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="1" />
            
            {/* Top ellipse (brightest) */}
            <ellipse cx="12" cy="7" rx="8" ry="2.5" fill="url(#dbGradient)" />
            <ellipse cx="12" cy="7" rx="8" ry="2.5" fill="none" stroke="rgba(255,255,255,0.9)" strokeWidth="1.5" />
            
            {/* Highlight on top edge */}
            <ellipse cx="12" cy="6.5" rx="6" ry="1.5" fill="rgba(255,255,255,0.4)" />
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
