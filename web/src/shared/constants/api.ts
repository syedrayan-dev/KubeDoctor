// Runtime environment variable from window object (injected by entrypoint script)
// Falls back to build-time env var, then to /api for local development
export const API_BASE_URL = (window as any).__ENV__?.VITE_API_BASE_URL || import.meta.env.VITE_API_BASE_URL || '/api';

export const API_ENDPOINTS = {
  // Apollo Diagnosis API endpoints
  DASHBOARD: '/diagnosis/v1alpha1/metrics',
  REPORTS: '/diagnosis/v1alpha1/reports',
  REQUESTS: '/diagnosis/v1alpha1/requests',
  POLICIES: '/diagnosis/v1alpha1/policies',
  // Kubernetes API endpoints
  NAMESPACES: '/v1/namespaces',
  PODS: '/v1/namespaces',
} as const;

export const POLLING_INTERVALS = {
  DASHBOARD: 30000,    // 30s
  REPORTS: 60000,      // 1min
  REQUESTS: 5000,      // 5s
} as const;

export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  INTERNAL_SERVER_ERROR: 500,
} as const;