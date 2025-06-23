import { Toast } from './Toast'
import type { ToastProps } from './Toast'

interface ToastContainerProps {
  toasts: ToastProps[]
  onRemoveToast: (id: string) => void
}

export const ToastContainer = ({
  toasts,
  onRemoveToast,
}: ToastContainerProps) => {
  return (
    <div className="fixed top-4 right-4 z-50 space-y-3 w-80">
      {toasts.map((toast) => (
        <Toast key={toast.id} {...toast} onClose={onRemoveToast} />
      ))}
    </div>
  )
}
