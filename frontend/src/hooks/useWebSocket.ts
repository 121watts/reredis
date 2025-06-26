import { useState, useEffect, useRef, useCallback } from 'react'
import type { ServerMessage, CommandMessage, ClusterNode } from '@/types/messages'
import { dispatchClusterEvent } from '@/components/ClusterEvents'

interface ClusterInfo {
  nodes: ClusterNode[]
  currentNodeId: string
  totalSlots: number
  clusterSize: number
  totalKeys?: number
}

interface UseWebSocketReturn {
  data: Record<string, string>
  isConnected: boolean
  clusterInfo: ClusterInfo | null
  sendCommand: (action: 'set' | 'del' | 'cluster_info', key?: string, value?: string) => void
}

export const useWebSocket = (url: string): UseWebSocketReturn => {
  const [data, setData] = useState<Record<string, string>>({})
  const [isConnected, setIsConnected] = useState(false)
  const [clusterInfo, setClusterInfo] = useState<ClusterInfo | null>(null)
  const ws = useRef<WebSocket | null>(null)

  useEffect(() => {
    const socket = new WebSocket(url)
    ws.current = socket

    socket.onopen = () => {
      console.log('Connected to WebSocket')
      setIsConnected(true)
      // On connect, ask the server for the current state and cluster info
      socket.send(JSON.stringify({ action: 'get_all' }))
      socket.send(JSON.stringify({ action: 'cluster_info' }))
    }

    socket.onmessage = (event) => {
      const message: ServerMessage = JSON.parse(event.data)
      console.log('Received message:', message)

      switch (message.action) {
        case 'sync':
          // Full state sync from server
          setData(message.data)
          break
        case 'set':
          // A single key was set
          setData((prevData) => ({ ...prevData, [message.key]: message.value }))
          break
        case 'del':
          // A single key was deleted
          setData((prevData) => {
            const newData = { ...prevData }
            delete newData[message.key]
            return newData
          })
          break
        case 'cluster_info':
          // Cluster information update
          setClusterInfo({
            nodes: message.nodes.map(node => ({
              ...node,
              isCurrentNode: node.id === message.currentNodeId
            })),
            currentNodeId: message.currentNodeId,
            totalSlots: message.totalSlots,
            clusterSize: message.clusterSize
          })
          break
        case 'cluster_stats':
          // Real-time cluster statistics update
          setClusterInfo({
            nodes: message.nodes.map(node => ({
              ...node,
              isCurrentNode: node.id === message.currentNodeId
            })),
            currentNodeId: message.currentNodeId,
            totalSlots: message.totalSlots,
            clusterSize: message.clusterSize,
            totalKeys: message.totalKeys
          })
          break
        case 'cluster_event':
          // Cluster event notification
          dispatchClusterEvent({
            event: message.event,
            nodeId: message.nodeId,
            message: message.message
          })
          // Re-request cluster info to get updated state
          if (ws.current && ws.current.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({ action: 'cluster_info' }))
          }
          break
      }
    }

    socket.onclose = () => {
      console.log('Disconnected from WebSocket')
      setIsConnected(false)
    }

    socket.onerror = (error) => {
      console.error('WebSocket Error:', error)
      setIsConnected(false)
    }

    // Cleanup on component unmount
    return () => {
      socket.close()
    }
  }, [url])

  const sendCommand = useCallback(
    (action: 'set' | 'del' | 'cluster_info', key?: string, value?: string) => {
      if (ws.current && ws.current.readyState === WebSocket.OPEN) {
        if ((action === 'set' || action === 'del') && !key) {
          console.error('Key cannot be empty for set/del operations.')
          return
        }

        const command: CommandMessage = {
          action,
          ...(key && { key }),
          ...(action === 'set' && value !== undefined && { value }),
        }

        ws.current.send(JSON.stringify(command))
      } else {
        console.error('WebSocket is not connected.')
      }
    },
    []
  )

  return { data, isConnected, clusterInfo, sendCommand }
}
