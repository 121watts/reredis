import React from 'react'
import { cn } from '@/lib/className'
import type { ButtonVariant, ButtonSize } from '@/types/ui'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
  loading?: boolean
  children: React.ReactNode
}

export const Button: React.FC<ButtonProps> = ({ 
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled,
  className,
  children,
  ...props 
}) => {
  const baseClasses = 'inline-flex items-center justify-center font-medium rounded-lg transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-transparent'
  
  const variants = {
    primary: 'bg-purple-500 text-white hover:bg-purple-600 focus:ring-purple-500 shadow-md hover:shadow-lg',
    secondary: 'bg-white/10 text-slate-300 hover:bg-white/20 hover:text-white focus:ring-white/20',
    success: 'bg-emerald-500 text-white hover:bg-emerald-600 focus:ring-emerald-500 shadow-md hover:shadow-lg',
    warning: 'bg-amber-500 text-white hover:bg-amber-600 focus:ring-amber-500 shadow-md hover:shadow-lg',
    danger: 'bg-red-500 text-white hover:bg-red-600 focus:ring-red-500 shadow-md hover:shadow-lg',
    ghost: 'text-slate-300 hover:text-white hover:bg-white/10 focus:ring-white/20'
  }
  
  const sizes = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-sm',
    lg: 'px-6 py-3 text-base'
  }
  
  const disabledClasses = 'opacity-50 cursor-not-allowed'
  
  return (
    <button
      className={cn(
        baseClasses,
        variants[variant],
        sizes[size],
        (disabled || loading) && disabledClasses,
        className
      )}
      disabled={disabled || loading}
      {...props}
    >
      {loading && (
        <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin mr-2" />
      )}
      {children}
    </button>
  )
}