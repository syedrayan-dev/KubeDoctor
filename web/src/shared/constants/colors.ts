import { REQUEST_STATUS } from './status';

export const STATUS_COLORS = {
  [REQUEST_STATUS.PENDING]: 'bg-white text-yellow-600 border-yellow-300',
  [REQUEST_STATUS.IN_PROGRESS]: 'bg-white text-blue-600 border-blue-300',
  [REQUEST_STATUS.COMPLETED]: 'bg-white text-green-600 border-green-300',
  [REQUEST_STATUS.FAILED]: 'bg-white text-red-600 border-red-300',
} as const;

export const CHART_COLORS = [
  '#EF4444CC', // Red - Failed (80% opacity)
  '#F59E0BCC', // Orange - Pending (80% opacity)
  '#3B82F6CC', // Blue - Running (80% opacity)
  '#6B7280CC', // Gray - Unknown (80% opacity)
  '#8B5CF6CC', // Purple (80% opacity)
  '#F97316CC', // Orange-500 (80% opacity)
  '#EC4899CC', // Pink-500 (80% opacity)
  '#84CC16CC', // Lime-500 (80% opacity)
] as const;

// Pod Phase specific colors
export const POD_PHASE_COLORS = {
  Failed: '#EF4444CC',   // Red (80% opacity)
  Pending: '#F59E0BCC',  // Orange (80% opacity)
  Running: '#3B82F6CC',  // Blue (80% opacity)
  Unknown: '#6B7280CC',  // Gray (80% opacity)
} as const;

export const SEVERITY_COLORS = {
  LOW: 'text-green-600 bg-green-50',
  MEDIUM: 'text-yellow-600 bg-yellow-50',
  HIGH: 'text-red-600 bg-red-50',
  CRITICAL: 'text-red-800 bg-red-100',
} as const;

// Problem type badge colors
export const PROBLEM_TYPE_BADGE_COLORS = {
  Failed: 'bg-white text-red-600 border-red-300',
  Pending: 'bg-white text-yellow-600 border-yellow-300',
  Running: 'bg-white text-blue-600 border-blue-300',
  Unknown: 'bg-white text-gray-600 border-gray-300',
} as const;