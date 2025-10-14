export interface DashboardStats {
  totalReports: number;
  activeIssues: number;
  successRate: number;
  averageResolutionTime: string;
  trendsData: {
    reports: Array<{
      date: string;
      count: number;
    }>;
    issues: Array<{
      date: string;
      count: number;
    }>;
  };
}

export interface ProblemDistribution {
  type: string;
  count: number;
  percentage: number;
  trend: {
    value: number;
    isIncreasing: boolean;
  };
}

export interface RecentActivity {
  type: 'report_created' | 'issue_resolved' | 'policy_updated' | 'diagnosis_failed';
  message: string;
  timestamp: string;
  metadata?: {
    podName?: string;
    namespace?: string;
    policyName?: string;
  };
}

export interface SystemHealth {
  operatorStatus: 'healthy' | 'degraded' | 'unhealthy';
  apiConnectivity: boolean;
  llmProviderStatus: 'available' | 'limited' | 'unavailable';
  lastHealthCheck: string;
  errorCount: number;
}

export interface DashboardFilters {
  timeRange: '1h' | '24h' | '7d' | '30d';
  namespaces: string[];
  problemTypes: string[];
}

export interface DashboardProps {
  refreshInterval?: number;
  autoRefresh?: boolean;
}