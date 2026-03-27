import { useState, useCallback } from 'react'
import { GetShards, CreateShard } from '../../wailsjs/go/main/App'
import type { Shard } from '../types'

export function useShards(projectId: string | null) {
  const [shards, setShards] = useState<Shard[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetch = useCallback(async (pid?: string) => {
    const id = pid ?? projectId
    if (!id) return
    setLoading(true)
    setError(null)
    try {
      const data = await (GetShards as unknown as (id: string) => Promise<Shard[]>)(id)
      setShards(data ?? [])
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }, [projectId])

  const create = useCallback(async (pid: string) => {
    const s = await (CreateShard as unknown as (id: string) => Promise<Shard>)(pid)
    setShards(prev => [...prev, s])
    return s
  }, [])

  return { shards, loading, error, fetchShards: fetch, createShard: create }
}
