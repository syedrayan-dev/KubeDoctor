import React from 'react';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import Layout from '../../shared/components/layout/Layout';
import MetricsCards from './components/MetricsCards';
import RecentReports from './components/RecentReports';
import ProblemChart from './components/ProblemChart';
import Button from '../../shared/components/ui/Button';
import { useDashboard } from './hooks/useDashboard';

const DashboardPage: React.FC = () => {
  const { metrics, loading, error, refetch } = useDashboard();

  if (error) {
    return (
      <Layout title="Dashboard" subtitle="Monitor your Kubernetes pod diagnostics" onRefresh={refetch}>
        <div className="text-center py-12">
          <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">Failed to load dashboard</h3>
          <p className="text-gray-600 mb-4">{error}</p>
          <Button onClick={refetch}>
            Try Again
          </Button>
        </div>
      </Layout>
    );
  }

  return (
    <Layout title="Dashboard" subtitle="Monitor your Kubernetes pod diagnostics" onRefresh={refetch}>
      <div className="space-y-6">
        {/* Metrics Cards */}
        <MetricsCards 
          metrics={metrics ? {
            totalReports: metrics.totalReports,
            activePolicies: metrics.activePolicies,
            successRate: metrics.successRate,
            averageResolutionTime: metrics.averageResolutionTime,
            trends: metrics.trends,
          } : {
            totalReports: 0,
            activePolicies: 0,
            successRate: 0,
            averageResolutionTime: '0s',
          }}
          loading={loading}
        />

        {/* Recent Reports and Problem Distribution */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Recent Reports */}
          <div>
            <RecentReports 
              reports={metrics?.recentReports || []}
              loading={loading}
            />
          </div>

          {/* Problem Distribution Chart */}
          <div>
            <ProblemChart 
              data={metrics?.problemDistribution || []}
              loading={loading}
            />
          </div>
        </div>


        {/* Status Footer */}
        <div className="text-center text-sm text-gray-500">
          <p>
            Dashboard updates every 30 seconds • 
            Last updated: {new Date().toLocaleTimeString()}
          </p>
        </div>
      </div>
    </Layout>
  );
};

export default DashboardPage;