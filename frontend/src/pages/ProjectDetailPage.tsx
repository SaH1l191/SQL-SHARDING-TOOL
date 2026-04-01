import { useState, useEffect } from 'react'
import { StatusBadge } from '../components/StatusBadge'
import { EmptyState } from '../components/EmptyState'
import { Modal } from '../components/Modal'
import { SchemaEditor } from '../components/SchemaEditor'
import { useShardActions } from '../hooks/useShardActions'
import { useProjectActions } from '../hooks/useProjectActions'
import { useShardStore } from '../stores/shardStore'
import type { Project, Shard, ShardConnection } from '../types'

interface Props {
  project: Project
  onBack: () => void
}

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

export function ProjectDetailPage({ project, onBack }: Props) {
  const [activeTab, setActiveTab] = useState<'overview' | 'shards' | 'schema'>('overview')
  const { shards, loading, error } = useShardStore()
  const { 
    fetchShards,
    createShard,
    deleteShard,
    activateShard,
    deactivateShard,
    fetchShardConnection,
    updateShardConnectionInfo,
  } = useShardActions()
  
  const { deleteProject } = useProjectActions()
  
  const [connTarget, setConnTarget] = useState<string | null>(null)
  const [form, setForm] = useState<ConnectionForm>(DEFAULT_FORM)
  const [saving, setSaving] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const [deleteProjectTarget, setDeleteProjectTarget] = useState<Project | null>(null)

  useEffect(() => {
    if (project?.id) {
      fetchShards(project.id)
    }
  }, [project?.id, fetchShards])

  const openConnModal = async (shardId: string) => {
    setConnTarget(shardId)
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
  }

  const handleCreateShard = async () => {
    try { 
      await createShard(project.id) 
    } catch (error) {
      // Error handled by hook
    }
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

  const handleDeleteShard = async () => {
    if (!deleteTarget) return
    try {
      await deleteShard(deleteTarget, project.id)
      setDeleteTarget(null)
    } catch (error) {
      // Error handled by hook, just close modal
      setDeleteTarget(null)
    }
  }

  const handleDeleteProject = async () => {
    if (!deleteProjectTarget) return
    await deleteProject(deleteProjectTarget.id)
    setDeleteProjectTarget(null)
    onBack()
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-gray-800 px-6 py-4">
        <div className="flex items-center gap-4">
          <button
            onClick={onBack}
            className="text-gray-500 hover:text-brand-400 transition-colors text-sm leading-none"
          >
            ← Back to Projects
          </button>
          <div>
            <h1 className="text-xl font-semibold text-gray-100">{project.name}</h1>
            <p className="text-sm text-gray-500 mt-1">{project.description}</p>
          </div>
        </div>
        <button
          onClick={() => setDeleteProjectTarget(project)}
          className="rounded-lg px-3 py-2 text-sm font-medium text-red-600 hover:bg-red-900/20 transition-colors"
        >
          Delete Project
        </button>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-gray-800">
        {(['overview', 'shards', 'schema'] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-6 py-3 text-sm font-medium transition-colors border-b-2 ${
              activeTab === tab
                ? 'text-brand-400 border-brand-400'
                : 'text-gray-500 border-transparent hover:text-gray-300 hover:border-gray-700'
            }`}
          >
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto scrollbar-thin">
        {loading && (
          <div className="flex items-center justify-center py-20">
            <div className="h-8 w-8 rounded-full border-2 border-brand-500 border-t-transparent animate-spin" />
          </div>
        )}

        {error && (
          <div className="rounded-lg border border-red-800 bg-red-900/20 px-4 py-3 text-sm text-red-400 m-4">
            {error}
          </div>
        )}

        {!loading && !error && (
          <>
            {activeTab === 'overview' && (
              <div className="p-6">
                <h2 className="text-lg font-semibold text-gray-100 mb-4">Project Overview</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  <div className="rounded-lg border border-gray-800 bg-gray-900 p-4">
                    <h3 className="text-sm font-medium text-gray-400 mb-2">Project Status</h3>
                    <div className="flex items-center gap-2">
                      <StatusBadge status={project.status} />
                      <span className="text-sm text-gray-500">
                        {project.status === 'active' ? 'Project is currently active' : 'Project is inactive'}
                      </span>
                    </div>
                  </div>
                  <div className="rounded-lg border border-gray-800 bg-gray-900 p-4">
                    <h3 className="text-sm font-medium text-gray-400 mb-2">Shard Count</h3>
                    <p className="text-2xl font-bold text-gray-100">{shards.length}</p>
                    <p className="text-xs text-gray-500 mt-1">
                      {shards.length === 0 ? 'No shards configured' : 
                       shards.length === 1 ? '1 shard configured' : 
                       `${shards.length} shards configured`}
                    </p>
                  </div>
                  <div className="rounded-lg border border-gray-800 bg-gray-900 p-4">
                    <h3 className="text-sm font-medium text-gray-400 mb-2">Created</h3>
                    <p className="text-sm text-gray-500">
                      {formatDate(project.created_at)}
                    </p>
                  </div>
                  <div className="rounded-lg border border-gray-800 bg-gray-900 p-4">
                    <h3 className="text-sm font-medium text-gray-400 mb-2">Last Updated</h3>
                    <p className="text-sm text-gray-500">
                      {formatDate(project.updated_at)}
                    </p>
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'shards' && (
              <div className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold text-gray-100">Shards</h2>
                  <button
                    onClick={handleCreateShard}
                    className="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3 py-2 text-sm font-medium text-white hover:bg-brand-700 transition-colors"
                  >
                    <span className="text-base leading-none">+</span> Add Shard
                  </button>
                </div>

                {shards.length === 0 ? (
                  <EmptyState
                    icon="⬡"
                    title="No shards yet"
                    description="Add shards to start distributing your data."
                    action={{ label: 'Add shard', onClick: handleCreateShard }}
                  />
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-gray-800">
                          <th className="pb-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Index</th>
                          <th className="pb-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                          <th className="pb-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                          <th className="pb-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Info</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-800/60">
                        {shards.map((shard: Shard) => (
                          <tr key={shard.id} className="hover:bg-gray-800/50 transition-colors group">
                            <td className="py-3 text-gray-400 font-mono text-sm">{shard.shard_index}</td>
                            <td className="py-3">
                              <StatusBadge status={shard.status} />
                            </td>
                            <td className="py-3 text-gray-500 text-sm">
                              {formatDate(shard.created_at)}
                            </td>
                            <td className="py-3 text-right">
                              <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                {shard.status === 'inactive' && (
                                  <button
                                    onClick={() => activateShard(shard.id)}
                                    className="rounded px-2 py-1 text-xs font-medium text-emerald-400 hover:bg-emerald-900/30 transition-colors"
                                  >
                                    Activate
                                  </button>
                                )}
                                {shard.status === 'active' && (
                                  <button
                                    onClick={() => deactivateShard(shard.id)}
                                    className="rounded px-2 py-1 text-xs font-medium text-amber-400 hover:bg-amber-900/30 transition-colors"
                                  >
                                    Deactivate
                                  </button>
                                )}
                                <button
                                  onClick={() => openConnModal(shard.id)}
                                  className="rounded px-2 py-1 text-xs font-medium text-brand-400 hover:bg-brand-900/30 transition-colors"
                                >
                                  Connection
                                </button>
                                <button
                                  onClick={() => setDeleteTarget(shard.id)}
                                  className="rounded px-2 py-1 text-xs font-medium text-red-400 hover:bg-red-900/30 transition-colors"
                                >
                                  Delete
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            )}

            {activeTab === 'schema' && (
              <div className="p-6">
                <SchemaEditor projectId={project.id} />
              </div>
            )}
          </>
        )}
      </div>

      {/* Connection Modal */}
      {connTarget && (
        <Modal
          title="Shard Connection"
          onClose={() => setConnTarget(null)}
          footer={
            <>
              <button onClick={() => setConnTarget(null)} className="rounded-lg px-3 py-2 text-sm text-gray-400 hover:text-gray-200 transition-colors">
                Cancel
              </button>
              <button
                onClick={handleSaveConn}
                disabled={saving}
                className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white hover:bg-brand-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                {saving ? 'Saving…' : 'Save'}
              </button>
            </>
          }
        >
          <div className="flex flex-col gap-4">
            {(['host', 'port', 'database_name', 'username', 'password'] as const).map((field) => (
              <div key={field}>
                <label className="block text-xs font-medium text-gray-400 mb-1 capitalize">
                  {field.replace('_', ' ')}
                </label>
                <input
                  type={field === 'password' ? 'password' : 'text'}
                  value={form[field as keyof ConnectionForm]}
                  onChange={(e) => setForm((f) => ({ ...f, [field]: e.target.value }))}
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-600
                    focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500 transition-colors"
                />
              </div>
            ))}
          </div>
        </Modal>
      )}

      {/* Delete Shard Modal */}
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
                onClick={handleDeleteShard}
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

      {/* Delete Project Modal */}
      {deleteProjectTarget && (
        <Modal
          title="Delete project"
          onClose={() => setDeleteProjectTarget(null)}
          footer={
            <>
              <button onClick={() => setDeleteProjectTarget(null)} className="rounded-lg px-3 py-2 text-sm text-gray-400 hover:text-gray-200 transition-colors">
                Cancel
              </button>
              <button
                onClick={handleDeleteProject}
                className="rounded-lg bg-red-700 px-4 py-2 text-sm font-medium text-white hover:bg-red-600 transition-colors"
              >
                Delete
              </button>
            </>
          }
        >
          <p className="text-sm text-gray-300">
            Delete project <span className="font-medium text-gray-100">"{deleteProjectTarget.name}"</span>? This cannot be undone.
          </p>
        </Modal>
      )}
    </div>
  )
}
