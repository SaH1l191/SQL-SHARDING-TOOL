import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

interface ProjectSchema {
  id: string
  project_id: string
  ddl_sql: string
  state: 'draft' | 'pending' | 'applying' | 'applied' | 'failed'
  created_at: string
  updated_at: string
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

interface SchemaState {
  schemas: ProjectSchema[]
  executionStatuses: Record<string, SchemaExecutionStatus[]>
  currentSchema: ProjectSchema | null
  loading: boolean
  error: string | null
  
  // Actions
  setSchemas: (schemas: ProjectSchema[]) => void
  setCurrentSchema: (schema: ProjectSchema | null) => void
  addSchema: (schema: ProjectSchema) => void
  updateSchema: (id: string, updates: Partial<ProjectSchema>) => void
  removeSchema: (id: string) => void
  setExecutionStatuses: (schemaId: string, statuses: SchemaExecutionStatus[]) => void
  updateExecutionStatus: (schemaId: string, shardId: string, updates: Partial<SchemaExecutionStatus>) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
  clearSchemas: () => void
}

export const useSchemaStore = create<SchemaState>()(
  devtools(
    (set, get) => ({
      schemas: [],
      executionStatuses: {},
      currentSchema: null,
      loading: false,
      error: null,

      setSchemas: (schemas) => set({ schemas }),
      
      setCurrentSchema: (schema) => set({ currentSchema: schema }),
      
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
      
      setError: (error) => set({ error }),
      
      clearError: () => set({ error: null }),
      
      clearSchemas: () => set({ schemas: [], executionStatuses: {}, currentSchema: null })
    }),
    {
      name: 'schema-store'
    }
  )
)
