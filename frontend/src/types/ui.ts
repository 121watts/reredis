/**
 * UI component type definitions
 * Provides consistent typing for reusable UI components
 */

export type ButtonVariant = 'primary' | 'secondary' | 'success' | 'warning' | 'danger' | 'ghost'
export type ButtonSize = 'sm' | 'md' | 'lg'

export type CardVariant = 'default' | 'gradient' | 'glass'

export type BadgeVariant = 'default' | 'success' | 'warning' | 'danger' | 'info' | 'outline'
export type BadgeSize = 'sm' | 'md' | 'lg'

export type ProgressVariant = 'default' | 'success' | 'warning' | 'danger'
export type ProgressSize = 'sm' | 'md' | 'lg'

export type MetricVariant = 'default' | 'success' | 'warning' | 'danger' | 'info'
export type MetricTrend = 'up' | 'down' | 'stable'

export type InputVariant = 'default' | 'filled'

export type TabType = 'data' | 'cluster'
export type ViewMode = 'grid' | 'topology'

/**
 * Common props that many components share
 */
export interface BaseComponentProps {
  className?: string
  children?: React.ReactNode
}

export interface VariantProps<T = string> {
  variant?: T
}

export interface SizeProps<T = string> {
  size?: T
}