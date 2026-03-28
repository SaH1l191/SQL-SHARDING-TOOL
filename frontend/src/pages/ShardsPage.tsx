import { useEffect, useState, useCallback } from 'react'
import { StatusBadge } from '../components/StatusBadge'
import { EmptyState } from '../components/EmptyState'
import { Modal } from '../components/Modal'
import { useShardActions } from '../hooks/useShardActions'
import { useShardStore } from '../stores/shardStore'
import type { Project, ShardConnection } from '../types'

interface ConnectionForm {
  host: string
  port: string
  database_name: string
  username: string
  password: string
}

const DEFAULT_FORM: ConnectionForm = {
  host: 'localhost',
  port: '5432',
  database_name: '',
  username: 'postgres',
  password: '',
}

interface Props {
  project: Project | null
  onNoProject: () => void
}

export function ShardsPage({ project, onNoProject }: Props) {
  const { shards, connections, loading, error } = useShardStore()
  const {
    fetchShards,
    createShard,
    deleteShard,
    activateShard,
    deactivateShard,
    fetchShardConnection,
    updateShardConnectionInfo,
  } = useShardActions()

  const [creating, setCreating] = useState(false)
  const [connTarget, setConnTarget] = useState<string | null>(null)
  const [form, setForm] = useState<ConnectionForm>(DEFAULT_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  // Fetch shards whenever the selected project changes
  useEffect(() => {
    if (project?.id) {
      fetchShards(project.id)
    }
  }, [project?.id, fetchShards])

  const openConnModal = useCallback(async (shardId: string) => {
    setConnTarget(shardId)
    // Pre-populate form with existing connection if available
    try {
      const conn: ShardConnection | null = await fetchShardConnection(shardId)
      if (conn) {
        setForm({
          host: conn.host || 'localhost',
          port: String(conn.port || 5432),
          database_name: conn.database_name || '',
          username: conn.username || 'postgres',
          password: conn.password || '',
        })
      } else {
        setForm(DEFAULT_FORM)
      }
    } catch {
      setForm(DEFAULT_FORM)
    }
  }, [fetchShardConnection])

  if (!project) {
    return (
      <EmptyState
        icon="⬡"
        title="No project selected"
        description="Select a project from the sidebar to manage its shards."
        action={{ label: 'Go to projects', onClick: onNoProject }}
      />
    )
  }

  const handleCreate = async () => {
    setCreating(true)
    try { await createShard(project.id) }
    finally { setCreating(false) }
  }

  const handleSaveConn = async () => {
    if (!connTarget) return
    setSaving(true)
    try {
      await updateShardConnectionInfo({
        shard_id: connTarget,
        host: form.host,
        port: parseInt(form.port, 10) || 5432,
        database_name: form.database_name,
        username: form.username,
        password: form.password,
      })
      setConnTarget(null)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    await deleteShard(deleteTarget)
    setDeleteTarget(null)
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-gray-800 px-6 py-4">
        <div>
          <h1 className="text-base font-semibold text-gray-100">Shards</h1>
          <p className="text-xs text-gray-500 mt-0.5">
            {project.name} · {shards.length} shard{shards.length !== 1 ? 's' : ''}
          </p>
        </div>
        <button
          onClick={handleCreate}
          disabled={creating}
          className="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3 py-2 text-sm font-medium text-white hover:bg-brand-700 disabled:opacity-40 transition-colors"
        >
          {creating ? 'Adding…' : '+ Add shard'}
        </button>
      </div>

      <div className="flex-1 overflow-y-auto scrollbar-thin px-6 py-4">
        {loading && (
          <div className="flex items-center justify-center py-20">
            <div className="h-6 w-6 rounded-full border-2 border-brand-500 border-t-transparent animate-spin" />
          </div>
        )}
        {error && (
          <div className="rounded-lg border border-red-800 bg-red-900/20 px-4 py-3 text-sm text-red-400">{error}</div>
        )}
        {!loading && shards.length === 0 && (
          <EmptyState
            icon="⬡"
            title="No shards yet"
            description="Add shards to start distributing your data."
            action={{ label: 'Add shard', onClick: handleCreate }}
          />
        )}
        {shards.length > 0 && (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-800">
                <th className="pb-2 text-left text-xs font-medium text-gray-500">Index</th>
                <th className="pb-2 text-left text-xs font-medium text-gray-500">ID</th>
                <th className="pb-2 text-left text-xs font-medium text-gray-500">Status</th>
                <th className="pb-2 text-left text-xs font-medium text-gray-500">Created</th>
                <th className="pb-2" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800/60">
              {shards.map(s => (
                <tr key={s.id} className="group">
                  <td className="py-3 text-gray-400 font-mono text-xs">{s.shard_index}</td>
                  <td className="py-3 text-gray-500 font-mono text-xs">{s.id.slice(0, 8)}…</td>
                  <td className="py-3"><StatusBadge status={s.status} /></td>
                  <td className="py-3 text-gray-500 text-xs">{new Date(s.created_at).toLocaleDateString()}</td>
                  <td className="py-3 text-right">
                    <div className="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      {s.status === 'inactive' ? (
                        <button
                          onClick={() => activateShard(s.id)}
                          className="rounded px-2 py-1 text-[11px] font-medium text-emerald-400 hover:bg-emerald-900/30 transition-colors"
                        >
                          Activate
                        </button>
                      ) : (
                        <button
                          onClick={() => deactivateShard(s.id)}
                          className="rounded px-2 py-1 text-[11px] font-medium text-amber-400 hover:bg-amber-900/30 transition-colors"
                        >
                          Deactivate
                        </button>
                      )}
                      <button
                        onClick={() => openConnModal(s.id)}
                        className="rounded px-2 py-1 text-[11px] font-medium text-brand-400 hover:bg-brand-900/30 transition-colors"
                      >
                        Connection
                      </button>
                      <button
                        onClick={() => setDeleteTarget(s.id)}
                        className="rounded px-2 py-1 text-[11px] font-medium text-red-400 hover:bg-red-900/30 transition-colors"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Connection modal */}
      {connTarget && (
        <Modal
          title="Shard connection"
          onClose={() => setConnTarget(null)}
          footer={
            <>
              <button onClick={() => setConnTarget(null)} className="rounded-lg px-3 py-2 text-sm text-gray-400 hover:text-gray-200 transition-colors">
                Cancel
              </button>
              <button
                onClick={handleSaveConn}
                disabled={saving}
                className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white hover:bg-brand-700 disabled:opacity-40 transition-colors"
              >
                {saving ? 'Saving…' : 'Save'}
              </button>
            </>
          }
        >
          <div className="flex flex-col gap-3">
            {(['host', 'port', 'database_name', 'username', 'password'] as const).map(field => (
              <div key={field}>
                <label className="block text-xs font-medium text-gray-400 mb-1 capitalize">
                  {field.replace('_', ' ')}
                </label>
                <input
                  type={field === 'password' ? 'password' : 'text'}
                  value={form[field]}
                  onChange={e => setForm(f => ({ ...f, [field]: e.target.value }))}
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-600
                    focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500 transition-colors"
                />
              </div>
            ))}
          </div>
        </Modal>
      )}

      {/* Delete confirm */}
      {deleteTarget && (
        <Modal
          title="Delete shard"
          onClose={() => setDeleteTarget(null)}
          footer={
            <>
              <button onClick={() => setDeleteTarget(null)} className="rounded-lg px-3 py-2 text-sm text-gray-400 hover:text-gray-200 transition-colors">
                Cancel
              </button>
              <button
                onClick={handleDelete}
                className="rounded-lg bg-red-700 px-4 py-2 text-sm font-medium text-white hover:bg-red-600 transition-colors"
              >
                Delete
              </button>
            </>
          }
        >
          <p className="text-sm text-gray-300">
            Delete shard <span className="font-mono text-gray-100">{deleteTarget.slice(0, 8)}…</span>? This cannot be undone.
          </p>
        </Modal>
      )}
    </div>
  )
}