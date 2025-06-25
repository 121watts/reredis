import { useState, useCallback, useEffect } from 'react'

interface PaginationResponse {
  keys: string[]
  next_cursor: string
  has_more: boolean
}

interface UsePaginationResult {
  keys: string[]
  loading: boolean
  hasMore: boolean
  loadMore: () => void
  reset: () => void
}

export const usePagination = (
  url: string,
  limit: number = 20
): UsePaginationResult => {
  const [keys, setKeys] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [cursor, setCursor] = useState('')
  const [initialized, setInitialized] = useState(false)

  const loadMore = useCallback(async () => {
    if (loading || !hasMore) return

    setLoading(true)

    try {
      const params = new URLSearchParams({
        limit: limit.toString(),
        ...(cursor && { cursor }),
      })

      const response = await fetch(`${url}?${params}`)

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data: PaginationResponse = await response.json()

      setKeys((prevKeys) => [...prevKeys, ...data.keys])
      setCursor(data.next_cursor)
      setHasMore(data.has_more)
    } catch (error) {
      console.error('Failed to load more keys:', error)
      setHasMore(false)
    } finally {
      setLoading(false)
    }
  }, [url, limit, cursor, loading, hasMore])

  const reset = useCallback(() => {
    setKeys([])
    setCursor('')
    setHasMore(true)
    setLoading(false)
  }, [])

  // Auto-load first page on mount
  useEffect(() => {
    if (!initialized) {
      setInitialized(true)
      loadMore()
    }
  }, [initialized, loadMore])

  return {
    keys,
    loading,
    hasMore,
    loadMore,
    reset,
  }
}
