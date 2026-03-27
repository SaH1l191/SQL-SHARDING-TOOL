import { useState, useCallback } from 'react'
import { EmptyState } from '../components/EmptyState'
import { ApplyDDL, GetSchemaHistory } from '../../wailsjs/go/main/App'
import type { Project } from '../types'

interface SchemaVersion {
  id: string
  version: number
  state: string
  created_at: string
}

interface Props {
  project: Project | null
  onNoProject: () => void
}

const PLACEHOLDER = `-- Paste your DDL here, e.g.:
CREATE TABLE users (
    id   UUID PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE orders (
    id      UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    amount  NUMERIC(10,2) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);`

export function SchemaPage({ project, onNoProject }: Props) {
  const [ddl, setDdl] = useState('')
  const [status, setStatus] = useState<'idle' | 'loading' | 'ok' | 'error'>('idle')
  const [message, setMessage] = useState('')
  const [versions, setVersions] = useState<SchemaVersion[]>([])
  const [loadingVersions, setLoadingVersions] = useState(false)

  const loadSchemaVersions = useCallback(async (projectId: string) => {
    setLoadingVersions(true)
    try {
      const data = await (GetSchemaHistory as unknown as (id: string) => Promise<SchemaVersion[]>)(projectId)
      setVersions(data ?? [])
    } catch (e) {
      console.error('Failed to load schema history', e)
    } finally {
      setLoadingVersions(false)
    }
  }, [])

  if (!project) {
    return (
      <div className="flex items-center justify-center h-full">
        <EmptyState
          icon="⊞"
          title="No project selected"
          description="Select a project to manage its schema."
          action={{ label: 'Go to projects', onClick: onNoProject }}
        />
      </div>
    )
  }

  const handleApply = async () => {
    if (!ddl.trim()) return
    setStatus('loading')
    setMessage('')
    try {
      await (ApplyDDL as unknown as (pid: string, ddl: string) => Promise<void>)(project.id, ddl.trim())
      setStatus('ok')
      setMessage('Schema applied — shard key inference complete.')
      loadSchemaVersions(project.id)
    } catch (e) {
      setStatus('error')
      setMessage(String(e))
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-gray-800 px-6 py-4">
        <div>
          <h1 className="text-base font-semibold text-gray-100">Schema</h1>
          <p className="text-xs text-gray-500 mt-0.5">{project.name}</p>
        </div>
        <button
          onClick={handleApply}
          disabled={status === 'loading' || !ddl.trim()}
          className="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3 py-2 text-sm font-medium text-white hover:bg-brand-700 disabled:opacity-40 transition-colors"
        >
          {status === 'loading' ? (
            <>
              <span className="h-3 w-3 rounded-full border border-white border-t-transparent animate-spin" />
              Applying…
            </>
          ) : 'Apply DDL'}
        </button>
      </div>

      <div className="flex flex-1 overflow-hidden">
        {/* DDL editor */}
        <div className="flex flex-col flex-1 border-r border-gray-800">
          <div className="px-4 py-2 border-b border-gray-800 flex items-center justify-between">
            <span className="text-xs font-medium text-gray-500">DDL input</span>
            {ddl && (
              <button onClick={() => setDdl('')} className="text-xs text-gray-600 hover:text-gray-400 transition-colors">
                Clear
              </button>
            )}
          </div>
          <textarea
            value={ddl}
            onChange={e => setDdl(e.target.value)}
            placeholder={PLACEHOLDER}
            spellCheck={false}
            className="flex-1 resize-none bg-gray-950 px-4 py-3 font-mono text-xs text-gray-300 placeholder-gray-700
              focus:outline-none scrollbar-thin leading-relaxed"
          />
          {message && (
            <div className={`px-4 py-2 text-xs border-t border-gray-800 ${status === 'ok' ? 'text-emerald-400' : 'text-red-400'}`}>
              {status === 'ok' ? '✓ ' : '✗ '}{message}
            </div>
          )}
        </div>

        {/* Version history */}
        <div className="w-64 flex flex-col shrink-0">
          <div className="px-4 py-2 border-b border-gray-800 flex items-center justify-between">
            <span className="text-xs font-medium text-gray-500">Versions</span>
            <button
              onClick={() => loadSchemaVersions(project.id)}
              disabled={loadingVersions}
              className="text-xs text-gray-600 hover:text-gray-400 transition-colors"
            >
              {loadingVersions ? '…' : '↻'}
            </button>
          </div>
          <ul className="flex-1 overflow-y-auto scrollbar-thin divide-y divide-gray-800/40">
            {versions.length === 0 && (
              <li className="px-4 py-6 text-center text-xs text-gray-600">No versions yet</li>
            )}
            {versions.map(v => (
              <li key={v.id} className="px-4 py-2.5">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-gray-300">v{v.version}</span>
                  <span className={`text-[10px] rounded px-1.5 py-0.5 font-medium
                    ${v.state === 'applied' ? 'bg-emerald-900/50 text-emerald-400' : 'bg-gray-800 text-gray-500'}`}>
                    {v.state}
                  </span>
                </div>
                <div className="text-[10px] text-gray-600 mt-0.5">
                  {new Date(v.created_at).toLocaleDateString()}
                </div>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  )
}