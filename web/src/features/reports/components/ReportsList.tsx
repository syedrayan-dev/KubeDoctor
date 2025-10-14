import React from 'react';
import { Link } from 'react-router-dom';
import { 
  EyeIcon, 
  ExclamationTriangleIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ClockIcon 
} from '@heroicons/react/24/outline';
import Button from '../../../shared/components/ui/Button';
import Badge from '../../../shared/components/ui/Badge';
import Card from '../../../shared/components/ui/Card';
import { ROUTES } from '../../../shared/constants/routes';
import { PROBLEM_TYPE_LABELS } from '../../../shared/constants/status';
import { PROBLEM_TYPE_BADGE_COLORS } from '../../../shared/constants/colors';
import type { DiagnosisReport, PaginatedResponse } from '../../../shared/types/api';
import { formatDistanceToNow } from 'date-fns';

interface ReportsListProps {
  reports: PaginatedResponse<DiagnosisReport>;
  loading?: boolean;
  onPageChange: (page: number) => void;
}

const ReportsList: React.FC<ReportsListProps> = ({
  reports,
  loading = false,
  onPageChange,
}) => {
  const formatTimeAgo = (timestamp: string) => {
    try {
      return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
    } catch {
      return 'Unknown';
    }
  };

  const getProblemTypeLabel = (type: string) => {
    return PROBLEM_TYPE_LABELS[type as keyof typeof PROBLEM_TYPE_LABELS] || type;
  };

  const truncateText = (text: string, maxLength: number) => {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
  };


  if (loading) {
    return (
      <Card padding="none">
        <div className="p-6">
          <div className="animate-pulse space-y-4">
            {[...Array(10)].map((_, index) => (
              <div key={index} className="border-b border-gray-100 pb-4 last:border-b-0">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-2">
                      <div className="h-4 bg-gray-200 rounded w-32"></div>
                      <div className="h-5 bg-gray-200 rounded w-16"></div>
                      <div className="h-5 bg-gray-200 rounded w-20"></div>
                    </div>
                    <div className="h-3 bg-gray-200 rounded w-3/4 mb-2"></div>
                    <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                  </div>
                  <div className="ml-4 flex space-x-2">
                    <div className="h-8 bg-gray-200 rounded w-8"></div>
                    <div className="h-8 bg-gray-200 rounded w-8"></div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </Card>
    );
  }

  if (reports.items.length === 0) {
    return (
      <Card padding="lg">
        <div className="text-center py-12">
          <ExclamationTriangleIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No reports found</h3>
          <p className="text-gray-600 mb-4">
            No diagnosis reports match your current filters.
          </p>
          <p className="text-sm text-gray-500">
            Try adjusting your search criteria or removing some filters.
          </p>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* Reports List */}
      <Card padding="none">
        <div className="divide-y divide-gray-100">
          {reports.items.map((report) => (
            <div key={report.metadata.name} className="p-6 hover:bg-gray-50 transition-colors">
              <div className="flex items-center justify-between">
                <div className="flex-1 min-w-0">
                  {/* Header */}
                  <div className="flex items-center space-x-3 mb-3">
                    <h3 className="text-lg font-medium text-gray-900 truncate">
                      {report.spec.targetPod.name}
                    </h3>
                    <Badge variant="outline" size="sm">
                      {report.spec.targetPod.namespace}
                    </Badge>
                    <span 
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                        PROBLEM_TYPE_BADGE_COLORS[report.spec.triggerCondition.type as keyof typeof PROBLEM_TYPE_BADGE_COLORS] || 
                        'bg-gray-100 text-gray-800 border-gray-200'
                      }`}
                    >
                      {getProblemTypeLabel(report.spec.triggerCondition.type)}
                    </span>
                  </div>

                  {/* Summary */}
                  <p className="text-gray-700 mb-3">
                    {truncateText(report.spec.analysis.summary, 200)}
                  </p>

                  {/* Root Cause Preview */}
                  <p className="text-sm text-gray-600 mb-3">
                    <span className="font-medium">Root Cause:</span>{' '}
                    {truncateText(report.spec.analysis.rootCause, 150)}
                  </p>

                  {/* Metadata */}
                  <div className="flex items-center text-sm text-gray-500 space-x-4">
                    <div className="flex items-center">
                      <ClockIcon className="h-4 w-4 mr-1" />
                      {formatTimeAgo(report.metadata.creationTimestamp)}
                    </div>
                    <span>•</span>
                    <span>LLM: {report.spec.analysis.provider} {report.spec.analysis.model}</span>
                    <span>•</span>
                    <span>Processing: {report.spec.analysis.processingTime}</span>
                    <span>•</span>
                    <span>Logs: {report.spec.collectedData.logLines} lines</span>
                  </div>

                  {/* Recommendations Count */}
                  <div className="mt-2">
                    <span className="text-sm text-gray-600">
                      {report.spec.analysis.recommendations.length} recommendation(s)
                    </span>
                  </div>
                </div>

                {/* Actions */}
                <div className="ml-6 flex-shrink-0 flex space-x-2">
                  <Link to={`${ROUTES.REPORTS}/${report.metadata.name}`}>
                    <Button variant="outline" size="sm">
                      <EyeIcon className="h-4 w-4 mr-2" />
                      View Details
                    </Button>
                  </Link>
                </div>
              </div>
            </div>
          ))}
        </div>
      </Card>

      {/* Pagination */}
      {reports.totalCount > reports.pageSize && (
        <Card padding="sm">
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-700">
              Showing {(reports.page - 1) * reports.pageSize + 1} to{' '}
              {Math.min(reports.page * reports.pageSize, reports.totalCount)} of{' '}
              {reports.totalCount} results
            </div>
            
            <div className="flex items-center space-x-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => onPageChange(reports.page - 1)}
                disabled={!reports.hasPrevious}
              >
                <ChevronLeftIcon className="h-4 w-4" />
              </Button>
              
              <span className="text-sm text-gray-700">
                Page {reports.page}
              </span>
              
              <Button
                variant="outline"
                size="sm"
                onClick={() => onPageChange(reports.page + 1)}
                disabled={!reports.hasNext}
              >
                <ChevronRightIcon className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
};

export default ReportsList;