import { useEffect } from 'react'

export interface ToastProps {
  id: string
  message: string
  type: 'success' | 'error' | 'warning' | 'info'
  onClose: (id: string) => void
  duration?: number
}

export const Toast = ({
  id,
  message,
  type,
  onClose,
  duration = 4000,
}: ToastProps) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose(id)
    }, duration)

    return () => clearTimeout(timer)
  }, [id, duration, onClose])

  const getIcon = () => {
    switch (type) {
      case 'success':
        return (
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
              d="M5 13l4 4L19 7"
            />
          </svg>
        )
      case 'error':
        return (
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
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        )
      case 'warning':
        return (
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
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
            />
          </svg>
        )
      case 'info':
        return (
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
              d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        )
    }
  }

  const getColors = () => {
    switch (type) {
      case 'success':
        return 'bg-emerald-500/80 border-emerald-500/60 text-white'
      case 'error':
        return 'bg-red-500/80 border-red-500/60 text-white'
      case 'warning':
        return 'bg-yellow-500/80 border-yellow-500/60 text-white'
      case 'info':
        return 'bg-blue-500/80 border-blue-500/60 text-white'
    }
  }

  return (
    <div
      className={`flex items-center gap-3 p-4 rounded-2xl backdrop-blur-xl border shadow-2xl transform transition-all duration-300 animate-in slide-in-from-right ${getColors()}`}
      style={{ filter: 'drop-shadow(0 10px 20px rgba(0, 0, 0, 0.5))' }}
    >
      <div className="flex-shrink-0">{getIcon()}</div>
      <div className="flex-1 text-sm font-medium">{message}</div>
      <button
        onClick={() => onClose(id)}
        className="flex-shrink-0 p-1 rounded-lg hover:bg-white/10 transition-colors"
      >
        <svg
          className="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
    </div>
  )
}
