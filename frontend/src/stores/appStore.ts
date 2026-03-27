import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import type { View } from '../types'

interface AppState {
  currentView: View
  sidebarOpen: boolean
  notifications: Array<{
    id: string
    type: 'success' | 'error' | 'warning' | 'info'
    message: string
    timestamp: number
  }>
  
  // Actions
  setCurrentView: (view: View) => void
  setSidebarOpen: (open: boolean) => void
  addNotification: (notification: Omit<Notification, 'id' | 'timestamp'>) => void
  removeNotification: (id: string) => void
  clearNotifications: () => void
}

interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  message: string
  timestamp: number
}

export const useAppStore = create<AppState>()(
  devtools(
    (set, get) => ({
      currentView: 'projects',
      sidebarOpen: true,
      notifications: [],

      setCurrentView: (view) => set({ currentView: view }),
      
      setSidebarOpen: (open) => set({ sidebarOpen: open }),
      
      addNotification: (notification) => {
        const id = Date.now().toString()
        const timestamp = Date.now()
        set((state) => ({
          notifications: [...state.notifications, { ...notification, id, timestamp }]
        }))
        
        // Auto-remove notification after 5 seconds
        setTimeout(() => {
          get().removeNotification(id)
        }, 5000)
      },
      
      removeNotification: (id) => set((state) => ({
        notifications: state.notifications.filter(n => n.id !== id)
      })),
      
      clearNotifications: () => set({ notifications: [] })
    }),
    {
      name: 'app-store'
    }
  )
)
