import { useState, useEffect } from 'react';
import apiService from '../../../shared/services/apiService';
import type { DashboardMetrics } from '../../../shared/types/api';

interface UseDashboardReturn {
  metrics: DashboardMetrics | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export const useDashboard = (): UseDashboardReturn => {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchMetrics = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const metrics = await apiService.getDashboardMetrics();
      setMetrics(metrics);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch dashboard metrics');
      setMetrics(null);
    } finally {
      setLoading(false);
    }
  };

  const refetch = () => {
    fetchMetrics();
  };

  useEffect(() => {
    fetchMetrics();
  }, []);

  // Auto-refresh every 30 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      if (!loading) {
        fetchMetrics();
      }
    }, 30000);

    return () => clearInterval(interval);
  }, [loading]);

  return {
    metrics,
    loading,
    error,
    refetch,
  };
};