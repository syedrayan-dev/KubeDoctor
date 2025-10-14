import React from 'react';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import Layout from '../../shared/components/layout/Layout';
import ReportFilters from './components/ReportFilters';
import ReportsList from './components/ReportsList';
import Button from '../../shared/components/ui/Button';
import { useReports } from './hooks/useReports';

const ReportsPage: React.FC = () => {
  const {
    reports,
    loading,
    error,
    filters,
    namespaces,
    setFilters,
    setPage,
    refetch,
  } = useReports();


  if (error) {
    return (
      <Layout title="Diagnosis Reports" subtitle="View detailed analysis results">
        <div className="text-center py-12">
          <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">Failed to load reports</h3>
          <p className="text-gray-600 mb-4">{error}</p>
          <Button onClick={refetch}>
            Try Again
          </Button>
        </div>
      </Layout>
    );
  }

  return (
    <Layout title="Diagnosis Reports" subtitle="View detailed analysis results">
      <div className="space-y-6">
        {/* Summary Stats */}
        <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="text-2xl font-bold text-gray-900">{reports.totalCount}</div>
            <div className="text-sm text-gray-600">Total Reports</div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="text-2xl font-bold text-red-600">{reports.items.filter(r => r.spec.triggerCondition.type === 'Failed').length}</div>
            <div className="text-sm text-gray-600">Failed</div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="text-2xl font-bold text-yellow-600">{reports.items.filter(r => r.spec.triggerCondition.type === 'Pending').length}</div>
            <div className="text-sm text-gray-600">Pending</div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="text-2xl font-bold text-blue-600">{reports.items.filter(r => r.spec.triggerCondition.type === 'Running').length}</div>
            <div className="text-sm text-gray-600">Running</div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="text-2xl font-bold text-gray-600">{reports.items.filter(r => r.spec.triggerCondition.type === 'Unknown').length}</div>
            <div className="text-sm text-gray-600">Unknown</div>
          </div>
        </div>

        {/* Filters */}
        <ReportFilters
          filters={filters}
          onFiltersChange={setFilters}
          namespaces={namespaces}
          loading={loading}
        />

        {/* Reports List */}
        <ReportsList
          reports={reports}
          loading={loading}
          onPageChange={setPage}
        />
      </div>
    </Layout>
  );
};

export default ReportsPage;