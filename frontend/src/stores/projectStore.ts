import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import type { Project } from '../types'

interface ProjectState {
  projects: Project[]
  selectedProject: Project | null
  loading: boolean
  error: string | null
  
  // Actions
  setProjects: (projects: Project[]) => void
  setSelectedProject: (project: Project | null) => void
  addProject: (project: Project) => void
  updateProject: (id: string, updates: Partial<Project>) => void
  removeProject: (id: string) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
}

export const useProjectStore = create<ProjectState>()(
  devtools(
    (set, get) => ({
      projects: [],
      selectedProject: null,
      loading: false,
      error: null,

      setProjects: (projects) => set({ projects }),
      
      setSelectedProject: (project) => set({ selectedProject: project }),
      
      addProject: (project) => set((state) => ({
        projects: [project, ...state.projects]
      })),
      
      updateProject: (id, updates) => set((state) => ({
        projects: state.projects.map(p => 
          p.id === id ? { ...p, ...updates } : p
        ),
        selectedProject: state.selectedProject?.id === id 
          ? { ...state.selectedProject, ...updates } 
          : state.selectedProject
      })),
      
      removeProject: (id) => set((state) => ({
        projects: state.projects.filter(p => p.id !== id),
        selectedProject: state.selectedProject?.id === id ? null : state.selectedProject
      })),
      
      setLoading: (loading) => set({ loading }),
      
      setError: (error) => set({ error }),
      
      clearError: () => set({ error: null })
    }),
    {
      name: 'project-store'
    }
  )
)
