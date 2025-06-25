import { useState } from 'react'

interface ValueDisplayProps {
  value: string
  maxLength?: number
}

export const ValueDisplay = ({ value, maxLength = 100 }: ValueDisplayProps) => {
  const [isExpanded, setIsExpanded] = useState(false)
  
  // Try to parse as JSON for better formatting
  const isJSON = (() => {
    try {
      JSON.parse(value)
      return true
    } catch {
      return false
    }
  })()
  
  const formatValue = (val: string) => {
    if (isJSON && isExpanded) {
      try {
        return JSON.stringify(JSON.parse(val), null, 2)
      } catch {
        return val
      }
    }
    return val
  }
  
  const displayValue = formatValue(value)
  const shouldTruncate = displayValue.length > maxLength
  const truncatedValue = shouldTruncate && !isExpanded 
    ? displayValue.slice(0, maxLength) + '...'
    : displayValue

  return (
    <div className="font-mono text-slate-300">
      {isJSON && isExpanded ? (
        <pre className="bg-slate-900/50 p-3 rounded-lg text-xs overflow-x-auto whitespace-pre-wrap max-w-md">
          {truncatedValue}
        </pre>
      ) : (
        <span className="bg-slate-800/50 px-3 py-1 rounded-lg inline-block max-w-md break-words">
          {truncatedValue}
        </span>
      )}
      
      {shouldTruncate && (
        <button
          onClick={() => setIsExpanded(!isExpanded)}
          className="ml-2 text-xs text-purple-400 hover:text-purple-300 underline"
        >
          {isExpanded ? 'Show less' : 'Show more'}
        </button>
      )}
    </div>
  )
}