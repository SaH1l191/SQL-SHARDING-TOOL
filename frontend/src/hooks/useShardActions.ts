import { useCallback } from 'react'
import { useShardStore } from '../stores/shardStore'
import { useAppStore } from '../stores/appStore'
import { GetShards, CreateShard, DeleteShard, ActivateShard, DeactivateShard, GetShardConnection, UpdateShardConnection } from '../../wailsjs/go/main/App'

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
      const shards = await (GetShards as unknown as (id: string) => Promise<any[]>)(projectId)
      setShards(shards || [])
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to fetch shards: ${errorMessage}`
      })
    } finally {
      setLoading(false)
    }
  }, [setShards, setLoading, setError, clearError, addNotification])

  const createShard = useCallback(async (projectId: string) => {
    setLoading(true)
    clearError()
    try {
      const shard = await (CreateShard as unknown as (id: string) => Promise<any>)(projectId)
      addShard(shard)
      addNotification({
        type: 'success',
        message: 'Shard created successfully'
      })
      return shard
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to create shard: ${errorMessage}`
      })
      throw error
    } finally {
      setLoading(false)
    }
  }, [addShard, setLoading, setError, clearError, addNotification])

  const deleteShard = useCallback(async (shardId: string) => {
    setLoading(true)
    clearError()
    try {
      await (DeleteShard as unknown as (id: string) => Promise<void>)(shardId)
      removeShard(shardId)
      addNotification({
        type: 'success',
        message: 'Shard deleted successfully'
      })
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to delete shard: ${errorMessage}`
      })
      throw error
    } finally {
      setLoading(false)
    }
  }, [removeShard, setLoading, setError, clearError, addNotification])

  const activateShard = useCallback(async (shardId: string) => {
    setLoading(true)
    clearError()
    try {
      await (ActivateShard as unknown as (id: string) => Promise<void>)(shardId)
      updateShard(shardId, { status: 'active' })
      addNotification({
        type: 'success',
        message: 'Shard activated successfully'
      })
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to activate shard: ${errorMessage}`
      })
      throw error
    } finally {
      setLoading(false)
    }
  }, [updateShard, setLoading, setError, clearError, addNotification])

  const deactivateShard = useCallback(async (shardId: string) => {
    setLoading(true)
    clearError()
    try {
      await (DeactivateShard as unknown as (id: string) => Promise<void>)(shardId)
      updateShard(shardId, { status: 'inactive' })
      addNotification({
        type: 'success',
        message: 'Shard deactivated successfully'
      })
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to deactivate shard: ${errorMessage}`
      })
      throw error
    } finally {
      setLoading(false)
    }
  }, [updateShard, setLoading, setError, clearError, addNotification])

  const fetchShardConnection = useCallback(async (shardId: string) => {
    clearError()
    try {
      const connection = await (GetShardConnection as unknown as (id: string) => Promise<any>)(shardId)
      setConnection(shardId, connection)
      return connection
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to fetch shard connection: ${errorMessage}`
      })
      throw error
    }
  }, [setConnection, setError, clearError, addNotification])

  const updateShardConnectionInfo = useCallback(async (connection: any) => {
    setLoading(true)
    clearError()
    try {
      await (UpdateShardConnection as unknown as (conn: any) => Promise<void>)(connection)
      updateConnection(connection.shard_id, connection)
      addNotification({
        type: 'success',
        message: 'Shard connection updated successfully'
      })
    } catch (error) {
      const errorMessage = String(error)
      setError(errorMessage)
      addNotification({
        type: 'error',
        message: `Failed to update shard connection: ${errorMessage}`
      })
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
