import { useState, useEffect, useCallback } from 'react'

/**
 * Custom hook for managing state in URL search parameters
 * Automatically syncs state with URL and preserves it across page refreshes
 */
export function useURLState<T extends string>(
  key: string,
  defaultValue: T,
  validValues?: T[]
): [T, (value: T) => void] {
  // Initialize state from URL or default value
  const [state, setState] = useState<T>(() => {
    const params = new URLSearchParams(window.location.search)
    const urlValue = params.get(key) as T
    
    // Validate the URL value if validation array is provided
    if (validValues && urlValue && validValues.includes(urlValue)) {
      return urlValue
    }
    
    return urlValue || defaultValue
  })

  // Update URL when state changes
  const updateState = useCallback((newValue: T) => {
    setState(newValue)
    
    const url = new URL(window.location.href)
    
    if (newValue === defaultValue) {
      // Remove parameter if it's the default value to keep URL clean
      url.searchParams.delete(key)
    } else {
      url.searchParams.set(key, newValue)
    }
    
    // Update URL without triggering page refresh
    window.history.replaceState({}, '', url.toString())
  }, [key, defaultValue])

  // Listen for browser back/forward navigation
  useEffect(() => {
    const handlePopState = () => {
      const params = new URLSearchParams(window.location.search)
      const urlValue = params.get(key) as T
      
      if (validValues && urlValue && validValues.includes(urlValue)) {
        setState(urlValue)
      } else {
        setState(urlValue || defaultValue)
      }
    }

    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [key, defaultValue, validValues])

  return [state, updateState]
}