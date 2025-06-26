import React from 'react'
import { cn } from '@/lib/className'

interface CardProps {
  children: React.ReactNode
  className?: string
  variant?: 'default' | 'gradient' | 'glass'
  hover?: boolean
}

export const Card: React.FC<CardProps> = ({ 
  children, 
  className, 
  variant = 'default',
  hover = true 
}) => {
  const baseClasses = 'rounded-3xl shadow-2xl border transition-all duration-300'
  
  const variants = {
    default: 'bg-white/5 backdrop-blur-xl border-white/10',
    gradient: 'bg-gradient-to-br from-white/10 to-white/5 backdrop-blur-xl border-white/20',
    glass: 'bg-white/5 backdrop-blur-xl border-white/10 shadow-lg'
  }
  
  const hoverClasses = hover ? 'hover:border-white/20 hover:shadow-3xl' : ''
  
  return (
    <div className={cn(
      baseClasses,
      variants[variant],
      hoverClasses,
      className
    )}>
      {children}
    </div>
  )
}

interface CardHeaderProps {
  children: React.ReactNode
  className?: string
}

export const CardHeader: React.FC<CardHeaderProps> = ({ children, className }) => (
  <div className={cn('p-6 pb-0', className)}>
    {children}
  </div>
)

interface CardContentProps {
  children: React.ReactNode
  className?: string
}

export const CardContent: React.FC<CardContentProps> = ({ children, className }) => (
  <div className={cn('p-6', className)}>
    {children}
  </div>
)

interface CardTitleProps {
  children: React.ReactNode
  className?: string
}

export const CardTitle: React.FC<CardTitleProps> = ({ children, className }) => (
  <h3 className={cn('text-xl font-bold text-white', className)}>
    {children}
  </h3>
)

interface CardDescriptionProps {
  children: React.ReactNode
  className?: string
}

export const CardDescription: React.FC<CardDescriptionProps> = ({ children, className }) => (
  <p className={cn('text-slate-400 text-sm', className)}>
    {children}
  </p>
)