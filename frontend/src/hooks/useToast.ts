import { useState, useCallback } from 'react'
import type { ToastProps } from '@/components/Toast'

type ToastType = 'success' | 'error' | 'warning' | 'info'

interface UseToastReturn {
  toasts: ToastProps[]
  addToast: (message: string, type: ToastType, duration?: number) => void
  removeToast: (id: string) => void
  success: (message: string, duration?: number) => void
  error: (message: string, duration?: number) => void
  warning: (message: string, duration?: number) => void
  info: (message: string, duration?: number) => void
}

export const useToast = (): UseToastReturn => {
  const [toasts, setToasts] = useState<ToastProps[]>([])

  const addToast = useCallback(
    (message: string, type: ToastType, duration = 4000) => {
      const id = Math.random().toString(36).substr(2, 9)
      const newToast: ToastProps = {
        id,
        message,
        type,
        duration,
        onClose: () => {}, // This will be set by the component
      }

      setToasts((prevToasts) => [...prevToasts, newToast])
    },
    []
  )

  const removeToast = useCallback((id: string) => {
    setToasts((prevToasts) => prevToasts.filter((toast) => toast.id !== id))
  }, [])

  // Convenience methods
  const success = useCallback(
    (message: string, duration?: number) => {
      addToast(message, 'success', duration)
    },
    [addToast]
  )

  const error = useCallback(
    (message: string, duration?: number) => {
      addToast(message, 'error', duration)
    },
    [addToast]
  )

  const warning = useCallback(
    (message: string, duration?: number) => {
      addToast(message, 'warning', duration)
    },
    [addToast]
  )

  const info = useCallback(
    (message: string, duration?: number) => {
      addToast(message, 'info', duration)
    },
    [addToast]
  )

  return {
    toasts,
    addToast,
    removeToast,
    success,
    error,
    warning,
    info,
  }
}
