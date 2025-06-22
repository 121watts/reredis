import { useState, useEffect, useRef } from 'react';
//
// Define message types to match the Go backend for type safety
interface UpdateMessage {
  action: 'set' | 'del' | 'get_resp';
  key: string;
  value?: string;
}

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

  useEffect(() => {
    const socket = new WebSocket('ws://localhost:8080/ws');
    ws.current = socket;

    socket.onopen = () => {
      console.log('Connected to WebSocket');
      setIsConnected(true);
    };

    socket.onmessage = (event) => {
      const message: UpdateMessage = JSON.parse(event.data);
      console.log('Received message:', message);

      setData((prevData) => {
        const newData = { ...prevData };
        if (message.action === 'set') {
          newData[message.key] = message.value ?? '';
        } else if (message.action === 'del') {
          delete newData[message.key];
        } else if (message.action === 'get_resp') {
          // A toast notification would be better, but alert is simple
          alert(`GET ${message.key}: ${message.value}`);
        }
        return newData;
      });
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

  const sendCommand = (action: 'set' | 'get' | 'del') => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      if (!form.key) {
        alert('Key cannot be empty.');
        return;
      }
      const command = { action, key: form.key, value: form.value };
      ws.current.send(JSON.stringify(command));
      // Clear value field after sending a command
      setForm(prev => ({ ...prev, value: '' }));
    } else {
      alert('WebSocket is not connected.');
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
              name="key"
              placeholder="Key"
              value={form.key}
              onChange={handleFormChange}
              className="md:col-span-2 bg-slate-900/70 rounded-md p-3 ring-1 ring-inset ring-slate-700 focus:ring-2 focus:ring-inset focus:ring-indigo-500 outline-none"
            />
            <input
              name="value"
              placeholder="Value (for SET)"
              value={form.value}
              onChange={handleFormChange}
              className="md:col-span-2 bg-slate-900/70 rounded-md p-3 ring-1 ring-inset ring-slate-700 focus:ring-2 focus:ring-inset focus:ring-indigo-500 outline-none"
            />
            <div className="grid grid-cols-3 gap-2 md:col-span-1">
              <button onClick={() => sendCommand('set')} className="bg-indigo-600 hover:bg-indigo-500 rounded-md p-3 font-semibold shadow-sm transition-colors">SET</button>
              <button onClick={() => sendCommand('get')} className="bg-sky-600 hover:bg-sky-500 rounded-md p-3 font-semibold shadow-sm transition-colors">GET</button>
              <button onClick={() => sendCommand('del')} className="bg-red-600 hover:bg-red-500 rounded-md p-3 font-semibold shadow-sm transition-colors">DEL</button>
            </div>
          </div>
        </div>

        <div className="bg-slate-800/50 rounded-lg shadow-lg ring-1 ring-white/10">
          <table className="w-full text-left">
            <thead className="border-b border-slate-700">
              <tr>
                <th className="p-4 text-sm font-semibold text-slate-300">Key</th>
                <th className="p-4 text-sm font-semibold text-slate-300">Value</th>
              </tr>
            </thead>
            <tbody>
              {Object.entries(data).length > 0 ? (
                Object.entries(data).map(([key, value]) => (
                  <tr key={key} className="border-b border-slate-800">
                    <td className="p-4 font-mono text-indigo-400">{key}</td>
                    <td className="p-4 font-mono text-slate-300">{value}</td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={2} className="p-8 text-center text-slate-500">No data in store. Use SET to add some!</td>
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
