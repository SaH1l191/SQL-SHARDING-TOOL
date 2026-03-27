import { useCallback } from 'react'
import { useProjectStore } from '../stores/projectStore'
import { useAppStore } from '../stores/appStore'
import { GetProjects, CreateProject, DeleteProject, ActivateProject } from '../../wailsjs/go/main/App'

export function useProjectActions() {
  const { 
    setProjects, 
    addProject, 
    removeProject, 
    updateProject,
    setLoading, 
    setError, 
    clearError 
  } = useProjectStore()
  
  const { addNotification } = useAppStore()

  const fetchProjects = useCallback(async () => {
    setLoading(true)
    clearError()
    try {
      const projects = await (GetProjects as unknown as () => Promise<any[]>)()
      setProjects(projects || [])
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to fetch projects: ${errorMessage}`
      })
    } finally {
      setLoading(false)
    }
  }, [setProjects, setLoading, setError, clearError, addNotification])

  const createProject = useCallback(async (name: string, description: string) => {
    setLoading(true)
    clearError()
    try {
      const project = await (CreateProject as unknown as (n: string, d: string) => Promise<any>)(name, description)
      addProject(project)
      addNotification({
        type: 'success',
        message: `Project "${name}" created successfully`
      })
      return project
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to create project: ${errorMessage}`
      })
      throw error
    } finally {
      setLoading(false)
    }
  }, [addProject, setLoading, setError, clearError, addNotification])

  const deleteProject = useCallback(async (id: string) => {
    setLoading(true)
    clearError()
    try {
      await (DeleteProject as unknown as (id: string) => Promise<void>)(id)
      removeProject(id)
      addNotification({
        type: 'success',
        message: 'Project deleted successfully'
      })
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to delete project: ${errorMessage}`
      })
      throw error
    } finally {
      setLoading(false)
    }
  }, [removeProject, setLoading, setError, clearError, addNotification])

  const activateProject = useCallback(async (id: string) => {
    setLoading(true)
    clearError()
    try {
      await (ActivateProject as unknown as (id: string) => Promise<void>)(id)
      // Refetch projects to get updated status
      await fetchProjects()
      addNotification({
        type: 'success',
        message: 'Project activated successfully'
      })
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to activate project: ${errorMessage}`
      })
      throw error
    } finally {
      setLoading(false)
    }
  }, [fetchProjects, setLoading, setError, clearError, addNotification])

  return {
    fetchProjects,
    createProject,
    deleteProject,
    activateProject
  }
}
