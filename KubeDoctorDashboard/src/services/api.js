import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8000'

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
})

export const fetchPods = async () => {
  try {
    const response = await apiClient.get('/api/pods')
    return response.data
  } catch (error) {
    throw new Error(`Failed to fetch pods: ${error.message}`)
  }
}

export const fetchNodes = async () => {
  try {
    const response = await apiClient.get('/api/nodes')
    return response.data
  } catch (error) {
    throw new Error(`Failed to fetch nodes: ${error.message}`)
  }
}

export const fetchHealth = async () => {
  try {
    const response = await apiClient.get('/api/health')
    return response.data
  } catch (error) {
    throw new Error(`Health check failed: ${error.message}`)
  }
}
