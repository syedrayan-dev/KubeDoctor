import React, { useState } from 'react';
import { 
  ClockIcon,
  CpuChipIcon,
  ServerIcon,
  LightBulbIcon,
  ChartBarIcon,
  DocumentTextIcon,
  CalendarIcon
} from '@heroicons/react/24/outline';
import Card, { CardHeader, CardTitle, CardBody } from '../../../shared/components/ui/Card';
import Badge from '../../../shared/components/ui/Badge';
import { PROBLEM_TYPE_LABELS } from '../../../shared/constants/status';
import { PROBLEM_TYPE_BADGE_COLORS } from '../../../shared/constants/colors';
import type { DiagnosisReport } from '../../../shared/types/api';
import { formatDistanceToNow } from 'date-fns';

interface ReportDetailProps {
  report: DiagnosisReport;
  loading?: boolean;
}

const ReportDetail: React.FC<ReportDetailProps> = ({ report, loading = false }) => {
  const [activeTab, setActiveTab] = useState<'logs' | 'description'>('description');

  const formatTimeAgo = (timestamp: string) => {
    try {
      return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
    } catch {
      return 'Unknown';
    }
  };

  const formatDate = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleString();
    } catch {
      return 'Unknown';
    }
  };

  const getProblemTypeLabel = (type: string) => {
    return PROBLEM_TYPE_LABELS[type as keyof typeof PROBLEM_TYPE_LABELS] || type;
  };

  if (loading) {
    return (
      <div className="space-y-6">
        {[...Array(4)].map((_, index) => (
          <Card key={index}>
            <CardBody>
              <div className="animate-pulse">
                <div className="h-6 bg-gray-200 rounded mb-4 w-1/3"></div>
                <div className="space-y-2">
                  <div className="h-4 bg-gray-200 rounded"></div>
                  <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                  <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                </div>
              </div>
            </CardBody>
          </Card>
        ))}
      </div>
    );
  }


  return (
    <div className="space-y-6">
      {/* Header Information */}
      <Card>
        <CardBody>
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900 mb-2">
                {report.spec.targetPod.name}
              </h1>
              <div className="flex items-center space-x-3 mb-4">
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
            </div>
            <div className="text-right text-sm text-gray-500 min-w-0">
              <dl className="space-y-1">
                <div className="flex justify-between items-center gap-4">
                  <dt className="text-xs text-gray-400">Problem Detected</dt>
                  <dd className="text-xs text-gray-600">{formatDate(report.spec.triggerCondition.detectedAt)}</dd>
                </div>
                <div className="flex justify-between items-center gap-4">
                  <dt className="text-xs text-gray-400">Report Created</dt>
                  <dd className="text-xs text-gray-600">{formatDate(report.metadata.creationTimestamp)}</dd>
                </div>
                <div className="text-xs text-gray-500 mt-1">
                  ({formatTimeAgo(report.metadata.creationTimestamp)})
                </div>
              </dl>
            </div>
          </div>

          {/* Analysis - Main Content */}
          <div className="space-y-6">
            {/* Summary */}
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Summary</h3>
              <p className="text-base text-gray-800 leading-relaxed">{report.spec.analysis.summary}</p>
            </div>

            {/* Root Cause */}
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Root Cause</h3>
              <p className="text-base text-gray-800 leading-relaxed">{report.spec.analysis.rootCause}</p>
            </div>

            {/* Recommendations */}
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-3">
                Recommendations ({report.spec.analysis.recommendations.length})
              </h3>
              <div className="space-y-2">
                {report.spec.analysis.recommendations.map((recommendation, index) => (
                  <div key={index} className="flex items-start">
                    <span className="text-gray-400 mr-3 mt-1 text-sm">•</span>
                    <p className="text-base text-gray-800 leading-relaxed">{recommendation}</p>
                  </div>
                ))}
              </div>
            </div>

            {/* Analysis Metadata */}
            <div className="mt-8 pt-4 border-t border-gray-200">
              <div className="text-xs text-gray-500 space-y-1">
                <div className="flex justify-between">
                  <span>Policy:</span>
                  <span>{report.spec.policyRef.name}</span>
                </div>
                <div className="flex justify-between">
                  <span>LLM:</span>
                  <span className="capitalize">{report.spec.analysis.provider} {report.spec.analysis.model}</span>
                </div>
                <div className="flex justify-between">
                  <span>Processing Time:</span>
                  <span>{report.spec.analysis.processingTime}</span>
                </div>
              </div>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Pod Information */}
      <Card>
        <CardHeader>
          <div className="flex items-center border-b border-gray-200">
            <button
              onClick={() => setActiveTab('description')}
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'description'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Pod Description
            </button>
            <button
              onClick={() => setActiveTab('logs')}
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'logs'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Pod Logs ({report.spec.collectedData.logLines} lines)
            </button>
          </div>
        </CardHeader>
        <CardBody>
          <div className="relative h-96 overflow-hidden">
            {/* Pod Description */}
            <div 
              className={`absolute inset-0 transition-transform duration-300 ease-in-out ${
                activeTab === 'description' ? 'translate-x-0' : '-translate-x-full'
              }`}
            >
              <div className="bg-gray-50 rounded-lg p-4 h-full overflow-auto">
                <pre className="text-xs text-gray-600 whitespace-pre-wrap">
                  {report.spec.collectedData.podDescription}
                </pre>
              </div>
            </div>

            {/* Pod Logs */}
            <div 
              className={`absolute inset-0 transition-transform duration-300 ease-in-out ${
                activeTab === 'logs' ? 'translate-x-0' : 'translate-x-full'
              }`}
            >
              <div className="bg-gray-900 rounded-lg p-4 h-full overflow-auto">
                {report.spec.collectedData.logs && report.spec.collectedData.logs.trim() ? (
                  <pre className="text-sm text-gray-300 whitespace-pre-wrap">
                    {report.spec.collectedData.logs}
                  </pre>
                ) : (
                  <div className="flex items-center justify-center h-full">
                    <div className="text-center">
                      <ServerIcon className="h-12 w-12 text-gray-500 mx-auto mb-4" />
                      <p className="text-gray-400 text-sm">No logs available</p>
                      <p className="text-gray-500 text-xs mt-1">
                        Pod logs could not be collected or are empty
                      </p>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </CardBody>
      </Card>

    </div>
  );
};

export default ReportDetail;