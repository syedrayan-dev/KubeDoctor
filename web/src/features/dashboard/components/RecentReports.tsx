import React from 'react';
import { Link } from 'react-router-dom';
import { EyeIcon, ClockIcon } from '@heroicons/react/24/outline';
import Card, { CardHeader, CardTitle, CardBody } from '../../../shared/components/ui/Card';
import Badge from '../../../shared/components/ui/Badge';
import Button from '../../../shared/components/ui/Button';
import { ROUTES } from '../../../shared/constants/routes';
import { PROBLEM_TYPE_LABELS } from '../../../shared/constants/status';
import { PROBLEM_TYPE_BADGE_COLORS } from '../../../shared/constants/colors';
import type { DiagnosisReport } from '../../../shared/types/api';
import { formatDistanceToNow } from 'date-fns';

interface RecentReportsProps {
  reports: DiagnosisReport[];
  loading?: boolean;
}

const RecentReports: React.FC<RecentReportsProps> = ({ reports, loading = false }) => {
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
      <Card className="h-[480px] flex flex-col">
        <CardHeader>
          <CardTitle>Recent Reports</CardTitle>
        </CardHeader>
        <CardBody className="flex-1 overflow-hidden">
          <div className="space-y-4 h-full">
            {[...Array(3)].map((_, index) => (
              <div key={index} className="animate-pulse">
                <div className="flex items-center justify-between py-3 border-b border-gray-100 last:border-b-0">
                  <div className="flex-1">
                    <div className="h-4 bg-gray-200 rounded mb-2 w-1/3"></div>
                    <div className="h-3 bg-gray-200 rounded mb-1 w-3/4"></div>
                    <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                  </div>
                  <div className="ml-4">
                    <div className="h-6 bg-gray-200 rounded w-16"></div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardBody>
      </Card>
    );
  }

  if (reports.length === 0) {
    return (
      <Card className="h-[480px] flex flex-col">
        <CardHeader>
          <CardTitle>Recent Reports</CardTitle>
        </CardHeader>
        <CardBody className="flex-1 overflow-hidden">
          <div className="text-center h-full flex flex-col justify-center">
            <ClockIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-base font-medium text-gray-900 mb-2">No reports yet</h3>
            <p className="text-sm text-gray-600">Reports will appear here once pod issues are detected</p>
          </div>
        </CardBody>
      </Card>
    );
  }

  return (
    <Card className="h-[480px] flex flex-col">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Recent Reports</CardTitle>
          <Link to={ROUTES.REPORTS}>
            <Button variant="ghost" size="sm">
              View All
            </Button>
          </Link>
        </div>
      </CardHeader>
      <CardBody className="flex-1 overflow-hidden">
        <div className="space-y-4 h-full">
          {reports
            .sort((a, b) => new Date(b.metadata.creationTimestamp).getTime() - new Date(a.metadata.creationTimestamp).getTime())
            .slice(0, 3)
            .map((report) => (
            <div key={report.metadata.name} className="flex items-center justify-between py-3 border-b border-gray-100 last:border-b-0">
              <div className="flex-1 min-w-0">
                <div className="flex items-center space-x-3 mb-2">
                  <h4 className="text-sm font-medium text-gray-900 truncate">
                    {report.spec.targetPod.name}
                  </h4>
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
                <p className="text-sm text-gray-600 mb-1">
                  {truncateText(report.spec.analysis.summary, 100)}
                </p>
                <div className="flex items-center text-xs text-gray-500">
                  <ClockIcon className="h-3 w-3 mr-1" />
                  {formatTimeAgo(report.metadata.creationTimestamp)}
                  <span className="mx-2">•</span>
                  <span>LLM: {report.spec.analysis.provider} {report.spec.analysis.model}</span>
                  <span className="mx-2">•</span>
                  <span>{report.spec.analysis.processingTime}</span>
                </div>
              </div>
              <div className="ml-4 flex-shrink-0">
                <Link to={`${ROUTES.REPORTS}/${report.metadata.name}`}>
                  <Button variant="ghost" size="sm">
                    <EyeIcon className="h-4 w-4" />
                  </Button>
                </Link>
              </div>
            </div>
          ))}
        </div>
      </CardBody>
    </Card>
  );
};

export default RecentReports;