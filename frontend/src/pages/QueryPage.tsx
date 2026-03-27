import { useState } from 'react'
import { EmptyState } from '../components/EmptyState'
import type { Project } from '../types'

interface RoutingPlan {
  mode: string
  targets: { ShardID: string }[]
  reason: string
}

interface QueryResult {
  routing: RoutingPlan
  results: {
    shard_id: string
    columns: string[]
    rows: unknown[][]
    rows_affected: number
    error?: string
  }[]
}

interface Props {
  project: Project | null
  onNoProject: () => void
}

export function QueryPage({ project, onNoProject }: Props) {
  const [sql, setSql] = useState('')
  const [result, setResult] = useState<QueryResult | null>(null)
  const [status, setStatus] = useState<'idle' | 'loading' | 'ok' | 'error'>('idle')
  const [errMsg, setErrMsg] = useState('')

  if (!project) {
    return (
      <div className="flex items-center justify-center h-full">
        <EmptyState
          icon="⌘"
          title="No project selected"
          description="Select a project to run queries."
          action={{ label: 'Go to projects', onClick: onNoProject }}
        />
      </div>
    )
  }

  const handleRun = async () => {
    if (!sql.trim()) return
    setStatus('loading')
    setResult(null)
    setErrMsg('')
    try {
      const res = await fetch('http://localhost:8080/api/v1/query/execute', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ project_id: project.id, sql: sql.trim() }),
      })
      const body = await res.json()
      if (!res.ok) throw new Error(body.error ?? 'Request failed')
      setResult(body)
      setStatus('ok')
    } catch (e) {
      setStatus('error')
      setErrMsg(String(e))
    }
  }

  const modeColor: Record<string, string> = {
    single:   'bg-emerald-900/40 text-emerald-400',
    multi:    'bg-amber-900/40 text-amber-400',
    rejected: 'bg-red-900/40 text-red-400',
    broadcast:'bg-blue-900/40 text-blue-400',
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-gray-800 px-6 py-4">
        <div>
          <h1 className="text-base font-semibold text-gray-100">Query</h1>
          <p className="text-xs text-gray-500 mt-0.5">{project.name}</p>
        </div>
        <button
          onClick={handleRun}
          disabled={status === 'loading' || !sql.trim()}
          className="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3 py-2 text-sm font-medium text-white hover:bg-brand-700 disabled:opacity-40 transition-colors"
        >
          {status === 'loading' ? (
            <>
              <span className="h-3 w-3 rounded-full border border-white border-t-transparent animate-spin" />
              Running…
            </>
          ) : (
            <><span className="font-mono text-xs opacity-70">▶</span> Run</>
          )}
        </button>
      </div>

      <div className="flex flex-col flex-1 overflow-hidden">
        {/* SQL editor */}
        <div className="relative border-b border-gray-800" style={{ height: '180px' }}>
          <textarea
            value={sql}
            onChange={e => setSql(e.target.value)}
            placeholder="SELECT * FROM users WHERE id = 'abc';"
            spellCheck={false}
            onKeyDown={e => {
              if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') { e.preventDefault(); handleRun() }
            }}
            className="h-full w-full resize-none bg-gray-950 px-4 py-3 font-mono text-xs text-gray-300 placeholder-gray-700
              focus:outline-none scrollbar-thin leading-relaxed"
          />
          <span className="absolute bottom-2 right-3 text-[10px] text-gray-700 select-none">⌘↵ to run</span>
        </div>

        {/* Results area */}
        <div className="flex-1 overflow-y-auto scrollbar-thin px-6 py-4 space-y-4">
          {status === 'error' && (
            <div className="rounded-lg border border-red-800 bg-red-900/20 px-4 py-3 text-sm text-red-400">{errMsg}</div>
          )}

          {result && (
            <>
              {/* Routing plan */}
              <div className="rounded-xl border border-gray-800 bg-gray-900 overflow-hidden">
                <div className="flex items-center gap-3 px-4 py-2.5 border-b border-gray-800">
                  <span className="text-xs font-medium text-gray-400">Routing plan</span>
                  <span className={`text-[11px] rounded-full px-2 py-0.5 font-medium ${modeColor[result.routing.mode] ?? modeColor.rejected}`}>
                    {result.routing.mode}
                  </span>
                </div>
                <div className="px-4 py-3 text-xs text-gray-500 space-y-1">
                  <p>{result.routing.reason}</p>
                  {result.routing.targets?.length > 0 && (
                    <p className="font-mono text-gray-600">
                      Targets: {result.routing.targets.map(t => t.ShardID.slice(0,8)).join(', ')}
                    </p>
                  )}
                </div>
              </div>

              {/* Per-shard results */}
              {result.results?.map((r, i) => (
                <div key={i} className="rounded-xl border border-gray-800 bg-gray-900 overflow-hidden">
                  <div className="flex items-center gap-2 px-4 py-2.5 border-b border-gray-800">
                    <span className="text-xs font-medium text-gray-400">Shard</span>
                    <span className="font-mono text-xs text-gray-500">{r.shard_id?.slice(0, 8)}…</span>
                    {r.error && <span className="ml-auto text-xs text-red-400">{r.error}</span>}
                    {r.rows_affected > 0 && (
                      <span className="ml-auto text-xs text-emerald-400">{r.rows_affected} row{r.rows_affected !== 1 ? 's' : ''} affected</span>
                    )}
                  </div>

                  {r.columns?.length > 0 && (
                    <div className="overflow-x-auto scrollbar-thin">
                      <table className="w-full text-xs">
                        <thead>
                          <tr className="border-b border-gray-800">
                            {r.columns.map(col => (
                              <th key={col} className="px-4 py-2 text-left font-medium text-gray-500 whitespace-nowrap">{col}</th>
                            ))}
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-800/50">
                          {r.rows?.map((row, ri) => (
                            <tr key={ri} className="hover:bg-gray-800/30 transition-colors">
                              {row.map((cell, ci) => (
                                <td key={ci} className="px-4 py-2 font-mono text-gray-400 whitespace-nowrap">
                                  {cell === null ? <span className="text-gray-600 italic">null</span> : String(cell)}
                                </td>
                              ))}
                            </tr>
                          ))}
                          {(!r.rows || r.rows.length === 0) && (
                            <tr><td colSpan={r.columns.length} className="px-4 py-4 text-center text-gray-600">No rows returned</td></tr>
                          )}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>
              ))}
            </>
          )}

          {status === 'idle' && (
            <div className="flex items-center justify-center py-12 text-xs text-gray-700">
              Run a query to see results
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
