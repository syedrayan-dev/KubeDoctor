import { useState, useEffect } from 'react';
import apiService from '../../../shared/services/apiService';
import type { DiagnosisReport, DiagnosisRequest } from '../../../shared/types/api';

interface UseReportDetailReturn {
  report: DiagnosisReport | null;
  request: DiagnosisRequest | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export const useReportDetail = (reportId: string): UseReportDetailReturn => {
  const [report, setReport] = useState<DiagnosisReport | null>(null);
  const [request, setRequest] = useState<DiagnosisRequest | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchReport = async () => {
    if (!reportId) {
      setError('Report ID is required');
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      const report = await apiService.getReport(reportId);
      setReport(report);

      // Try to find the associated request
      // The request name is usually the same as report name or follows a similar pattern
      try {
        const allRequests = await apiService.getRequests({});
        
        // Find request that matches this report
        const matchingRequest = allRequests.find(req => {
          // Match by similar name pattern or by target pod and timestamp
          return req.metadata.name === reportId || 
                 (req.spec.targetPod.name === report.spec.targetPod.name && 
                  req.spec.targetPod.namespace === report.spec.targetPod.namespace &&
                  req.status.phase === 'Completed');
        });

        setRequest(matchingRequest || null);
      } catch (err) {
        console.error('Failed to fetch associated request:', err);
        // Not critical if request fetch fails
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch report');
      setReport(null);
      setRequest(null);
    } finally {
      setLoading(false);
    }
  };

  const refetch = () => {
    fetchReport();
  };

  useEffect(() => {
    fetchReport();
  }, [reportId]);

  return {
    report,
    request,
    loading,
    error,
    refetch,
  };
};