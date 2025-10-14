// Common utility types
export type Prettify<T> = {
  [K in keyof T]: T[K];
} & {};

// Loading states
export interface LoadingState {
  isLoading: boolean;
  error: string | null;
}

// API request state
export interface ApiRequestState<T = any> extends LoadingState {
  data: T | null;
}

// Pagination parameters
export interface PaginationParams {
  page: number;
  pageSize: number;
}

// Sort parameters
export interface SortParams {
  field: string;
  direction: 'asc' | 'desc';
}

// Generic list query parameters
export interface ListQueryParams extends PaginationParams {
  sort?: SortParams;
  filters?: Record<string, any>;
  search?: string;
}

// Chart data types
export interface PieChartData {
  name: string;
  value: number;
  color: string;
}

export interface LineChartData {
  timestamp: string;
  value: number;
}

export interface BarChartData {
  category: string;
  value: number;
}

// Form field types
export interface FormField<T = string> {
  value: T;
  error?: string;
  touched?: boolean;
}

export interface FormState<T extends Record<string, any>> {
  fields: {
    [K in keyof T]: FormField<T[K]>;
  };
  isValid: boolean;
  isSubmitting: boolean;
}

// Navigation item type
export interface NavItem {
  path: string;
  label: string;
  icon: string;
  badge?: number;
}

// Theme types
export type ThemeColor = 
  | 'primary'
  | 'secondary' 
  | 'success'
  | 'warning'
  | 'error'
  | 'info';

export type Size = 'sm' | 'md' | 'lg' | 'xl';

export type Variant = 'default' | 'outline' | 'ghost' | 'solid';

// Status types
export type RequestStatus = 'Pending' | 'InProgress' | 'Completed' | 'Failed';
export type ProblemType = 'Failed' | 'Pending' | 'Running' | 'Unknown';
export type DiagnosisRequestType = 'Automatic' | 'Manual';

// Event handler types
export type EventHandler<T = void> = (event?: any) => T;
export type ChangeHandler<T = string> = (value: T) => void;

// Component props helpers
export interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
}

export interface WithId {
  id: string;
}

export interface WithTestId {
  testId?: string;
}

// Error types
export interface ApiError {
  message: string;
  code?: string | number;
  details?: any;
}

// Time-related types
export interface TimeRange {
  start: Date;
  end: Date;
}

export interface Duration {
  value: number;
  unit: 'seconds' | 'minutes' | 'hours' | 'days';
}