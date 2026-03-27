import { useState, useEffect, useCallback } from 'react'
import { GetProjects, CreateProject, DeleteProject, ActivateProject } from '../../wailsjs/go/main/App'
import type { Project } from '../types'

export function useProjects() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetch = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      // Wails passes context automatically — no need to pass it from TS
      const data = await (GetProjects as unknown as () => Promise<Project[]>)()
      setProjects(data ?? [])
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { fetch() }, [fetch])

  const create = useCallback(async (name: string, description: string) => {
    const p = await (CreateProject as unknown as (n: string, d: string) => Promise<Project>)(name, description)
    setProjects(prev => [p, ...prev])
    return p
  }, [])

  const remove = useCallback(async (id: string) => {
    await (DeleteProject as unknown as (id: string) => Promise<void>)(id)
    setProjects(prev => prev.filter(p => p.id !== id))
  }, [])

  const activate = useCallback(async (id: string) => {
    await (ActivateProject as unknown as (id: string) => Promise<void>)(id)
    await fetch()
  }, [fetch])

  return { projects, loading, error, refetch: fetch, create, remove, activate }
}
