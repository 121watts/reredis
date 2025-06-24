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
    <div className="fixed top-4 right-4 z-[9999] space-y-3 w-80 max-w-[calc(100vw-2rem)] pointer-events-none">
      <div className="space-y-3 pointer-events-auto">
        {toasts.map((toast) => (
          <Toast key={toast.id} {...toast} onClose={onRemoveToast} />
        ))}
      </div>
    </div>
  )
}
