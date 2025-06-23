import { useState, useEffect, useRef, useCallback } from 'react';
import type { ServerMessage, CommandMessage } from '@/types/messages';

interface UseWebSocketReturn {
  data: Record<string, string>;
  isConnected: boolean;
  sendCommand: (action: 'set' | 'del', key: string, value?: string) => void;
}

export const useWebSocket = (url: string): UseWebSocketReturn => {
  const [data, setData] = useState<Record<string, string>>({});
  const [isConnected, setIsConnected] = useState(false);
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    const socket = new WebSocket(url);
    ws.current = socket;

    socket.onopen = () => {
      console.log('Connected to WebSocket');
      setIsConnected(true);
      // On connect, ask the server for the current state
      socket.send(JSON.stringify({ action: 'get_all' }));
    };

    socket.onmessage = (event) => {
      const message: ServerMessage = JSON.parse(event.data);
      console.log('Received message:', message);

      switch (message.action) {
        case 'sync':
          // Full state sync from server
          setData(message.data);
          break;
        case 'set':
          // A single key was set
          setData(prevData => ({ ...prevData, [message.key]: message.value }));
          break;
        case 'del':
          // A single key was deleted
          setData(prevData => {
            const newData = { ...prevData };
            delete newData[message.key];
            return newData;
          });
          break;
      }
    };

    socket.onclose = () => {
      console.log('Disconnected from WebSocket');
      setIsConnected(false);
    };

    socket.onerror = (error) => {
      console.error('WebSocket Error:', error);
      setIsConnected(false);
    };

    // Cleanup on component unmount
    return () => {
      socket.close();
    };
  }, [url]);

  const sendCommand = useCallback((action: 'set' | 'del', key: string, value?: string) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      if (!key) {
        console.error('Key cannot be empty.');
        return;
      }
      
      const command: CommandMessage = { 
        action, 
        key, 
        ...(action === 'set' && value !== undefined && { value })
      };
      
      ws.current.send(JSON.stringify(command));
    } else {
      console.error('WebSocket is not connected.');
    }
  }, []);

  return { data, isConnected, sendCommand };
};