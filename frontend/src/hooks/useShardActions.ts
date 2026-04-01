import { useCallback } from 'react'
import { useShardStore } from '../stores/shardStore'
import { useAppStore } from '../stores/appStore'
import { 
  GetShards, 
  CreateShard, 
  DeleteShard, 
  ActivateShard, 
  DeactivateShard,
  FetchConnectionInfo,
  UpdateConnection,
  AddConnection
} from '../../wailsjs/go/main/App'

export function useShardActions() {
  const { 
    setShards, 
    addShard, 
    removeShard, 
    updateShard,
    setConnection,
    updateConnection,
    setLoading, 
    setError, 
    clearError 
  } = useShardStore()
  
  const { addNotification } = useAppStore()

  const fetchShards = useCallback(async (projectId: string) => {
    setLoading(true)
    clearError()
    try {
      const shards = await GetShards(projectId)
      setShards((shards || []) as any[])
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      // Error logged to terminal, no popup
      throw error
    } finally {
      setLoading(false)
    }
  }, [setShards, setLoading, setError, clearError, addNotification])

  const createShard = useCallback(async (projectId: string) => {
    setLoading(true)
    clearError()
    try {
      const shard = await CreateShard(projectId)
      addShard(shard as any)
      addNotification({
        type: 'success',
        message: 'Shard created successfully'
      })
      return shard
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      // Error logged to terminal, no popup
      throw error
    } finally {
      setLoading(false)
    }
  }, [addShard, setLoading, setError, clearError, addNotification])

  const deleteShard = useCallback(async (shardId: string, projectId?: string) => {
    setLoading(true)
    clearError()
    try {
      await DeleteShard(shardId)
      removeShard(shardId)
      addNotification({
        type: 'success',
        message: 'Shard deleted successfully'
      })
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      // Error logged to terminal, no popup
      // Re-fetch shards to ensure UI is in sync with backend
      if (projectId) {
        await fetchShards(projectId)
      }
      throw error
    } finally {
      setLoading(false)
    }
  }, [removeShard, fetchShards, setLoading, setError, clearError, addNotification])

  const activateShard = useCallback(async (shardId: string) => {
    setLoading(true)
    clearError()
    try {
      await ActivateShard(shardId)
      updateShard(shardId, { status: 'active' })
      addNotification({
        type: 'success',
        message: 'Shard activated successfully'
      })
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      // Error logged to terminal, no popup
      throw error
    } finally {
      setLoading(false)
    }
  }, [updateShard, setLoading, setError, clearError, addNotification])

  const deactivateShard = useCallback(async (shardId: string) => {
    setLoading(true)
    clearError()
    try {
      await DeactivateShard(shardId)
      updateShard(shardId, { status: 'inactive' })
      addNotification({
        type: 'success',
        message: 'Shard deactivated successfully'
      })
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      // Error logged to terminal, no popup
      throw error
    } finally {
      setLoading(false)
    }
  }, [updateShard, setLoading, setError, clearError, addNotification])

  const fetchShardConnection = useCallback(async (shardId: string) => {
    clearError()
    try {
      const connection = await FetchConnectionInfo(shardId)
      setConnection(shardId, connection as any)
      return connection
    } catch (error) {
      // Don't treat "no connection found" as an error for the UI
      const errorMessage = String(error)
      if (errorMessage.includes('not found') || errorMessage.includes('no rows')) {
        // This is expected when no connection exists yet
        return null
      }
      setError(errorMessage)
      // Error logged to terminal, no popup
      throw error
    }
  }, [setConnection, setError, clearError, addNotification])

  const updateShardConnectionInfo = useCallback(async (connection: any) => {
    setLoading(true)
    clearError()
    try {
      // Check if connection already exists by trying to fetch it
      let existingConnection = null
      try {
        existingConnection = await FetchConnectionInfo(connection.shard_id)
      } catch {
        // No existing connection, that's fine
      }
      
      if (existingConnection && existingConnection.shard_id) {
        // Update existing connection
        await UpdateConnection(connection)
        addNotification({
          type: 'success',
          message: 'Shard connection updated successfully'
        })
      } else {
        // Create new connection
        await AddConnection(connection)
        addNotification({
          type: 'success',
          message: 'Shard connection created successfully'
        })
      }
      
      updateConnection(connection.shard_id, connection)
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      // Error logged to terminal, no popup
      throw error
    } finally {
      setLoading(false)
    }
  }, [updateConnection, setLoading, setError, clearError, addNotification])

  return {
    fetchShards,
    createShard,
    deleteShard,
    activateShard,
    deactivateShard,
    fetchShardConnection,
    updateShardConnectionInfo
  }
}
