import { useEffect } from 'react'
import { Sidebar } from './components/Sidebar'
import { ProjectsPage } from './pages/ProjectsPage'
import { ShardsPage } from './pages/ShardsPage'
import { SchemaPage } from './pages/SchemaPage'
import { QueryPage } from './pages/QueryPage'
import { useProjectStore } from './stores/projectStore'
import { useAppStore } from './stores/appStore'
import { useProjectActions } from './hooks/useProjectActions'
import { Notifications } from './components/Notifications'
import type { Project } from './types'

export default function App() {
  const { projects, selectedProject, setSelectedProject, loading, error } = useProjectStore()
  const { currentView, setCurrentView } = useAppStore()
  const { fetchProjects, createProject, deleteProject, activateProject } = useProjectActions()

  useEffect(() => {
    fetchProjects()
  }, [fetchProjects])

  const handleSelectProject = (project: Project) => {
    setSelectedProject(project)
    setCurrentView('shards')
  }

  const handleCreate = async (name: string, description: string) => {
    try {
      const p = await createProject(name, description)
      setSelectedProject(p)
    } catch (error) {
      // Error is handled by the hook
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteProject(id)
    } catch (error) {
      // Error is handled by the hook
    }
  }

  const handleActivate = async (id: string) => {
    try {
      await activateProject(id)
    } catch (error) {
      // Error is handled by the hook
    }
  }

  return (
    <div className="flex h-screen overflow-hidden bg-gray-950 text-gray-100">
      <Sidebar
        projects={projects}
        selectedProject={selectedProject}
        activeView={currentView}
        onSelectProject={handleSelectProject}
        onNewProject={() => setCurrentView('projects')}
        onViewChange={setCurrentView}
      />

      <main className="flex flex-1 flex-col overflow-hidden">
        {currentView === 'projects' && (
          <ProjectsPage
            projects={projects}
            loading={loading}
            error={error}
            onSelect={handleSelectProject}
            onCreate={handleCreate}
            onDelete={handleDelete}
            onActivate={handleActivate}
          />
        )}
        {currentView === 'shards' && (
          <ShardsPage
            project={selectedProject}
            onNoProject={() => setCurrentView('projects')}
          />
        )}
        {currentView === 'schema' && (
          <SchemaPage
            project={selectedProject}
            onNoProject={() => setCurrentView('projects')}
          />
        )}
        {currentView === 'query' && (
          <QueryPage
            project={selectedProject}
            onNoProject={() => setCurrentView('projects')}
          />
        )}
      </main>

      <Notifications />
    </div>
  )
}
