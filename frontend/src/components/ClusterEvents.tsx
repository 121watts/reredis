import React, { useState, useEffect } from 'react'
import type { ClusterEventMessage } from '@/types/messages'

interface ClusterEvent extends Omit<ClusterEventMessage, 'action'> {
  timestamp: string
  id: string
}

interface ClusterEventsProps {
  className?: string
}

export const ClusterEvents: React.FC<ClusterEventsProps> = ({ className = '' }) => {
  const [events, setEvents] = useState<ClusterEvent[]>([])

  const addEvent = (eventData: Omit<ClusterEventMessage, 'action'>) => {
    const event: ClusterEvent = {
      ...eventData,
      timestamp: new Date().toLocaleTimeString(),
      id: `${Date.now()}-${Math.random()}`
    }
    
    setEvents(prev => [event, ...prev.slice(0, 49)]) // Keep last 50 events
  }

  // Expose addEvent to parent components via custom event
  useEffect(() => {
    const handleClusterEvent = (event: CustomEvent<Omit<ClusterEventMessage, 'action'>>) => {
      addEvent(event.detail)
    }

    window.addEventListener('cluster-event', handleClusterEvent as EventListener)
    return () => window.removeEventListener('cluster-event', handleClusterEvent as EventListener)
  }, [])

  const getEventIcon = (event: ClusterEvent['event']) => {
    switch (event) {
      case 'node_added':
        return '➕'
      case 'node_removed':
        return '➖'
      case 'slots_redistributed':
        return '🔄'
      default:
        return '📡'
    }
  }

  const getEventColor = (event: ClusterEvent['event']) => {
    switch (event) {
      case 'node_added':
        return 'text-green-200 bg-green-500/20 border-green-400/30'
      case 'node_removed':
        return 'text-red-200 bg-red-500/20 border-red-400/30'
      case 'slots_redistributed':
        return 'text-blue-200 bg-blue-500/20 border-blue-400/30'
      default:
        return 'text-slate-300 bg-white/10 border-white/20'
    }
  }

  return (
    <div className={`bg-white/5 backdrop-blur-xl rounded-3xl shadow-2xl border border-white/10 hover:border-white/20 transition-all duration-300 p-8 ${className}`}>
      <h3 className="text-xl font-bold text-white mb-6">Cluster Events</h3>
      
      {events.length === 0 ? (
        <div className="text-center py-8 text-slate-400">
          <div className="text-2xl mb-2">📡</div>
          <div>No cluster events yet</div>
          <div className="text-sm">Events will appear here as the cluster changes</div>
        </div>
      ) : (
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {events.map((event) => (
            <div
              key={event.id}
              className={`border rounded-lg p-3 transition-all duration-300 backdrop-blur-sm ${getEventColor(event.event)}`}
            >
              <div className="flex items-start space-x-3">
                <div className="text-lg">{getEventIcon(event.event)}</div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <div className="font-medium capitalize">
                      {event.event.replace('_', ' ')}
                    </div>
                    <div className="text-xs opacity-75">
                      {event.timestamp}
                    </div>
                  </div>
                  <div className="text-sm mt-1">
                    {event.message}
                  </div>
                  {event.nodeId && (
                    <div className="text-xs font-mono mt-1 opacity-75 truncate">
                      Node: {event.nodeId}
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
      
      {events.length > 0 && (
        <div className="mt-6 text-center">
          <button
            onClick={() => setEvents([])}
            className="text-sm text-slate-400 hover:text-slate-200 underline transition-colors"
          >
            Clear all events
          </button>
        </div>
      )}
    </div>
  )
}

// Helper function to dispatch cluster events from other components
export const dispatchClusterEvent = (eventData: Omit<ClusterEventMessage, 'action'>) => {
  window.dispatchEvent(new CustomEvent('cluster-event', { detail: eventData }))
}