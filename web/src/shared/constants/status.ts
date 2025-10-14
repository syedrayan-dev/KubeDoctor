export const REQUEST_STATUS = {
  PENDING: 'Pending',
  IN_PROGRESS: 'InProgress',
  COMPLETED: 'Completed',
  FAILED: 'Failed',
} as const;

export const PROBLEM_TYPES = {
  FAILED: 'Failed',
  PENDING: 'Pending',
  RUNNING: 'Running',
  UNKNOWN: 'Unknown',
} as const;

export const LLM_PROVIDERS = {
  OPENAI: 'openai',
  AZURE: 'azure-openai',
  ANTHROPIC: 'anthropic',
  LOCAL: 'local',
} as const;

export const DIAGNOSIS_REQUEST_TYPE = {
  AUTOMATIC: 'Automatic',
  MANUAL: 'Manual',
} as const;

// Status display labels
export const STATUS_LABELS = {
  [REQUEST_STATUS.PENDING]: 'Pending',
  [REQUEST_STATUS.IN_PROGRESS]: 'In Progress',
  [REQUEST_STATUS.COMPLETED]: 'Completed',
  [REQUEST_STATUS.FAILED]: 'Failed',
} as const;

// Problem type display labels
export const PROBLEM_TYPE_LABELS = {
  [PROBLEM_TYPES.FAILED]: 'Failed',
  [PROBLEM_TYPES.PENDING]: 'Pending', 
  [PROBLEM_TYPES.RUNNING]: 'Running',
  [PROBLEM_TYPES.UNKNOWN]: 'Unknown',
} as const;