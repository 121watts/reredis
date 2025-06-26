import React from 'react'
import { cn } from '@/lib/className'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
  variant?: 'default' | 'filled'
}

export const Input: React.FC<InputProps> = ({ 
  label,
  error,
  variant = 'default',
  className,
  ...props 
}) => {
  const baseClasses = 'w-full px-3 py-2 rounded-lg transition-all duration-200 focus:outline-none focus:ring-2'
  
  const variants = {
    default: 'border border-white/20 bg-white/10 text-slate-200 placeholder-slate-400 focus:ring-purple-500 focus:border-purple-500 backdrop-blur-sm',
    filled: 'border-0 bg-white/20 text-slate-200 placeholder-slate-400 focus:ring-purple-500 focus:bg-white/30 backdrop-blur-sm'
  }
  
  const errorClasses = error ? 'border-red-400 focus:border-red-400 focus:ring-red-500' : ''
  
  return (
    <div className="space-y-2">
      {label && (
        <label className="block text-sm font-medium text-slate-300">
          {label}
        </label>
      )}
      <input
        className={cn(
          baseClasses,
          variants[variant],
          errorClasses,
          className
        )}
        {...props}
      />
      {error && (
        <p className="text-red-400 text-sm">{error}</p>
      )}
    </div>
  )
}