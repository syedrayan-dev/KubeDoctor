import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeftIcon, ExclamationTriangleIcon, EyeIcon, ClockIcon } from '@heroicons/react/24/outline';
import Layout from '../../shared/components/layout/Layout';
import ReportDetail from './components/ReportDetail';
import Button from '../../shared/components/ui/Button';
import { ROUTES } from '../../shared/constants/routes';
import { useReportDetail } from './hooks/useReportDetail';
import { formatDistanceToNow } from 'date-fns';
import Card, { CardBody } from '../../shared/components/ui/Card';
import Badge from '../../shared/components/ui/Badge';

const ReportDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { report, request, loading, error, refetch } = useReportDetail(id || '');
  const [showRequest, setShowRequest] = useState(false);

  const handleBack = () => {
    navigate(ROUTES.REPORTS);
  };

  if (error) {
    return (
      <Layout 
        title="Report Not Found" 
        subtitle="The requested diagnosis report could not be found"
      >
        <div className="text-center py-12">
          <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">Report not found</h3>
          <p className="text-gray-600 mb-4">{error}</p>
          <div className="space-x-2">
            <Button onClick={handleBack} variant="outline">
              Back to Reports
            </Button>
            <Button onClick={refetch}>
              Try Again
            </Button>
          </div>
        </div>
      </Layout>
    );
  }

  const reportTitle = report 
    ? `Report: ${report.spec.targetPod.name}`
    : 'Loading Report...';
  
  const reportSubtitle = report
    ? `Diagnosis report for ${report.spec.targetPod.name} in ${report.spec.targetPod.namespace}`
    : 'Loading diagnosis report details...';

  return (
    <Layout title={reportTitle} subtitle={reportSubtitle}>
      <div className="space-y-6">
        {/* Navigation */}
        <div className="flex items-center justify-between">
          <Button variant="ghost" size="sm" onClick={handleBack}>
            <ArrowLeftIcon className="h-4 w-4 mr-2" />
            Back to Reports
          </Button>
          
          <div className="flex items-center space-x-3">
            {request && (
              <Button 
                variant="outline" 
                size="sm" 
                onClick={() => setShowRequest(!showRequest)}
              >
                <EyeIcon className="h-4 w-4 mr-2" />
                {showRequest ? 'Hide Request' : 'View Request'}
              </Button>
            )}
            {report && (
              <div className="flex items-center space-x-2 text-sm text-gray-500">
                <span>Report ID:</span>
                <code className="bg-gray-100 px-2 py-1 rounded text-xs">
                  {report.metadata.name}
                </code>
              </div>
            )}
          </div>
        </div>

        {/* Request Details (when visible) */}
        {showRequest && request && (
          <Card>
            <CardBody className="p-6">
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-gray-900">Request Details</h3>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <span className="text-sm text-gray-500 block mb-1">Request ID</span>
                    <code className="text-sm font-mono bg-gray-100 px-2 py-1 rounded">
                      {request.metadata.name}
                    </code>
                  </div>
                  <div>
                    <span className="text-sm text-gray-500 block mb-1">Status</span>
                    <Badge variant="status" status={request.status.phase as any}>
                      {request.status.phase}
                    </Badge>
                  </div>
                  <div>
                    <span className="text-sm text-gray-500 block mb-1">Type</span>
                    <span className="text-sm">{request.spec.type}</span>
                  </div>
                  <div>
                    <span className="text-sm text-gray-500 block mb-1">Policy</span>
                    <span className="text-sm">{request.spec.policyRef.name}</span>
                  </div>
                  <div>
                    <span className="text-sm text-gray-500 block mb-1">Created</span>
                    <span className="text-sm">
                      {formatDistanceToNow(new Date(request.metadata.creationTimestamp), { addSuffix: true })}
                    </span>
                  </div>
                  {request.status.completionTime && (
                    <div>
                      <span className="text-sm text-gray-500 block mb-1">Completed</span>
                      <span className="text-sm">
                        {formatDistanceToNow(new Date(request.status.completionTime), { addSuffix: true })}
                      </span>
                    </div>
                  )}
                </div>

                {request.status.message && (
                  <div className="mt-4 pt-4 border-t border-gray-200">
                    <span className="text-sm text-gray-500 block mb-1">Status Message</span>
                    <p className="text-sm text-gray-700">{request.status.message}</p>
                  </div>
                )}
              </div>
            </CardBody>
          </Card>
        )}

        {/* Report Detail */}
        {report && <ReportDetail report={report} loading={loading} />}
      </div>
    </Layout>
  );
};

export default ReportDetailPage;