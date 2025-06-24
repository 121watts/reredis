import type { ReactNode } from 'react'
import { BackgroundEffects } from '@/components/BackgroundEffects'

interface LayoutProps {
  children: ReactNode
}

export const Layout = ({ children }: LayoutProps) => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-slate-100 font-sans relative overflow-hidden">
      <BackgroundEffects />

      <div className="relative z-0 container mx-auto p-4 md:p-8 max-w-6xl">
        {children}
      </div>
    </div>
  )
}
