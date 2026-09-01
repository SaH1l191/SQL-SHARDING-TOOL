import { useState } from 'react'
import { Sidebar } from './components/Sidebar'
import { ProjectsPage } from './pages/ProjectsPage'
import { ShardsPage } from './pages/ShardsPage'
import { SchemaPage } from './pages/SchemaPage'
import { QueryPage } from './pages/QueryPage'
import { useProjects } from './hooks/useProjects'
import type { Project, View } from './types'

export default function App() {
  const { projects, loading, error, create, remove, activate } = useProjects()
  const [selectedProject, setSelectedProject] = useState<Project | null>(null)
  const [view, setView] = useState<View>('projects')

  const handleSelectProject = (p: Project) => {
    setSelectedProject(p)
    setView('shards')
  }

  const handleCreate = async (name: string, description: string) => {
    const p = await create(name, description)
    setSelectedProject(p)
  }

  return (
    <div className="flex h-screen overflow-hidden bg-gray-950 text-gray-100">
      <Sidebar
        projects={projects}
        selectedProject={selectedProject}
        activeView={view}
        onSelectProject={handleSelectProject}
        onNewProject={() => setView('projects')}
        onViewChange={setView}
      />

      <main className="flex flex-1 flex-col overflow-hidden">
        {view === 'projects' && (
          <ProjectsPage
            projects={projects}
            loading={loading}
            error={error}
            onSelect={handleSelectProject}
            onCreate={handleCreate}
            onDelete={remove}
            onActivate={activate}
          />
        )}
        {view === 'shards' && (
          <ShardsPage
            project={selectedProject}
            onNoProject={() => setView('projects')}
          />
        )}
        {view === 'schema' && (
          <SchemaPage
            project={selectedProject}
            onNoProject={() => setView('projects')}
          />
        )}
        {view === 'query' && (
          <QueryPage
            project={selectedProject}
            onNoProject={() => setView('projects')}
          />
        )}
      </main>
    </div>
  )
}
