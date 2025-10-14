import React, { useState } from 'react';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import Layout from '../../shared/components/layout/Layout';
import PolicyList from './components/PolicyList';
import PolicyForm from './components/PolicyForm';
import Button from '../../shared/components/ui/Button';
import { ConfirmModal } from '../../shared/components/ui/Modal';
import { usePolicies } from './hooks/usePolicies';
import type { DiagnosisPolicy } from '../../shared/types/api';

type ViewMode = 'list' | 'create' | 'edit';

const PoliciesPage: React.FC = () => {
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [selectedPolicy, setSelectedPolicy] = useState<DiagnosisPolicy | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<DiagnosisPolicy | null>(null);

  const {
    policies,
    namespaces,
    loading,
    submitting,
    error,
    submitError,
    createPolicy,
    updatePolicy,
    deletePolicy,
    clearSubmitError,
  } = usePolicies();

  const handleCreateNew = () => {
    setSelectedPolicy(null);
    setViewMode('create');
    clearSubmitError();
  };

  const handleEdit = (policy: DiagnosisPolicy) => {
    setSelectedPolicy(policy);
    setViewMode('edit');
    clearSubmitError();
  };

  const handleView = (policy: DiagnosisPolicy) => {
    setSelectedPolicy(policy);
    setViewMode('edit');
    clearSubmitError();
  };

  const handleDelete = (policy: DiagnosisPolicy) => {
    setDeleteConfirm(policy);
  };

  const handleSubmit = async (data: any) => {
    let success = false;

    if (viewMode === 'create') {
      const result = await createPolicy(data);
      success = result !== null;
    } else if (viewMode === 'edit' && selectedPolicy) {
      const result = await updatePolicy(selectedPolicy, data);
      success = result !== null;
    }

    if (success) {
      setViewMode('list');
      setSelectedPolicy(null);
    }
  };

  const handleCancel = () => {
    setViewMode('list');
    setSelectedPolicy(null);
    clearSubmitError();
  };

  const confirmDelete = async () => {
    if (deleteConfirm) {
      await deletePolicy(deleteConfirm);
      setDeleteConfirm(null);
    }
  };

  const cancelDelete = () => {
    setDeleteConfirm(null);
  };

  if (error) {
    return (
      <Layout title="Diagnosis Policies" subtitle="Configure diagnosis policies">
        <div className="text-center py-12">
          <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">Failed to load policies</h3>
          <p className="text-gray-600 mb-4">{error}</p>
          <Button onClick={() => window.location.reload()}>
            Try Again
          </Button>
        </div>
      </Layout>
    );
  }

  return (
    <Layout 
      title="Diagnosis Policies" 
      subtitle={
        viewMode === 'list' 
          ? "Configure diagnosis policies" 
          : viewMode === 'create'
          ? "Create new diagnosis policy"
          : "Edit diagnosis policy"
      }
    >
      <div className="space-y-6">
        {/* Submit Error */}
        {submitError && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex items-start">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-400 mt-0.5 mr-3" />
              <div>
                <h3 className="text-sm font-medium text-red-800">
                  {viewMode === 'create' ? 'Failed to create policy' : 'Failed to update policy'}
                </h3>
                <p className="text-sm text-red-700 mt-1">{submitError}</p>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={clearSubmitError}
                className="ml-auto text-red-600 hover:text-red-700"
              >
                ×
              </Button>
            </div>
          </div>
        )}

        {/* Main Content */}
        {viewMode === 'list' ? (
          <PolicyList
            policies={policies}
            loading={loading}
            onCreateNew={handleCreateNew}
            onEdit={handleEdit}
            onDelete={handleDelete}
            onView={handleView}
          />
        ) : (
          <PolicyForm
            policy={selectedPolicy || undefined}
            namespaces={namespaces}
            onSubmit={handleSubmit}
            onCancel={handleCancel}
            loading={submitting}
          />
        )}

        {/* Delete Confirmation Modal */}
        <ConfirmModal
          isOpen={!!deleteConfirm}
          onClose={cancelDelete}
          onConfirm={confirmDelete}
          title="Delete Policy"
          message={`Are you sure you want to delete the policy "${deleteConfirm?.metadata.name}"? This action cannot be undone.`}
          confirmText="Delete Policy"
          cancelText="Cancel"
          confirmButtonClassName="bg-red-600 hover:bg-red-700 text-white"
          loading={submitting}
        />
      </div>
    </Layout>
  );
};

export default PoliciesPage;