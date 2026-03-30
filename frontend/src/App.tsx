import { useEffect } from 'react'
import { Sidebar } from './components/Sidebar'
import { ProjectsPage } from './pages/ProjectsPage'
import { ProjectDetailPage } from './pages/ProjectDetailPage'
import { SchemaPage } from './pages/SchemaPage'
import { QueryPage } from './pages/QueryPage'
import { Terminal } from './components/Terminal'
import { useProjectStore } from './stores/projectStore'
import { useAppStore } from './stores/appStore'
import { useProjectActions } from './hooks/useProjectActions'
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
    setCurrentView('project-detail')
  }

  const handleBackToProjects = () => {
    setSelectedProject(null)
    setCurrentView('projects')
  }

  const handleCreate = async (name: string, description: string) => {
    try {
      const p = await createProject(name, description)
      setSelectedProject(p)
      setCurrentView('project-detail')
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
        {currentView === 'project-detail' && selectedProject && (
          <ProjectDetailPage
            project={selectedProject}
            onBack={handleBackToProjects}
          />
        )}
        {currentView === 'schema' && selectedProject && (
          <SchemaPage
            project={selectedProject}
            onNoProject={handleBackToProjects}
          />
        )}
        {currentView === 'query' && selectedProject && (
          <QueryPage
            project={selectedProject}
            onNoProject={handleBackToProjects}
          />
        )}
      </main>

      <Terminal />
    </div>
  )
}
