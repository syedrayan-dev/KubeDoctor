import type { DiagnosisReport, ReportFilters } from '../../../shared/types/api';

export interface ReportsPageProps {
  initialFilters?: ReportFilters;
}

export interface ReportListItem extends DiagnosisReport {
  // Additional computed fields for list display
  severity?: 'low' | 'medium' | 'high' | 'critical';
  status?: 'new' | 'reviewed' | 'resolved';
  tags?: string[];
}

export interface ReportSortOption {
  field: 'creationTimestamp' | 'podName' | 'namespace' | 'problemType' | 'processingTime';
  direction: 'asc' | 'desc';
  label: string;
}

export interface ReportAction {
  id: string;
  label: string;
  icon?: string;
  action: (report: DiagnosisReport) => void;
  disabled?: (report: DiagnosisReport) => boolean;
}

export interface ReportExportOptions {
  format: 'json' | 'csv' | 'pdf';
  includeMetadata: boolean;
  includeLogs: boolean;
  includeRecommendations: boolean;
}

export interface ReportStats {
  totalReports: number;
  byProblemType: Record<string, number>;
  byNamespace: Record<string, number>;
  byTimeRange: {
    last24h: number;
    last7d: number;
    last30d: number;
  };
  avgProcessingTime: number;
  successRate: number;
}

export interface ReportDetailPageProps {
  reportId: string;
  onNavigateBack?: () => void;
  showRelatedReports?: boolean;
}

export interface RelatedReport {
  id: string;
  name: string;
  namespace: string;
  problemType: string;
  similarity: number;
  createdAt: string;
}

export interface ReportComparison {
  baseReport: DiagnosisReport;
  compareReport: DiagnosisReport;
  differences: {
    summary: boolean;
    rootCause: boolean;
    recommendations: boolean;
    logs: boolean;
  };
  similarity: number;
}