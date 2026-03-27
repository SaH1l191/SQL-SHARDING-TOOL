import { useEffect } from 'react'
import { useAppStore } from '../stores/appStore'

export function Notifications() {
  const { notifications, removeNotification } = useAppStore()

  const getNotificationColor = (type: string) => {
    switch (type) {
      case 'success':
        return 'bg-green-600 border-green-700'
      case 'error':
        return 'bg-red-600 border-red-700'
      case 'warning':
        return 'bg-yellow-600 border-yellow-700'
      case 'info':
        return 'bg-blue-600 border-blue-700'
      default:
        return 'bg-gray-600 border-gray-700'
    }
  }

  return (
    <div className="fixed top-4 right-4 z-50 space-y-2">
      {notifications.map((notification) => (
        <div
          key={notification.id}
          className={`max-w-sm p-4 border rounded-lg shadow-lg text-white ${getNotificationColor(
            notification.type
          )} transition-all duration-300 ease-in-out transform hover:scale-105`}
          onClick={() => removeNotification(notification.id)}
        >
          <div className="flex items-start">
            <div className="flex-1">
              <p className="text-sm font-medium">{notification.message}</p>
            </div>
            <button
              onClick={(e) => {
                e.stopPropagation()
                removeNotification(notification.id)
              }}
              className="ml-4 text-white hover:text-gray-200 focus:outline-none"
            >
              ×
            </button>
          </div>
        </div>
      ))}
    </div>
  )
}
