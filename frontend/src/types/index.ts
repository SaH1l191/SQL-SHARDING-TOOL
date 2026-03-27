export interface Project {
  id: string
  name: string
  description: string
  status: 'active' | 'inactive'
  shard_count: number
  created_at: string
  updated_at: string
}

export interface Shard {
  id: string
  project_id: string
  shard_index: number
  status: 'active' | 'inactive'
  created_at: string
}

export interface ShardConnection {
  shard_id: string
  host: string
  port: number
  database_name: string
  username: string
  password: string
  created_at: string
  updated_at: string
}

export type View = 'projects' | 'shards' | 'schema' | 'query'
