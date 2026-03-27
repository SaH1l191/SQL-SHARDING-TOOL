import type { Project, View } from '../types'

interface Props {
  projects: Project[]
  selectedProject: Project | null
  activeView: View
  onSelectProject: (p: Project) => void
  onNewProject: () => void
  onViewChange: (v: View) => void
}

const NAV: { view: View; label: string; icon: string }[] = [
  { view: 'projects', label: 'Projects',  icon: '▦' },
  { view: 'shards',   label: 'Shards',    icon: '⬡' },
  { view: 'schema',   label: 'Schema',    icon: '⊞' },
  { view: 'query',    label: 'Query',     icon: '⌘' },
]

export function Sidebar({ projects, selectedProject, activeView, onSelectProject, onNewProject, onViewChange }: Props) {
  return (
    <aside className="flex h-full w-56 flex-col border-r border-gray-800 bg-gray-950">
      {/* Logo */}
      <div className="flex items-center gap-2 px-4 py-4 border-b border-gray-800">
        <span className="text-brand-500 text-lg font-bold tracking-tight">SQL</span>
        <span className="text-gray-400 text-sm font-medium">Sharder</span>
      </div>

      {/* Nav */}
      <nav className="flex flex-col gap-0.5 px-2 pt-3">
        {NAV.map(({ view, label, icon }) => (
          <button
            key={view}
            onClick={() => onViewChange(view)}
            className={`flex items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors text-left
              ${activeView === view
                ? 'bg-brand-600/20 text-brand-400 font-medium'
                : 'text-gray-400 hover:bg-gray-800 hover:text-gray-200'
              }`}
          >
            <span className="w-4 text-center opacity-70">{icon}</span>
            {label}
          </button>
        ))}
      </nav>

      {/* Project list */}
      <div className="mt-4 flex flex-col flex-1 overflow-hidden">
        <div className="flex items-center justify-between px-4 py-1.5">
          <span className="text-xs font-semibold uppercase tracking-wider text-gray-500">Projects</span>
          <button
            onClick={onNewProject}
            className="text-gray-500 hover:text-brand-400 transition-colors text-base leading-none"
            title="New project"
          >
            +
          </button>
        </div>
        <ul className="flex-1 overflow-y-auto scrollbar-thin px-2 pb-2">
          {projects.map(p => (
            <li key={p.id}>
              <button
                onClick={() => onSelectProject(p)}
                className={`w-full text-left rounded-md px-3 py-2 text-sm transition-colors
                  ${selectedProject?.id === p.id
                    ? 'bg-gray-800 text-gray-100'
                    : 'text-gray-400 hover:bg-gray-800/60 hover:text-gray-200'
                  }`}
              >
                <div className="flex items-center justify-between gap-1">
                  <span className="truncate">{p.name}</span>
                  <span className={`shrink-0 text-[10px] rounded px-1.5 py-0.5 font-medium
                    ${p.status === 'active' ? 'bg-emerald-900/50 text-emerald-400' : 'bg-gray-800 text-gray-500'}`}>
                    {p.status}
                  </span>
                </div>
                <div className="text-[11px] text-gray-600 mt-0.5">{p.shard_count} shard{p.shard_count !== 1 ? 's' : ''}</div>
              </button>
            </li>
          ))}
          {projects.length === 0 && (
            <li className="px-3 py-4 text-xs text-gray-600 text-center">No projects yet</li>
          )}
        </ul>
      </div>
    </aside>
  )
}
