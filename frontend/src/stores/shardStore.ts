import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import type { Shard, ShardConnection } from '../types'

interface ShardState {
  shards: Shard[]
  connections: Record<string, ShardConnection>
  loading: boolean
  error: string | null
  
  // Actions
  setShards: (shards: Shard[]) => void
  addShard: (shard: Shard) => void
  updateShard: (id: string, updates: Partial<Shard>) => void
  removeShard: (id: string) => void
  setConnection: (shardId: string, connection: ShardConnection) => void
  updateConnection: (shardId: string, updates: Partial<ShardConnection>) => void
  removeConnection: (shardId: string) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
  clearShards: () => void
}

export const useShardStore = create<ShardState>()(
  devtools(
    (set, get) => ({
      shards: [],
      connections: {},
      loading: false,
      error: null,

      setShards: (shards) => set({ shards }),
      
      addShard: (shard) => set((state) => ({
        shards: [...state.shards, shard]
      })),
      
      updateShard: (id, updates) => set((state) => ({
        shards: state.shards.map(s => 
          s.id === id ? { ...s, ...updates } : s
        )
      })),
      
      removeShard: (id) => set((state) => ({
        shards: state.shards.filter(s => s.id !== id),
        connections: Object.fromEntries(
          Object.entries(state.connections).filter(([shardId]) => shardId !== id)
        )
      })),
      
      setConnection: (shardId, connection) => set((state) => ({
        connections: { ...state.connections, [shardId]: connection }
      })),
      
      updateConnection: (shardId, updates) => set((state) => ({
        connections: {
          ...state.connections,
          [shardId]: state.connections[shardId] 
            ? { ...state.connections[shardId], ...updates }
            : updates as ShardConnection
        }
      })),
      
      removeConnection: (shardId) => set((state) => {
        const newConnections = { ...state.connections }
        delete newConnections[shardId]
        return { connections: newConnections }
      }),
      
      setLoading: (loading) => set({ loading }),
      
      setError: (error) => set({ error }),
      
      clearError: () => set({ error: null }),
      
      clearShards: () => set({ shards: [], connections: {} })
    }),
    {
      name: 'shard-store'
    }
  )
)
