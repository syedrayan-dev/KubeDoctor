import React from 'react';
import { MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import Select from '../../../shared/components/ui/Select';
import { PROBLEM_TYPES } from '../../../shared/constants/status';
import type { ReportFilters } from '../../../shared/types/api';

interface ReportFiltersProps {
  filters: ReportFilters;
  onFiltersChange: (filters: ReportFilters) => void;
  namespaces: string[];
  loading?: boolean;
}

const ReportFiltersComponent: React.FC<ReportFiltersProps> = ({
  filters,
  onFiltersChange,
  namespaces,
  loading = false,
}) => {

  const handleFilterChange = (key: keyof ReportFilters, value: string) => {
    onFiltersChange({
      ...filters,
      [key]: value || undefined,
    });
  };


  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <div className="flex items-center gap-4">
        {/* Search */}
        <div className="relative flex-1">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <MagnifyingGlassIcon className="h-5 w-5 text-gray-400" />
          </div>
          <input
            type="text"
            placeholder="Search by report name..."
            value={filters.search || ''}
            onChange={(e) => handleFilterChange('search', e.target.value)}
            className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg text-sm placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
            disabled={false}
          />
        </div>

        {/* Namespace Filter */}
        <Select
          options={[
            { value: '', label: 'All Namespaces' },
            ...namespaces.map(ns => ({ value: ns, label: ns }))
          ]}
          value={filters.namespace || ''}
          onChange={(value) => handleFilterChange('namespace', value)}
          placeholder="All Namespaces"
          disabled={loading}
          className="w-48"
        />

        {/* Problem Type Filter */}
        <Select
          options={[
            { value: '', label: 'All Problem Types' },
            ...Object.entries(PROBLEM_TYPES).map(([key, value]) => ({
              value,
              label: value
            }))
          ]}
          value={filters.problemType || ''}
          onChange={(value) => handleFilterChange('problemType', value)}
          placeholder="All Problem Types"
          disabled={loading}
          className="w-48"
        />
      </div>
    </div>
  );
};

export default ReportFiltersComponent;