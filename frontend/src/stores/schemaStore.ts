import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

export interface ProjectSchema {
  id: string
  project_id: string
  ddl_sql: string
  version?: number
  state: string
  error_message?: string
  created_at: string
  updated_at?: string
  commited_at?: string
}

interface SchemaExecutionStatus {
  id: string
  schema_id: string
  shard_id: string
  state: 'pending' | 'applying' | 'applied' | 'failed'
  error_message?: string
  created_at: string
  updated_at: string
}

export interface SchemaCapabilities {
  can_create_draft: boolean
  can_edit_draft: boolean
  can_commit: boolean
  can_execute: boolean
  can_retry: boolean
  reason?: string
}

interface SchemaState {
  schemas: ProjectSchema[]
  schemaHistory: ProjectSchema[]
  executionStatuses: Record<string, SchemaExecutionStatus[]>
  currentSchema: ProjectSchema | null
  capabilities: SchemaCapabilities | null
  loading: boolean
  saving: boolean
  executing: boolean
  error: string | null
  
  // Actions
  setSchemas: (schemas: ProjectSchema[]) => void
  setSchemaHistory: (history: ProjectSchema[]) => void
  setCurrentSchema: (schema: ProjectSchema | null) => void
  setCapabilities: (caps: SchemaCapabilities | null) => void
  addSchema: (schema: ProjectSchema) => void
  updateSchema: (id: string, updates: Partial<ProjectSchema>) => void
  removeSchema: (id: string) => void
  setExecutionStatuses: (schemaId: string, statuses: SchemaExecutionStatus[]) => void
  updateExecutionStatus: (schemaId: string, shardId: string, updates: Partial<SchemaExecutionStatus>) => void
  setLoading: (loading: boolean) => void
  setSaving: (saving: boolean) => void
  setExecuting: (executing: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
  clearSchemas: () => void
}

export const useSchemaStore = create<SchemaState>()(
  devtools(
    (set, get) => ({
      schemas: [],
      schemaHistory: [],
      executionStatuses: {},
      currentSchema: null,
      capabilities: null,
      loading: false,
      saving: false,
      executing: false,
      error: null,

      setSchemas: (schemas) => set({ schemas }),
      
      setSchemaHistory: (history) => set({ schemaHistory: history }),
      
      setCurrentSchema: (schema) => set({ currentSchema: schema }),
      
      setCapabilities: (caps) => set({ capabilities: caps }),
      
      addSchema: (schema) => set((state) => ({
        schemas: [schema, ...state.schemas]
      })),
      
      updateSchema: (id, updates) => set((state) => ({
        schemas: state.schemas.map(s => 
          s.id === id ? { ...s, ...updates } : s
        ),
        currentSchema: state.currentSchema?.id === id 
          ? { ...state.currentSchema, ...updates } 
          : state.currentSchema
      })),
      
      removeSchema: (id) => set((state) => ({
        schemas: state.schemas.filter(s => s.id !== id),
        currentSchema: state.currentSchema?.id === id ? null : state.currentSchema,
        executionStatuses: Object.fromEntries(
          Object.entries(state.executionStatuses).filter(([schemaId]) => schemaId !== id)
        )
      })),
      
      setExecutionStatuses: (schemaId, statuses) => set((state) => ({
        executionStatuses: { ...state.executionStatuses, [schemaId]: statuses }
      })),
      
      updateExecutionStatus: (schemaId, shardId, updates) => set((state) => ({
        executionStatuses: {
          ...state.executionStatuses,
          [schemaId]: state.executionStatuses[schemaId]?.map(status =>
            status.shard_id === shardId ? { ...status, ...updates } : status
          ) || []
        }
      })),
      
      setLoading: (loading) => set({ loading }),
      
      setSaving: (saving) => set({ saving }),
      
      setExecuting: (executing) => set({ executing }),
      
      setError: (error) => set({ error }),
      
      clearError: () => set({ error: null }),
      
      clearSchemas: () => set({ schemas: [], schemaHistory: [], executionStatuses: {}, currentSchema: null, capabilities: null })
    }),
    {
      name: 'schema-store'
    }
  )
)
