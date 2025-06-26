import React from 'react'
import { cn } from '@/lib/className'

interface MetricProps {
  label: string
  value: string | number
  icon?: string
  trend?: 'up' | 'down' | 'stable'
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info'
  className?: string
}

export const Metric: React.FC<MetricProps> = ({ 
  label, 
  value, 
  icon,
  trend,
  variant = 'default',
  className 
}) => {
  const variants = {
    default: 'from-purple-500 to-purple-600',
    success: 'from-emerald-500 to-emerald-600',
    warning: 'from-amber-500 to-amber-600',
    danger: 'from-red-500 to-red-600',
    info: 'from-blue-500 to-blue-600'
  }
  
  const trendIcons = {
    up: '📈',
    down: '📉',
    stable: '➡️'
  }
  
  const formatValue = (val: string | number) => {
    if (typeof val === 'number') {
      return val.toLocaleString()
    }
    return val
  }
  
  return (
    <div className={cn(
      'bg-gradient-to-r rounded-xl p-6 text-white shadow-lg',
      variants[variant],
      className
    )}>
      <div className="flex items-center justify-between">
        <div>
          <div className="text-3xl font-bold">{formatValue(value)}</div>
          <div className="text-white/80 text-sm">{label}</div>
        </div>
        <div className="text-4xl opacity-80">
          {icon || (trend && trendIcons[trend]) || '📊'}
        </div>
      </div>
      {trend && (
        <div className="mt-2 flex items-center text-xs text-white/70">
          <span className="mr-1">{trendIcons[trend]}</span>
          <span className="capitalize">{trend}</span>
        </div>
      )}
    </div>
  )
}