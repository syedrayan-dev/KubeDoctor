// Base Kubernetes metadata interface
export interface KubernetesMetadata {
  name: string;
  namespace: string;
  uid?: string;
  creationTimestamp: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
}

// Pod reference interface
export interface PodReference {
  name: string;
  namespace: string;
  uid?: string;
  image?: string;
}

// Policy reference interface
export interface PolicyReference {
  name: string;
  namespace: string;
}

// Secret reference interface
export interface SecretReference {
  name: string;
  namespace: string;
  key: string;
}

// Trigger condition interface for reports
export interface TriggerCondition {
  type: string;
  detectedAt: string;
}

// Pod condition check for detailed conditions
export interface PodConditionCheck {
  name: string;
  status: string;
  minDuration?: string;
}

// Pod condition trigger for policies
export interface PodConditionTrigger {
  type: string;
  minDuration?: string;
  conditions?: PodConditionCheck[];
}

// LLM configuration interface
export interface LLMConfiguration {
  provider: string;
  model: string;
  apiKeySecretRef: SecretReference;
  baseURL?: string;
}

// Collected data interface
export interface CollectedData {
  podDescription: string;
  logs: string;
  logLines: number;
  collectionTime: string;
}

// Diagnosis analysis interface
export interface DiagnosisAnalysis {
  summary: string;
  rootCause: string;
  recommendations: string[];
  provider: string;
  model: string;
  processingTime: string;
}

// DiagnosisReport interface
export interface DiagnosisReport {
  metadata: KubernetesMetadata;
  spec: {
    targetPod: PodReference;
    triggerCondition: TriggerCondition;
    policyRef: PolicyReference;
    collectedData: CollectedData;
    analysis: DiagnosisAnalysis;
  };
  status: {
    phase: 'Pending' | 'InProgress' | 'Completed' | 'Failed';
    message: string;
    lastUpdateTime: string;
    completionTime?: string;
  };
}

// DiagnosisRequest interface
export interface DiagnosisRequest {
  metadata: KubernetesMetadata;
  spec: {
    type: 'Automatic' | 'Manual';
    targetPod: {
      name: string;
      namespace: string;
    };
    policyRef: PolicyReference;
    triggerCondition?: TriggerCondition;
  };
  status: {
    phase: 'Pending' | 'InProgress' | 'Completed' | 'Failed';
    message: string;
    lastUpdateTime: string;
    completionTime?: string;
  };
}

// DiagnosisPolicy interface
export interface DiagnosisPolicy {
  metadata: KubernetesMetadata;
  spec: {
    targetNamespaces: string[];
    podSelector?: {
      matchLabels?: Record<string, string>;
      matchExpressions?: Array<{
        key: string;
        operator: string;
        values?: string[];
      }>;
    };
    llmConfig: LLMConfiguration;
    triggerConditions: PodConditionTrigger[];
    diagnosisSettings?: {
      maxLogLines?: number;
      retryInterval?: string;
    };
  };
}

// API Response interfaces
export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  totalCount: number;
  page: number;
  pageSize: number;
  hasNext: boolean;
  hasPrevious: boolean;
}

// Dashboard metrics interface
export interface DashboardMetrics {
  totalReports: number;
  activePolicies: number;
  successRate: number;
  averageResolutionTime: string;
  problemDistribution: Array<{
    type: string;
    count: number;
    percentage: number;
  }>;
  recentReports: DiagnosisReport[];
  trends?: {
    totalReports?: { value: number; isPositive: boolean };
    activePolicies?: { value: number; isPositive: boolean };
    successRate?: { value: number; isPositive: boolean };
    averageResolutionTime?: { value: number; isPositive: boolean };
  };
}

// Filter interfaces
export interface ReportFilters {
  namespace?: string;
  problemType?: string;
  status?: string;
  dateFrom?: string;
  dateTo?: string;
  search?: string;
}

export interface RequestFilters {
  namespace?: string;
  type?: 'Automatic' | 'Manual';
  status?: string;
  dateFrom?: string;
  dateTo?: string;
}

// Kubernetes resources
export interface Namespace {
  metadata: {
    name: string;
    creationTimestamp: string;
  };
  status: {
    phase: string;
  };
}

export interface Pod {
  metadata: KubernetesMetadata;
  spec: {
    containers: Array<{
      name: string;
      image: string;
    }>;
    nodeName?: string;
  };
  status: {
    phase: string;
    conditions?: Array<{
      type: string;
      status: string;
      reason?: string;
      message?: string;
    }>;
  };
}