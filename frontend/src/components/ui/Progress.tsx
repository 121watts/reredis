import React from 'react'
import { cn } from '@/lib/className'

interface ProgressProps {
  value: number
  max?: number
  variant?: 'default' | 'success' | 'warning' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  showLabel?: boolean
  className?: string
}

export const Progress: React.FC<ProgressProps> = ({ 
  value,
  max = 100,
  variant = 'default',
  size = 'md',
  showLabel = false,
  className 
}) => {
  const percentage = Math.min((value / max) * 100, 100)
  
  const trackClasses = 'w-full bg-white/10 rounded-full overflow-hidden'
  
  const sizes = {
    sm: 'h-1',
    md: 'h-2',
    lg: 'h-3'
  }
  
  const variants = {
    default: 'bg-gradient-to-r from-purple-400 to-purple-500',
    success: 'bg-gradient-to-r from-emerald-400 to-emerald-500',
    warning: 'bg-gradient-to-r from-amber-400 to-amber-500',
    danger: 'bg-gradient-to-r from-red-400 to-red-500'
  }
  
  const getVariantByValue = (val: number) => {
    if (val >= 80) return 'success'
    if (val >= 60) return 'warning'
    if (val < 40) return 'danger'
    return 'default'
  }
  
  const autoVariant = variant === 'default' ? getVariantByValue(percentage) : variant
  
  return (
    <div className={cn('space-y-2', className)}>
      {showLabel && (
        <div className="flex justify-between text-sm">
          <span className="text-slate-300">{Math.round(percentage)}%</span>
          <span className="text-slate-400">{value}/{max}</span>
        </div>
      )}
      <div className={cn(trackClasses, sizes[size])}>
        <div
          className={cn(
            'h-full transition-all duration-300 ease-out',
            variants[autoVariant]
          )}
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  )
}