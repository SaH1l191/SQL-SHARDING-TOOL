import { useState, useEffect, useCallback } from 'react'
import { useSchemaStore } from '../stores/schemaStore'
import { 
  CreateSchemaDraft,
  CommitSchemaDraft,
  DeleteSchemaDraft,
  GetCurrentSchema,
  ExecuteProjectSchema,
  RetrySchemaExecution,
  GetSchemaCapabilities,
  UpdateProjectSchemaDraft,
  GetSchemaHistory,
} from '../../wailsjs/go/main/App'

interface SchemaEditorProps {
  projectId: string
}

export function SchemaEditor({ projectId }: SchemaEditorProps) {
  const [ddl, setDdl] = useState('')
  const [openDdlId, setOpenDdlId] = useState<string | null>(null)
  
  const { 
    currentSchema, 
    schemaHistory,
    capabilities, 
    loading, 
    saving, 
    executing,
    setCurrentSchema,
    setSchemaHistory,
    setCapabilities,
    setLoading,
    setSaving,
    setExecuting,
    setError,
    clearError
  } = useSchemaStore()

  const refresh = useCallback(async () => {
    setLoading(true)
    clearError()
    try {
      const [schemaRes, capsRes, historyRes] = await Promise.all([
        GetCurrentSchema(projectId).catch(() => null),
        GetSchemaCapabilities(projectId),
        GetSchemaHistory(projectId).catch(() => []),
      ])

      setCurrentSchema(schemaRes)
      setCapabilities(capsRes)
      setSchemaHistory(historyRes || [])
      setDdl(schemaRes?.ddl_sql ?? '')
    } catch (error) {
      setError(String(error))
    } finally {
      setLoading(false)
    }
  }, [projectId, setCurrentSchema, setCapabilities, setSchemaHistory, setLoading, setError, clearError])

  useEffect(() => {
    refresh()
  }, [refresh])

  const handleCreateDraft = async () => {
    setSaving(true)
    clearError()
    try {
      await CreateSchemaDraft(projectId, '')
      await refresh()
    } catch (error) {
      setError(String(error))
    } finally {
      setSaving(false)
    }
  }

  const handleSaveDraft = async () => {
    if (!ddl.trim() || !currentSchema) return

    setSaving(true)
    clearError()
    try {
      await UpdateProjectSchemaDraft(projectId, currentSchema.id, ddl)
      // Immediately update local history with new DDL
      const updatedSchema = { ...currentSchema, ddl_sql: ddl }
      setCurrentSchema(updatedSchema)
      // Update in history if exists
      const existingIndex = schemaHistory.findIndex(s => s.id === currentSchema.id)
      if (existingIndex >= 0) {
        const newHistory = [...schemaHistory]
        newHistory[existingIndex] = updatedSchema
        setSchemaHistory(newHistory)
      }
      await refresh()
    } catch (error) {
      setError(String(error))
    } finally {
      setSaving(false)
    }
  }

  const handleCommit = async () => {
    if (!currentSchema) return
    if (!ddl.trim()) {
      setError('Cannot commit: DDL is empty')
      return
    }

    setSaving(true)
    clearError()
    try {
      await CommitSchemaDraft(projectId, currentSchema.id)
      await refresh()
    } catch (error) {
      setError(String(error))
    } finally {
      setSaving(false)
    }
  }

  const handleDiscard = async () => {
    if (!currentSchema) return

    setSaving(true)
    clearError()
    try {
      await DeleteSchemaDraft(currentSchema.id)
      await refresh()
    } catch (error) {
      setError(String(error))
    } finally {
      setSaving(false)
    }
  }

  const handleExecute = async () => {
    if (!currentSchema) return
    if (!ddl.trim()) {
      setError('Cannot execute: DDL is empty')
      return
    }

    setExecuting(true)
    clearError()
    try {
      await ExecuteProjectSchema(projectId)
      await refresh()
    } catch (error) {
      setError(String(error))
    } finally {
      setExecuting(false)
    }
  }

  const handleRetry = async () => {
    setExecuting(true)
    clearError()
    try {
      await RetrySchemaExecution(projectId)
      await refresh()
    } catch (error) {
      setError(String(error))
    } finally {
      setExecuting(false)
    }
  }

  const formatDate = (date?: string) => {
    if (!date) return '—'
    return new Date(date).toLocaleString()
  }

  const getStateBadge = (state: string) => {
    switch (state) {
      case 'applied':
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-900/50 text-emerald-400">Applied</span>
      case 'committed':
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-900/50 text-blue-400">Committed</span>
      case 'failed':
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-900/50 text-red-400">Failed</span>
      case 'applying':
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-900/50 text-amber-400">Applying</span>
      default:
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-800 text-gray-400">{state}</span>
    }
  }

  if (loading && !currentSchema && !capabilities) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="h-8 w-8 rounded-full border-2 border-brand-500 border-t-transparent animate-spin" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* DDL Editor Card */}
      <div className="rounded-lg border border-gray-800 bg-gray-900">
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-800">
          <h3 className="text-sm font-semibold text-gray-100">Schema DDL</h3>
          {currentSchema && getStateBadge(currentSchema.state)}
        </div>
        
        <div className="p-4 space-y-4">
          <textarea
            value={ddl}
            onChange={(e) => setDdl(e.target.value)}
            placeholder="-- Enter your DDL statements here...\nCREATE TABLE users (\n  id SERIAL PRIMARY KEY,\n  name TEXT,\n  email TEXT UNIQUE\n);"
            className="w-full h-48 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 font-mono
              focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500 transition-colors resize-none"
            disabled={!capabilities?.can_edit_draft}
          />

          {capabilities?.reason && (
            <div className="text-sm text-gray-500">
              {capabilities.reason}
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex flex-wrap gap-2">
            {capabilities?.can_create_draft && (
              <button
                onClick={handleCreateDraft}
                disabled={saving}
                className="rounded-lg bg-brand-600 px-3 py-2 text-sm font-medium text-white hover:bg-brand-700 
                  disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                Create New Draft
              </button>
            )}

            {capabilities?.can_edit_draft && (
              <button
                onClick={handleSaveDraft}
                disabled={!ddl.trim() || saving}
                className="rounded-lg bg-gray-700 px-3 py-2 text-sm font-medium text-gray-200 hover:bg-gray-600 
                  disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                Save Draft
              </button>
            )}

            {capabilities?.can_commit && (
              <button
                onClick={handleCommit}
                disabled={saving}
                className="rounded-lg bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700 
                  disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                Commit Schema
              </button>
            )}

            {capabilities?.can_edit_draft && currentSchema && (
              <button
                onClick={handleDiscard}
                disabled={saving}
                className="rounded-lg border border-gray-700 px-3 py-2 text-sm font-medium text-gray-400 hover:text-gray-200 
                  hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                Discard Draft
              </button>
            )}

            {capabilities?.can_execute && (
              <button
                onClick={handleExecute}
                disabled={executing}
                className="rounded-lg bg-emerald-600 px-3 py-2 text-sm font-medium text-white hover:bg-emerald-700 
                  disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                {executing ? 'Executing…' : 'Execute Schema'}
              </button>
            )}

            {capabilities?.can_retry && (
              <button
                onClick={handleRetry}
                disabled={executing}
                className="rounded-lg border border-gray-700 px-3 py-2 text-sm font-medium text-gray-400 hover:text-gray-200 
                  hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                {executing ? 'Retrying…' : 'Retry Execution'}
              </button>
            )}
          </div>

          {/* Status Messages */}
          {currentSchema?.state === 'applying' && (
            <div className="flex items-center gap-2 text-sm text-amber-400">
              <div className="h-4 w-4 rounded-full border-2 border-amber-400 border-t-transparent animate-spin" />
              Schema execution in progress…
            </div>
          )}

          {currentSchema?.state === 'failed' && currentSchema.error_message && (
            <div className="rounded-lg border border-red-800 bg-red-900/20 px-3 py-2 text-sm text-red-400">
              Execution failed: {currentSchema.error_message}
            </div>
          )}

          {currentSchema?.state === 'applied' && (
            <div className="text-sm text-emerald-400">
              Schema applied successfully.
            </div>
          )}
        </div>
      </div>

      {/* Schema History Card */}
      <div className="rounded-lg border border-gray-800 bg-gray-900">
        <div className="px-4 py-3 border-b border-gray-800">
          <h3 className="text-sm font-semibold text-gray-100">Schema History</h3>
        </div>
        
        <div className="p-4">
          {schemaHistory.length === 0 ? (
            <div className="text-sm text-gray-500 text-center py-8">
              No schema history.
            </div>
          ) : (
            <div className="space-y-3">
              {schemaHistory.map((schema) => (
                <div
                  key={schema.id}
                  className="rounded-lg border border-gray-800 bg-gray-800/50 p-3 space-y-2"
                >
                  {/* Header */}
                  <div className="flex items-center justify-between">
                    <div className="text-sm font-medium text-gray-200">
                      Version v{schema.version || 1}
                    </div>
                    {getStateBadge(schema.state)}
                  </div>

                  {/* Timestamps */}
                  <div className="text-xs text-gray-500 space-y-1">
                    <div>Created: {formatDate(schema.created_at)}</div>
                    {schema.commited_at && (
                      <div>Committed: {formatDate(schema.commited_at)}</div>
                    )}
                  </div>

                  {/* Error */}
                  {schema.error_message && (
                    <div className="text-xs text-red-400">
                      {schema.error_message}
                    </div>
                  )}

                  {/* DDL Toggle */}
                  <button
                    onClick={() => setOpenDdlId(openDdlId === schema.id ? null : schema.id)}
                    className="text-xs text-brand-400 hover:text-brand-300 transition-colors"
                  >
                    {openDdlId === schema.id ? 'Hide DDL' : 'View DDL'}
                  </button>

                  {openDdlId === schema.id && (
                    <pre className="bg-gray-950 rounded-lg p-3 text-xs text-gray-300 font-mono overflow-x-auto">
                      {schema.ddl_sql}
                    </pre>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
