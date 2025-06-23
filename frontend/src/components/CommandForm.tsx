import { useState, useRef, useEffect } from 'react'

interface CommandFormProps {
  onSendCommand: (action: 'set' | 'del', key: string, value?: string) => void
  isConnected: boolean
}

export const CommandForm = ({
  onSendCommand,
  isConnected,
}: CommandFormProps) => {
  const [form, setForm] = useState({ key: '', value: '' })
  const keyInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    keyInputRef.current?.focus()
  }, [])

  const handleFormChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm({ ...form, [e.target.name]: e.target.value })
  }

  const handleSendCommand = (action: 'set' | 'del') => {
    onSendCommand(action, form.key, action === 'set' ? form.value : undefined)

    // Clear form and refocus after sending a SET command
    if (action === 'set') {
      setForm({ key: '', value: '' })
      keyInputRef.current?.focus()
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleSendCommand('set')
    }
  }

  return (
    <div className="bg-white/5 backdrop-blur-xl p-8 rounded-3xl shadow-2xl border border-white/10 hover:border-white/20 transition-all duration-300">
      <div className="mb-6">
        <h2 className="text-xl font-semibold text-white mb-2">Add New Entry</h2>
        <p className="text-slate-400 text-sm">
          Enter a key-value pair to store in your Redis cache
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4">
        <div className="lg:col-span-4 group">
          <label className="block text-sm font-medium text-slate-300 mb-2">
            Key
          </label>
          <input
            ref={keyInputRef}
            name="key"
            placeholder="Enter key name..."
            value={form.key}
            onChange={handleFormChange}
            onKeyDown={handleKeyDown}
            disabled={!isConnected}
            className="w-full bg-white/5 backdrop-blur-sm rounded-2xl p-4 ring-1 ring-white/10 focus:ring-2 focus:ring-purple-500/50 outline-none disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 placeholder-slate-500 group-hover:ring-white/20"
          />
        </div>

        <div className="lg:col-span-5 group">
          <label className="block text-sm font-medium text-slate-300 mb-2">
            Value
          </label>
          <input
            name="value"
            placeholder="Enter value to store..."
            value={form.value}
            onChange={handleFormChange}
            onKeyDown={handleKeyDown}
            disabled={!isConnected}
            className="w-full bg-white/5 backdrop-blur-sm rounded-2xl p-4 ring-1 ring-white/10 focus:ring-2 focus:ring-purple-500/50 outline-none disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 placeholder-slate-500 group-hover:ring-white/20"
          />
        </div>

        <div className="lg:col-span-3 flex items-end">
          <button
            onClick={() => handleSendCommand('set')}
            disabled={!isConnected}
            className="w-full bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-500 hover:to-blue-500 disabled:from-slate-600 disabled:to-slate-700 disabled:cursor-not-allowed rounded-2xl p-4 font-semibold shadow-lg transition-all duration-200 transform hover:scale-105 hover:shadow-purple-500/25 active:scale-95"
          >
            <span className="flex items-center justify-center gap-2">
              <svg
                className="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 4v16m8-8H4"
                />
              </svg>
              SET
            </span>
          </button>
        </div>
      </div>
    </div>
  )
}
