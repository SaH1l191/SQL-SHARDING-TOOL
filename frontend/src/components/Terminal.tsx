// import { useEffect, useRef, useState } from 'react'
// import { EventsOn } from '../../wailsjs/runtime/runtime'
// import { useAppStore } from '../stores/appStore'

// interface LogEvent {
//   level: 'info' | 'warn' | 'error'
//   message: string
//   source: string
//   timestamp: string
//   fields?: Record<string, string>
// }

// export function Terminal() {
//   const [logs, setLogs] = useState<LogEvent[]>([])
//   const [isVisible, setIsVisible] = useState(false)
//   const terminalRef = useRef<HTMLDivElement>(null)
//   const { addNotification } = useAppStore()

//   useEffect(() => {
//     // Listen for log events from backend
//     const unsubscribe = EventsOn("log:", (event: LogEvent) => {
//       setLogs(prev => [...prev.slice(-199), event]) // Keep last 200 logs

//       // Still show critical errors as notifications
//       if (event.level === 'error') {
//         addNotification({
//           type: 'error',
//           message: `${event.source}: ${event.message}`
//         })
//       }
//     })

//     return () => {
//       unsubscribe()
//     }
//   }, [addNotification])

//   useEffect(() => {
//     // Auto-scroll to bottom when new logs arrive
//     if (terminalRef.current) {
//       terminalRef.current.scrollTop = terminalRef.current.scrollHeight
//     }
//   }, [logs])

//   const getLogLevelColor = (level: string) => {
//     switch (level) {
//       case 'error':
//         return 'text-red-400'
//       case 'warn':
//         return 'text-yellow-400'
//       case 'info':
//         return 'text-blue-400'
//       default:
//         return 'text-gray-300'
//     }
//   }

//   const getLogLevelBg = (level: string) => {
//     switch (level) {
//       case 'error':
//         return 'bg-red-900/20 border-red-800/30'
//       case 'warn':
//         return 'bg-yellow-900/20 border-yellow-800/30'
//       case 'info':
//         return 'bg-blue-900/20 border-blue-800/30'
//       default:
//         return 'bg-gray-900/20 border-gray-800/30'
//     }
//   }

//   const clearLogs = () => {
//     setLogs([])
//   }

//   const exportLogs = () => {
//     const logText = logs.map(log => 
//       `[${log.timestamp}] [${log.level.toUpperCase()}] ${log.source}: ${log.message}`
//     ).join('\n')

//     const blob = new Blob([logText], { type: 'text/plain' })
//     const url = URL.createObjectURL(blob)
//     const a = document.createElement('a')
//     a.href = url
//     a.download = `sql-sharder-logs-${new Date().toISOString().split('T')[0]}.log`
//     a.click()
//     URL.revokeObjectURL(url)
//   }

//   if (!isVisible) {
//     return (
//       <button
//         onClick={() => setIsVisible(true)}
//         className="fixed bottom-4 right-4 z-40 px-4 py-2 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-lg shadow-lg border border-gray-700 transition-all duration-200 flex items-center gap-2"
//       >
//         <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
//           <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
//         </svg>
//         Terminal ({logs.length})
//       </button>
//     )
//   }

//   return (
//     <div className="fixed inset-4 z-50 bg-gray-900 border border-gray-700 rounded-lg shadow-2xl flex flex-col">
//       {/* Header */}
//       <div className="flex items-center justify-between px-4 py-3 bg-gray-800 border-b border-gray-700 rounded-t-lg">
//         <div className="flex items-center gap-3">
//           <div className="flex items-center gap-2">
//             <div className="w-3 h-3 rounded-full bg-red-500"></div>
//             <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
//             <div className="w-3 h-3 rounded-full bg-green-500"></div>
//           </div>
//           <h3 className="text-gray-200 font-mono text-sm">SQL Sharding Terminal</h3>
//           <span className="text-gray-500 text-xs">({logs.length} events)</span>
//         </div>
//         <div className="flex items-center gap-2">
//           <button
//             onClick={clearLogs}
//             className="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
//           >
//             Clear
//           </button>
//           <button
//             onClick={exportLogs}
//             className="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
//           >
//             Export
//           </button>
//           <button
//             onClick={() => setIsVisible(false)}
//             className="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
//           >
//             Minimize
//           </button>
//         </div>
//       </div>

//       {/* Terminal Content */}
//       <div 
//         ref={terminalRef}
//         className="flex-1 overflow-y-auto p-4 font-mono text-sm bg-black"
//         style={{ fontFamily: 'Consolas, Monaco, "Courier New", monospace' }}
//       >
//         {logs.length === 0 ? (
//           <div className="text-gray-600 text-center py-8">
//             Waiting for events...
//           </div>
//         ) : (
//           <div className="space-y-1">
//             {logs.map((log, index) => (
//               <div 
//                 key={index} 
//                 className={`border-l-2 pl-3 py-1 ${getLogLevelBg(log.level)}`}
//               >
//                 <div className="flex items-start gap-2">
//                   <span className="text-gray-500 text-xs whitespace-nowrap">
//                     {log.timestamp}
//                   </span>
//                   <span className={`font-semibold text-xs uppercase ${getLogLevelColor(log.level)} whitespace-nowrap`}>
//                     [{log.level}]
//                   </span>
//                   <span className="text-cyan-400 text-xs whitespace-nowrap">
//                     {log.source}:
//                   </span>
//                   <span className="text-gray-300 flex-1 break-words">
//                     {log.message}
//                   </span>
//                 </div>
//                 {log.fields && Object.keys(log.fields).length > 0 && (
//                   <div className="ml-6 mt-1 text-xs text-gray-500">
//                     {Object.entries(log.fields).map(([key, value]) => (
//                       <span key={key} className="mr-4">
//                         {key}={value}
//                       </span>
//                     ))}
//                   </div>
//                 )}
//               </div>
//             ))}
//           </div>
//         )}
//       </div>

//       {/* Footer */}
//       <div className="px-4 py-2 bg-gray-800 border-t border-gray-700 rounded-b-lg">
//         <div className="flex items-center justify-between">
//           <div className="text-xs text-gray-500">
//             Last updated: {logs.length > 0 ? logs[logs.length - 1].timestamp : 'Never'}
//           </div>
//           <div className="flex items-center gap-4 text-xs text-gray-500">
//             <div className="flex items-center gap-1">
//               <div className="w-2 h-2 rounded-full bg-blue-400"></div>
//               <span>Info</span>
//             </div>
//             <div className="flex items-center gap-1">
//               <div className="w-2 h-2 rounded-full bg-yellow-400"></div>
//               <span>Warning</span>
//             </div>
//             <div className="flex items-center gap-1">
//               <div className="w-2 h-2 rounded-full bg-red-400"></div>
//               <span>Error</span>
//             </div>
//           </div>
//         </div>
//       </div>
//     </div>
//   )
// }


import { useEffect, useRef, useState } from 'react'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { useAppStore } from '../stores/appStore'

interface LogEvent {
  level: 'info' | 'warn' | 'error'
  message: string
  source: string
  timestamp: string
  fields?: Record<string, string>
}

export function Terminal() {
  const [logs, setLogs] = useState<LogEvent[]>([])
  const [isVisible, setIsVisible] = useState(false)
  const terminalRef = useRef<HTMLDivElement>(null)
  const { addNotification } = useAppStore()

  useEffect(() => {
    const unsubscribe = EventsOn("log:", (event: LogEvent) => {
      setLogs(prev => [...prev.slice(-199), event])
      if (event.level === 'error') {
        addNotification({
          type: 'error',
          message: `${event.source}: ${event.message}`
        })
      }
    })

    return () => unsubscribe()
  }, [addNotification])

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight
    }
  }, [logs])

  const getLogLevelColor = (level: string) => {
    switch (level) {
      case 'error': return 'text-red-400'
      case 'warn': return 'text-yellow-400'
      case 'info': return 'text-blue-400'
      default: return 'text-gray-300'
    }
  }

  const getLogLevelBg = (level: string) => {
    switch (level) {
      case 'error': return 'bg-red-900/20 border-red-800/30'
      case 'warn': return 'bg-yellow-900/20 border-yellow-800/30'
      case 'info': return 'bg-blue-900/20 border-blue-800/30'
      default: return 'bg-gray-900/20 border-gray-800/30'
    }
  }

  const clearLogs = () => setLogs([])

  const exportLogs = () => {
    const logText = logs.map(log =>
      `[${log.timestamp}] [${log.level.toUpperCase()}] ${log.source}: ${log.message}`
    ).join('\n')

    const blob = new Blob([logText], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `sql-sharder-logs-${new Date().toISOString().split('T')[0]}.log`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <>
      {/* Toggle Button */}
      {!isVisible && (
        <button
          onClick={() => setIsVisible(true)}
          className="fixed bottom-4 right-4 z-40 px-4 py-2 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-lg shadow-lg border border-gray-700 transition-all duration-200 flex items-center gap-2"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          Terminal ({logs.length})
        </button>
      )}

      {/* Terminal Window */}
      <div
        className={`fixed left-0 bottom-0 w-full z-50 flex flex-col bg-gray-900 border-t border-gray-700 rounded-t-lg shadow-2xl overflow-hidden transition-transform duration-300 ease-in-out`}
        style={{
          transform: isVisible ? 'translateY(0%)' : 'translateY(100%)',  // fully off-screen
          height: '40vh',
          pointerEvents: isVisible ? 'auto' : 'none',                   // disable interactions when hidden
          opacity: isVisible ? 1 : 0,                                   // hide visually
        }}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 bg-gray-800 border-b border-gray-700">
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-red-500"></div>
              <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
              <div className="w-3 h-3 rounded-full bg-green-500"></div>
            </div>
            <h3 className="text-gray-200 font-mono text-sm">SQL Sharding Terminal</h3>
            <span className="text-gray-500 text-xs">({logs.length} events)</span>
          </div>
          <div className="flex items-center gap-2">
            <button onClick={clearLogs} className="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors">Clear</button>
            <button onClick={exportLogs} className="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors">Export</button>
            <button
              onClick={() => setIsVisible(false)}
              className="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors ease-in-out duration-100"
            >Minimize</button>
          </div>
        </div>

        {/* Terminal Content */}
        <div
          ref={terminalRef}
          className="flex-1 overflow-y-auto p-4 font-mono text-sm bg-black"
          style={{ fontFamily: 'Consolas, Monaco, "Courier New", monospace' }}
        >
          {logs.length === 0 ? (
            <div className="text-gray-600 text-center py-8">Waiting for events...</div>
          ) : (
            <div className="space-y-1">
              {logs.map((log, index) => (
                <div key={index} className={`border-l-2 pl-3 py-1 ${getLogLevelBg(log.level)}`}>
                  <div className="flex items-start gap-2">
                    <span className="text-gray-500 text-xs whitespace-nowrap">{log.timestamp}</span>
                    <span className={`font-semibold text-xs uppercase ${getLogLevelColor(log.level)} whitespace-nowrap`}>[{log.level}]</span>
                    <span className="text-cyan-400 text-xs whitespace-nowrap">{log.source}:</span>
                    <span className="text-gray-300 flex-1 break-words">{log.message}</span>
                  </div>
                  {log.fields && Object.keys(log.fields).length > 0 && (
                    <div className="ml-6 mt-1 text-xs text-gray-500">
                      {Object.entries(log.fields).map(([key, value]) => (
                        <span key={key} className="mr-4">{key}={value}</span>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  )
}
