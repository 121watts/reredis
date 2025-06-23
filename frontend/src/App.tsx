import { useState, useEffect, useRef } from 'react';
//
// Define message types to match the Go backend for type safety
interface SetMessage {
  action: 'set';
  key: string;
  value: string;
}

interface DelMessage {
  action: 'del';
  key: string;
}

interface SyncMessage {
  action: 'sync';
  data: Record<string, string>;
}

type ServerMessage = SetMessage | DelMessage | SyncMessage;

// Component to display connection status with a pulsing dot
const ConnectionStatus = ({ isConnected }: { isConnected: boolean }) => (
  <div className="flex items-center gap-2">
    <span className={`relative flex h-3 w-3`}>
      <span className={`animate-ping absolute inline-flex h-full w-full rounded-full ${isConnected ? 'bg-green-400' : 'bg-red-400'} opacity-75`}></span>
      <span className={`relative inline-flex rounded-full h-3 w-3 ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></span>
    </span>
    <span className="text-sm text-slate-400">
      {isConnected ? 'Connected' : 'Disconnected'}
    </span>
  </div>
);

function App() {
  const [data, setData] = useState<Record<string, string>>({});
  const [isConnected, setIsConnected] = useState(false);
  const [form, setForm] = useState({ key: '', value: '' });
  const ws = useRef<WebSocket | null>(null);
  const keyInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const socket = new WebSocket('ws://localhost:8080/ws');
    ws.current = socket;

    keyInputRef.current?.focus();

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
  }, []);

  const handleFormChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const sendCommand = (action: 'set' | 'del', keyOverride?: string) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      const key = keyOverride || form.key;
      if (!key) {
        alert('Key cannot be empty.');
        return;
      }
      const command = { action, key, value: action === 'set' ? form.value : '' };
      ws.current.send(JSON.stringify(command));

      // Clear form and refocus after sending a SET command from the main input
      if (action === 'set' && !keyOverride) {
        setForm({ key: '', value: '' });
        keyInputRef.current?.focus();
      }
    } else {
      alert('WebSocket is not connected.');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault(); // prevent form submission if it were in a form
      sendCommand('set');
    }
  };

  return (
    <div className="min-h-screen bg-slate-900 text-slate-100 font-sans">
      <div className="container mx-auto p-4 md:p-8 max-w-4xl">
        <header className="flex justify-between items-center mb-8">
          <h1 className="text-3xl font-bold text-white">Reredis Live View</h1>
          <ConnectionStatus isConnected={isConnected} />
        </header>

        <div className="bg-slate-800/50 p-6 rounded-lg shadow-lg mb-8 ring-1 ring-white/10">
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
            <input
              ref={keyInputRef}
              name="key"
              placeholder="Key"
              value={form.key}
              onChange={handleFormChange}
              onKeyDown={handleKeyDown}
              className="md:col-span-2 bg-slate-900/70 rounded-md p-3 ring-1 ring-inset ring-slate-700 focus:ring-2 focus:ring-inset focus:ring-indigo-500 outline-none"
            />
            <input
              name="value"
              placeholder="Value (for SET)"
              value={form.value}
              onChange={handleFormChange}
              onKeyDown={handleKeyDown}
              className="md:col-span-2 bg-slate-900/70 rounded-md p-3 ring-1 ring-inset ring-slate-700 focus:ring-2 focus:ring-inset focus:ring-indigo-500 outline-none"
            />
            <div className="md:col-span-1">
              <button onClick={() => sendCommand('set')} className="w-full bg-indigo-600 hover:bg-indigo-500 rounded-md p-3 font-semibold shadow-sm transition-colors">SET</button>
            </div>
          </div>
        </div>

        <div className="bg-slate-800/50 rounded-lg shadow-lg ring-1 ring-white/10">
          <table className="w-full text-left">
            <thead className="border-b border-slate-700">
              <tr>
                <th className="p-4 text-sm font-semibold text-slate-300">Key</th>
                <th className="p-4 text-sm font-semibold text-slate-300">Value</th>
                <th className="p-4 w-20"></th>
              </tr>
            </thead>
            <tbody>
              {Object.entries(data).length > 0 ? (
                Object.entries(data).map(([key, value]) => (
                  <tr key={key} className="border-b border-slate-800 group">
                    <td className="p-4 font-mono text-indigo-400">{key}</td>
                    <td className="p-4 font-mono text-slate-300">{value}</td>
                    <td className="p-4 text-right">
                      <button
                        onClick={() => sendCommand('del', key)}
                        className="invisible group-hover:visible bg-red-600 hover:bg-red-500 rounded-md px-3 py-1 text-xs font-semibold shadow-sm transition-colors"
                      >
                        DEL
                      </button>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={3} className="p-8 text-center text-slate-500">No data in store. Use SET to add some!</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

export default App
