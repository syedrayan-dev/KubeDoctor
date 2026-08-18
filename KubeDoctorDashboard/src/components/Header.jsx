import { useState, useEffect } from 'react'
import { fetchHealth } from '../services/api'

function Header() {
  const [clusterInfo, setClusterInfo] = useState(null)
  const [connected, setConnected] = useState(false)

  useEffect(() => {
    const checkHealth = async () => {
      try {
        const health = await fetchHealth()
        setClusterInfo(health.cluster)
        setConnected(true)
      } catch (error) {
        setConnected(false)
      }
    }

    checkHealth()
    const interval = setInterval(checkHealth, 10000)
    return () => clearInterval(interval)
  }, [])

  return (
    <header className="bg-white shadow">
      <div className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8 flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">KubeDoctor</h1>
          <p className="text-gray-600 text-sm mt-1">Kubernetes Cluster Dashboard</p>
        </div>
        <div className="flex items-center gap-4">
          {clusterInfo && (
            <div className="text-right">
              <p className="text-sm text-gray-600">Cluster</p>
              <p className="font-semibold text-gray-900">{clusterInfo}</p>
            </div>
          )}
          <div className={`flex items-center gap-2 ${
            connected ? 'text-green-600' : 'text-red-600'
          }`}>
            <div className={`w-3 h-3 rounded-full ${
              connected ? 'bg-green-600' : 'bg-red-600'
            }`}></div>
            <span className="text-sm font-medium">
              {connected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
        </div>
      </div>
    </header>
  )
}

export default Header
