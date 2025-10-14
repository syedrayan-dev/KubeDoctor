import React from 'react';
import { 
  DocumentTextIcon, 
  CogIcon, 
  CheckCircleIcon, 
  ClockIcon 
} from '@heroicons/react/24/outline';
import Card from '../../../shared/components/ui/Card';

interface MetricCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ComponentType<{ className?: string }>;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  color?: 'blue' | 'green' | 'yellow' | 'red';
}

const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  subtitle,
  icon: Icon,
  trend,
  color = 'blue'
}) => {
  const colorClasses = {
    blue: 'text-blue-600 bg-blue-50',
    green: 'text-green-600 bg-green-50',
    yellow: 'text-yellow-600 bg-yellow-50',
    red: 'text-red-600 bg-red-50',
  };

  return (
    <Card padding="md" hover>
      <div className="flex items-center">
        <div className={`p-3 rounded-full ${colorClasses[color]}`}>
          <Icon className="h-6 w-6" />
        </div>
        <div className="ml-4 flex-1">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium text-gray-600">{title}</p>
            {trend && (
              <span className={`flex items-center text-xs font-medium ${
                trend.value === 0 ? 'text-gray-500' : (trend.isPositive ? 'text-green-600' : 'text-red-600')
              }`}>
                <span className="mr-1">
                  {trend.value === 0 ? '+' : (trend.isPositive ? '+' : '-')}
                </span>
                {trend.value}{title.includes('Rate') || title.includes('Time') ? '%' : ''}
              </span>
            )}
          </div>
          <p className="text-2xl font-bold text-gray-900">{value}</p>
          {subtitle && (
            <p className="text-sm text-gray-500">{subtitle}</p>
          )}
        </div>
      </div>
    </Card>
  );
};

interface MetricsCardsProps {
  metrics: {
    totalReports: number;
    activePolicies: number;
    successRate: number;
    averageResolutionTime: string;
    trends?: {
      totalReports?: { value: number; isPositive: boolean };
      activePolicies?: { value: number; isPositive: boolean };
      successRate?: { value: number; isPositive: boolean };
      averageResolutionTime?: { value: number; isPositive: boolean };
    };
  };
  loading?: boolean;
}

const MetricsCards: React.FC<MetricsCardsProps> = ({ metrics, loading = false }) => {
  if (loading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[...Array(4)].map((_, index) => (
          <Card key={index} padding="md">
            <div className="animate-pulse">
              <div className="flex items-center">
                <div className="w-12 h-12 bg-gray-200 rounded-full"></div>
                <div className="ml-4 flex-1">
                  <div className="h-4 bg-gray-200 rounded mb-2"></div>
                  <div className="h-8 bg-gray-200 rounded mb-1"></div>
                  <div className="h-3 bg-gray-200 rounded w-3/4"></div>
                </div>
              </div>
            </div>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <MetricCard
        title="Total Reports"
        value={metrics.totalReports}
        subtitle="Last 24 hours"
        icon={DocumentTextIcon}
        color="blue"
        trend={metrics.trends?.totalReports}
      />
      <MetricCard
        title="Active Policies"
        value={metrics.activePolicies}
        subtitle="Currently enabled"
        icon={CogIcon}
        color="yellow"
        trend={metrics.trends?.activePolicies}
      />
      <MetricCard
        title="Success Rate"
        value={`${metrics.successRate}%`}
        subtitle="Successful diagnoses"
        icon={CheckCircleIcon}
        color="green"
        trend={metrics.trends?.successRate}
      />
      <MetricCard
        title="Avg Resolution Time"
        value={metrics.averageResolutionTime}
        subtitle="Time to diagnosis"
        icon={ClockIcon}
        color="blue"
        trend={metrics.trends?.averageResolutionTime}
      />
    </div>
  );
};

export default MetricsCards;