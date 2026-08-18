import { useState, useEffect } from 'react'
import Header from './components/Header'
import PodList from './components/PodList'
import NodeList from './components/NodeList'
import { fetchPods, fetchNodes } from './services/api'
import './App.css'

function App() {
  const [pods, setPods] = useState([])
  const [nodes, setNodes] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [activeTab, setActiveTab] = useState('pods')

  const loadData = async () => {
    setLoading(true)
    setError(null)
    try {
      const [podsData, nodesData] = await Promise.all([
        fetchPods(),
        fetchNodes(),
      ])
      setPods(podsData)
      setNodes(nodesData)
    } catch (err) {
      setError(err.message)
      console.error('Failed to fetch cluster data:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
    // Auto-refresh every 5 seconds
    const interval = setInterval(loadData, 5000)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      
      <main className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        {/* Error Alert */}
        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-800">
              <span className="font-semibold">Error:</span> {error}
            </p>
            <button
              onClick={loadData}
              className="mt-2 px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700 text-sm"
            >
              Retry
            </button>
          </div>
        )}

        {/* Tabs */}
        <div className="flex space-x-4 mb-6 border-b border-gray-200">
          <button
            onClick={() => setActiveTab('pods')}
            className={`px-4 py-2 font-medium ${
              activeTab === 'pods'
                ? 'border-b-2 border-blue-500 text-blue-600'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Pods ({pods.length})
          </button>
          <button
            onClick={() => setActiveTab('nodes')}
            className={`px-4 py-2 font-medium ${
              activeTab === 'nodes'
                ? 'border-b-2 border-blue-500 text-blue-600'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Nodes ({nodes.length})
          </button>
        </div>

        {/* Content */}
        {loading && pods.length === 0 && nodes.length === 0 && (
          <div className="flex justify-center items-center h-64">
            <div className="text-center">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              <p className="mt-4 text-gray-600">Loading cluster data...</p>
            </div>
          </div>
        )}

        {activeTab === 'pods' && pods.length > 0 && <PodList pods={pods} loading={loading} />}
        {activeTab === 'nodes' && nodes.length > 0 && <NodeList nodes={nodes} loading={loading} />}
      </main>
    </div>
  )
}

export default App
