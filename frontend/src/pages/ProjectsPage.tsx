import { useState } from 'react'
import { Modal } from '../components/Modal'
import { StatusBadge } from '../components/StatusBadge'
import { EmptyState } from '../components/EmptyState'
import type { Project } from '../types'

interface Props {
  projects: Project[]
  loading: boolean
  error: string | null
  onSelect: (p: Project) => void
  onCreate: (name: string, description: string) => Promise<void>
  onDelete: (id: string) => Promise<void>
  onActivate: (id: string) => Promise<void>
}

export function ProjectsPage({ projects, loading, error, onSelect, onCreate, onDelete, onActivate }: Props) {
  const [showCreate, setShowCreate] = useState(false)
  const [name, setName] = useState('')
  const [desc, setDesc] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Project | null>(null)

  const handleCreate = async () => {
    if (!name.trim() || !desc.trim()) return
    setSubmitting(true)
    try {
      await onCreate(name.trim(), desc.trim())
      setName(''); setDesc('')
      setShowCreate(false)
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    await onDelete(deleteTarget.id)
    setDeleteTarget(null)
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-gray-800 px-6 py-4">
        <div>
          <h1 className="text-base font-semibold text-gray-100">Projects</h1>
          <p className="text-xs text-gray-500 mt-0.5">{projects.length} project{projects.length !== 1 ? 's' : ''}</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3 py-2 text-sm font-medium text-white hover:bg-brand-700 transition-colors"
        >
          <span className="text-base leading-none">+</span> New project
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto scrollbar-thin px-6 py-4">
        {loading && (
          <div className="flex items-center justify-center py-20">
            <div className="h-6 w-6 rounded-full border-2 border-brand-500 border-t-transparent animate-spin" />
          </div>
        )}
        {error && (
          <div className="rounded-lg border border-red-800 bg-red-900/20 px-4 py-3 text-sm text-red-400">{error}</div>
        )}
        {!loading && !error && projects.length === 0 && (
          <EmptyState
            icon="▦"
            title="No projects yet"
            description="Create a project to start configuring your sharded database."
            action={{ label: 'New project', onClick: () => setShowCreate(true) }}
          />
        )}
        {!loading && projects.length > 0 && (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
            {projects.map(p => (
              <div
                key={p.id}
                className="group relative rounded-xl border border-gray-800 bg-gray-900 p-4 hover:border-gray-700 transition-colors cursor-pointer"
                onClick={() => onSelect(p)}
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-100 truncate">{p.name}</p>
                    <p className="text-xs text-gray-500 mt-0.5 truncate">{p.description}</p>
                  </div>
                  <StatusBadge status={p.status} />
                </div>

                <div className="mt-3 flex items-center gap-3 text-xs text-gray-600">
                  <span>{p.shard_count} shard{p.shard_count !== 1 ? 's' : ''}</span>
                  <span>·</span>
                  <span>{new Date(p.created_at).toLocaleDateString()}</span>
                </div>

                {/* Actions — visible on hover */}
                <div
                  className="absolute inset-x-0 bottom-0 flex items-center justify-end gap-1 rounded-b-xl bg-gray-900/95 px-3 py-2
                    opacity-0 group-hover:opacity-100 transition-opacity border-t border-gray-800"
                  onClick={e => e.stopPropagation()}
                >
                  {p.status === 'inactive' && (
                    <button
                      onClick={() => onActivate(p.id)}
                      className="rounded px-2 py-1 text-[11px] font-medium text-emerald-400 hover:bg-emerald-900/30 transition-colors"
                    >
                      Activate
                    </button>
                  )}
                  <button
                    onClick={() => setDeleteTarget(p)}
                    className="rounded px-2 py-1 text-[11px] font-medium text-red-400 hover:bg-red-900/30 transition-colors"
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create modal */}
      {showCreate && (
        <Modal
          title="New project"
          onClose={() => setShowCreate(false)}
          footer={
            <>
              <button onClick={() => setShowCreate(false)} className="rounded-lg px-3 py-2 text-sm text-gray-400 hover:text-gray-200 transition-colors">
                Cancel
              </button>
              <button
                onClick={handleCreate}
                disabled={submitting || !name.trim() || !desc.trim()}
                className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white hover:bg-brand-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                {submitting ? 'Creating…' : 'Create'}
              </button>
            </>
          }
        >
          <div className="flex flex-col gap-3">
            <div>
              <label className="block text-xs font-medium text-gray-400 mb-1">Name</label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="e-commerce-db"
                className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-600
                  focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500 transition-colors"
                autoFocus
                onKeyDown={e => e.key === 'Enter' && handleCreate()}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-400 mb-1">Description</label>
              <input
                type="text"
                value={desc}
                onChange={e => setDesc(e.target.value)}
                placeholder="Main product database"
                className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-600
                  focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500 transition-colors"
                onKeyDown={e => e.key === 'Enter' && handleCreate()}
              />
            </div>
          </div>
        </Modal>
      )}

      {/* Delete confirm */}
      {deleteTarget && (
        <Modal
          title="Delete project"
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
            Delete <span className="font-medium text-gray-100">"{deleteTarget.name}"</span>? This cannot be undone.
          </p>
        </Modal>
      )}
    </div>
  )
}
