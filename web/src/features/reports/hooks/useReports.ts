import { useState, useEffect, useCallback } from 'react';
import apiService from '../../../shared/services/apiService';
import type { DiagnosisReport, PaginatedResponse, ReportFilters } from '../../../shared/types/api';

// Custom hook for debouncing
const useDebounce = (value: string, delay: number) => {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
};

interface UseReportsReturn {
  reports: PaginatedResponse<DiagnosisReport>;
  loading: boolean;
  error: string | null;
  filters: ReportFilters;
  namespaces: string[];
  setFilters: (filters: ReportFilters) => void;
  setPage: (page: number) => void;
  refetch: () => void;
}

export const useReports = (): UseReportsReturn => {
  const [reports, setReports] = useState<PaginatedResponse<DiagnosisReport>>({
    items: [],
    totalCount: 0,
    page: 1,
    pageSize: 20,
    hasNext: false,
    hasPrevious: false,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<ReportFilters>({});
  const [page, setPage] = useState(1);
  const [namespaces, setNamespaces] = useState<string[]>([]);
  
  // Debounce search term
  const debouncedSearch = useDebounce(filters.search || '', 300);

  const fetchReports = useCallback(async (skipLoadingForSearch = false, currentFilters = filters, currentPage = page, searchTerm = debouncedSearch) => {
    try {
      if (!skipLoadingForSearch) {
        setLoading(true);
      }
      setError(null);
      
      const apiReports = await apiService.getReports({
        namespace: currentFilters.namespace,
        type: currentFilters.problemType,
      });
      
      // Ensure apiReports is an array
      if (!Array.isArray(apiReports)) {
        throw new Error('Invalid response format from API');
      }
      
      // Client-side filtering for other filters
      let filteredReports = [...apiReports];
      
      if (searchTerm) {
        const searchLower = searchTerm.toLowerCase();
        filteredReports = filteredReports.filter(report =>
          report.metadata.name.toLowerCase().includes(searchLower)
        );
      }
      
      if (filters.dateFrom) {
        const fromDate = new Date(filters.dateFrom);
        filteredReports = filteredReports.filter(report =>
          new Date(report.metadata.creationTimestamp) >= fromDate
        );
      }
      
      if (filters.dateTo) {
        const toDate = new Date(filters.dateTo);
        filteredReports = filteredReports.filter(report =>
          new Date(report.metadata.creationTimestamp) <= toDate
        );
      }
      
      // Sort by creation time (newest first)
      filteredReports.sort((a, b) => 
        new Date(b.metadata.creationTimestamp).getTime() - 
        new Date(a.metadata.creationTimestamp).getTime()
      );
      
      // Pagination
      const pageSize = 20;
      const totalCount = filteredReports.length;
      const startIndex = (currentPage - 1) * pageSize;
      const endIndex = startIndex + pageSize;
      const paginatedItems = filteredReports.slice(startIndex, endIndex);
      
      setReports({
        items: paginatedItems,
        totalCount,
        page: currentPage,
        pageSize,
        hasNext: endIndex < totalCount,
        hasPrevious: currentPage > 1,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch reports');
      setReports({
        items: [],
        totalCount: 0,
        page: 1,
        pageSize: 20,
        hasNext: false,
        hasPrevious: false,
      });
    } finally {
      if (!skipLoadingForSearch) {
        setLoading(false);
      }
    }
  }, [filters.namespace, filters.problemType, page, debouncedSearch]);

  const fetchNamespaces = async () => {
    try {
      const apiNamespaces = await apiService.getNamespaces();
      setNamespaces(apiNamespaces.map(ns => ns.metadata.name));
    } catch (err) {
      console.error('Failed to fetch namespaces:', err);
      setNamespaces([]);
    }
  };

  const handleSetFilters = useCallback((newFilters: ReportFilters) => {
    setFilters(newFilters);
    setPage(1); // Reset to first page when filters change
  }, []);

  const handleSetPage = (newPage: number) => {
    setPage(newPage);
  };

  const refetch = useCallback(() => {
    fetchReports(false, filters, page, debouncedSearch);
  }, [fetchReports, filters, page, debouncedSearch]);

  useEffect(() => {
    fetchNamespaces();
  }, []);

  // Handle all filters including search
  useEffect(() => {
    fetchReports(debouncedSearch !== (filters.search || ''), filters, page, debouncedSearch);
  }, [debouncedSearch, filters.namespace, filters.problemType, page, fetchReports]);

  return {
    reports,
    loading,
    error,
    filters,
    namespaces,
    setFilters: handleSetFilters,
    setPage: handleSetPage,
    refetch,
  };
};