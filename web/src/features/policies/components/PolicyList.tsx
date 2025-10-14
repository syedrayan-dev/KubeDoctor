import React, { useState } from 'react';
import { 
  MagnifyingGlassIcon,
  PlusIcon,
  ExclamationTriangleIcon
} from '@heroicons/react/24/outline';
import Card, { CardHeader, CardTitle, CardBody } from '../../../shared/components/ui/Card';
import Button from '../../../shared/components/ui/Button';
import Select from '../../../shared/components/ui/Select';
import PolicyCard from './PolicyCard';
import type { DiagnosisPolicy } from '../../../shared/types/api';

interface PolicyListProps {
  policies: DiagnosisPolicy[];
  loading?: boolean;
  onCreateNew?: () => void;
  onEdit?: (policy: DiagnosisPolicy) => void;
  onDelete?: (policy: DiagnosisPolicy) => void;
  onView?: (policy: DiagnosisPolicy) => void;
}

interface PolicyFilters {
  search: string;
  namespace: string;
  provider: string;
}

const PolicyList: React.FC<PolicyListProps> = ({
  policies,
  loading = false,
  onCreateNew,
  onEdit,
  onDelete,
  onView,
}) => {
  const [filters, setFilters] = useState<PolicyFilters>({
    search: '',
    namespace: '',
    provider: '',
  });


  // Get unique values for filters
  const namespaces = Array.from(new Set(policies.map(p => p.metadata.namespace)));
  const providers = Array.from(new Set(policies.map(p => p.spec.llmConfig.provider)));

  // Filter policies
  const filteredPolicies = policies.filter(policy => {
    const matchesSearch = !filters.search || 
      policy.metadata.name.toLowerCase().includes(filters.search.toLowerCase()) ||
      policy.spec.targetNamespaces.some(ns => ns.toLowerCase().includes(filters.search.toLowerCase()));
    
    const matchesNamespace = !filters.namespace || policy.metadata.namespace === filters.namespace;
    const matchesProvider = !filters.provider || policy.spec.llmConfig.provider === filters.provider;

    return matchesSearch && matchesNamespace && matchesProvider;
  });

  const handleFilterChange = (key: keyof PolicyFilters, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }));
  };

  if (loading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="animate-pulse">
            <CardBody>
              <div className="space-y-3">
                <div className="h-4 bg-gray-200 rounded w-1/4"></div>
                <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                <div className="h-16 bg-gray-200 rounded"></div>
              </div>
            </CardBody>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Diagnosis Policies</CardTitle>
              <p className="text-sm text-gray-600 mt-1">
                Manage LLM configuration policies for pod diagnosis
              </p>
            </div>
            {onCreateNew && (
              <Button onClick={onCreateNew}>
                <PlusIcon className="h-4 w-4 mr-2" />
                Create Policy
              </Button>
            )}
          </div>
        </CardHeader>
        <CardBody>
          {/* Search and Filters */}
          <div className="flex items-center gap-4">
            {/* Search */}
            <div className="relative flex-1">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <MagnifyingGlassIcon className="h-5 w-5 text-gray-400" />
              </div>
              <input
                type="text"
                placeholder="Search by policy name..."
                value={filters.search}
                onChange={(e) => handleFilterChange('search', e.target.value)}
                className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg text-sm placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
              />
            </div>

            {/* Namespace Filter */}
            <Select
              options={[
                { value: '', label: 'All Namespaces' },
                ...namespaces.map(ns => ({ value: ns, label: ns }))
              ]}
              value={filters.namespace}
              onChange={(value) => handleFilterChange('namespace', value)}
              placeholder="All Namespaces"
              disabled={loading}
              className="w-48"
            />

            {/* Provider Filter */}
            <Select
              options={[
                { value: '', label: 'All Providers' },
                ...providers.map(provider => ({ value: provider, label: provider }))
              ]}
              value={filters.provider}
              onChange={(value) => handleFilterChange('provider', value)}
              placeholder="All Providers"
              disabled={loading}
              className="w-48"
            />
          </div>
        </CardBody>
      </Card>

      {/* Results Summary */}
      <div className="flex items-center justify-between">
        <div className="text-sm text-gray-600">
          {filteredPolicies.length === policies.length ? (
            `${policies.length} ${policies.length === 1 ? 'policy' : 'policies'} total`
          ) : (
            `${filteredPolicies.length} of ${policies.length} ${policies.length === 1 ? 'policy' : 'policies'}`
          )}
        </div>
      </div>

      {/* Policy List */}
      {filteredPolicies.length === 0 ? (
        <Card>
          <CardBody>
            <div className="text-center py-8">
              {policies.length === 0 ? (
                <>
                  <ExclamationTriangleIcon className="h-8 w-8 text-gray-400 mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">No policies found</h3>
                  <p className="text-gray-600 mb-4">
                    Get started by creating your first diagnosis policy.
                  </p>
                  {onCreateNew && (
                    <Button onClick={onCreateNew}>
                      <PlusIcon className="h-4 w-4 mr-2" />
                      Create Policy
                    </Button>
                  )}
                </>
              ) : (
                <>
                  <MagnifyingGlassIcon className="h-8 w-8 text-gray-400 mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">No matching policies</h3>
                  <p className="text-gray-600 mb-4">
                    Try adjusting your search terms or filters.
                  </p>
                </>
              )}
            </div>
          </CardBody>
        </Card>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {filteredPolicies.map((policy) => (
            <PolicyCard
              key={policy.metadata.uid}
              policy={policy}
              onEdit={onEdit}
              onDelete={onDelete}
              onView={onView}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default PolicyList;