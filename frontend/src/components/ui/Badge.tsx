import React from 'react'
import { cn } from '@/lib/className'
import type { BadgeVariant, BadgeSize } from '@/types/ui'

interface BadgeProps {
  children: React.ReactNode
  variant?: BadgeVariant
  size?: BadgeSize
  className?: string
}

export const Badge: React.FC<BadgeProps> = ({ 
  children, 
  variant = 'default',
  size = 'md',
  className 
}) => {
  const baseClasses = 'inline-flex items-center font-semibold rounded-full border'
  
  const variants = {
    default: 'bg-white/10 text-slate-300 border-white/20',
    success: 'bg-emerald-500/30 text-emerald-200 border-emerald-400/30',
    warning: 'bg-amber-500/30 text-amber-200 border-amber-400/30',
    danger: 'bg-red-500/30 text-red-200 border-red-400/30',
    info: 'bg-blue-500/30 text-blue-200 border-blue-400/30',
    outline: 'bg-transparent text-slate-300 border-white/30'
  }
  
  const sizes = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-3 py-1 text-sm',
    lg: 'px-4 py-1.5 text-base'
  }
  
  return (
    <span className={cn(
      baseClasses,
      variants[variant],
      sizes[size],
      className
    )}>
      {children}
    </span>
  )
}