import React, { useState } from 'react';
import { 
  CogIcon, 
  GlobeAltIcon,
  KeyIcon,
  CalendarIcon,
  PencilIcon,
  TrashIcon,
  ExclamationCircleIcon,
  ClockIcon,
  ChevronDownIcon,
  ChevronUpIcon
} from '@heroicons/react/24/outline';
import Card, { CardBody } from '../../../shared/components/ui/Card';
import Button from '../../../shared/components/ui/Button';
import type { DiagnosisPolicy } from '../../../shared/types/api';
import { formatDistanceToNow } from 'date-fns';

interface PolicyCardProps {
  policy: DiagnosisPolicy;
  onEdit?: (policy: DiagnosisPolicy) => void;
  onDelete?: (policy: DiagnosisPolicy) => void;
  onView?: (policy: DiagnosisPolicy) => void;
}

const PolicyCard: React.FC<PolicyCardProps> = ({
  policy,
  onEdit,
  onDelete,
}) => {
  const [showTriggerConditions, setShowTriggerConditions] = useState(false);

  const formatTimeAgo = (timestamp: string) => {
    try {
      return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
    } catch {
      return 'Unknown';
    }
  };

  const getProviderIcon = (provider: string) => {
    switch (provider.toLowerCase()) {
      case 'openai':
        return <CogIcon className="h-4 w-4" />;
      case 'ollama':
        return <CogIcon className="h-4 w-4" />;
      default:
        return <CogIcon className="h-4 w-4" />;
    }
  };

  const getProviderColor = (provider: string) => {
    switch (provider.toLowerCase()) {
      case 'openai':
        return 'bg-white text-green-600 border border-green-300';
      case 'ollama':
        return 'bg-white text-purple-600 border border-purple-300';
      default:
        return 'bg-white text-gray-600 border border-gray-300';
    }
  };

  const getPodPhaseColor = (type: string) => {
    switch (type) {
      case 'Failed':
        return 'bg-white text-red-600 border-red-300';
      case 'Pending':
        return 'bg-white text-yellow-600 border-yellow-300';
      case 'Unknown':
        return 'bg-white text-gray-600 border-gray-300';
      case 'Running':
        return 'bg-white text-blue-600 border-blue-300';
      default:
        return 'bg-white text-gray-600 border-gray-300';
    }
  };

  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardBody>
        <div className="space-y-4">
          {/* Header */}
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <div className="flex items-center space-x-2 mb-2">
                <h3 className="text-lg font-semibold text-gray-900">
                  {policy.metadata.name}
                </h3>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-white text-gray-600 border border-gray-300">
                  {policy.metadata.namespace}
                </span>
              </div>
              <p className="text-sm text-gray-600 mb-3">
                Created {formatTimeAgo(policy.metadata.creationTimestamp)}
              </p>
            </div>
            <div className="flex space-x-1">
              {onEdit && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onEdit(policy)}
                  title="Edit policy"
                >
                  <PencilIcon className="h-4 w-4" />
                </Button>
              )}
              {onDelete && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onDelete(policy)}
                  title="Delete policy"
                  className="text-red-600 hover:text-red-700 hover:bg-red-50"
                >
                  <TrashIcon className="h-4 w-4" />
                </Button>
              )}
            </div>
          </div>

          {/* Trigger Conditions */}
          {policy.spec.triggerConditions && policy.spec.triggerConditions.length > 0 && (
            <div>
              <div 
                className="flex items-center justify-between mb-2 cursor-pointer"
                onClick={() => setShowTriggerConditions(!showTriggerConditions)}
              >
                <div className="flex items-center space-x-2">
                  <ExclamationCircleIcon className="h-4 w-4 text-gray-500" />
                  <span className="text-sm font-medium text-gray-700">Trigger Conditions</span>
                  <span className="text-xs text-gray-500">
                    ({(() => {
                      let count = 0;
                      policy.spec.triggerConditions?.forEach(tc => {
                        count++; // Count the phase trigger itself
                        if (tc.conditions && tc.conditions.length > 0) {
                          count += tc.conditions.length; // Count pod conditions
                        }
                      });
                      return count;
                    })()} condition{(() => {
                      let count = 0;
                      policy.spec.triggerConditions?.forEach(tc => {
                        count++;
                        if (tc.conditions && tc.conditions.length > 0) {
                          count += tc.conditions.length;
                        }
                      });
                      return count > 1 ? 's' : '';
                    })()})
                  </span>
                </div>
                {showTriggerConditions ? (
                  <ChevronUpIcon className="h-4 w-4 text-gray-400" />
                ) : (
                  <ChevronDownIcon className="h-4 w-4 text-gray-400" />
                )}
              </div>
              {showTriggerConditions && (
                <div className="space-y-3">
                  {(() => {
                  // Group conditions by type
                  const groupedConditions = policy.spec.triggerConditions.reduce((acc, condition) => {
                    if (!acc[condition.type]) {
                      acc[condition.type] = [];
                    }
                    acc[condition.type].push(condition);
                    return acc;
                  }, {} as Record<string, typeof policy.spec.triggerConditions>);


                  // Define the order of types
                  const typeOrder = ['Failed', 'Pending', 'Unknown', 'Running'];
                  
                  return typeOrder
                    .filter(type => groupedConditions[type])
                    .map((type) => (
                      <div key={type} className="bg-gray-50 rounded-lg p-3 border border-gray-200">
                        <div className={`flex items-center justify-between ${
                          groupedConditions[type].some(cond => cond.conditions && cond.conditions.length > 0) ? 'mb-2' : ''
                        }`}>
                          <div className="flex items-center space-x-2">
                            <span 
                              className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                                getPodPhaseColor(type)
                              }`}
                            >
                              {type}
                            </span>
                          </div>
                          {/* Show minDuration at phase level if exists */}
                          {groupedConditions[type].some(cond => cond.minDuration) && (
                            <div className="flex items-center text-xs text-gray-500">
                              <ClockIcon className="h-3 w-3 mr-1" />
                              {groupedConditions[type].find(cond => cond.minDuration)?.minDuration}
                            </div>
                          )}
                        </div>
                        {/* Only show conditions section if there are specific pod conditions */}
                        {groupedConditions[type].some(cond => cond.conditions && cond.conditions.length > 0) && (
                          <div className="space-y-2 mt-2">
                            {groupedConditions[type].map((condition, condIndex) => {
                              const hasSpecificConditions = condition.conditions && condition.conditions.length > 0;
                              
                              if (!hasSpecificConditions) {
                                return null;
                              }
                              
                              return (
                                <div key={condIndex} className="space-y-1.5">
                                  {condition.conditions?.map((podCond, pcIndex) => {
                                    // Use pod condition's minDuration if exists, otherwise use trigger condition's minDuration
                                    const displayMinDuration = podCond.minDuration || condition.minDuration;
                                    return (
                                    <div key={pcIndex} className="flex items-center justify-between bg-white rounded-md p-1.5">
                                      <div className="flex items-center gap-2">
                                        <span className="text-xs text-gray-700">{podCond.name}</span>
                                        <span className="text-xs text-gray-400">:</span>
                                        <span 
                                          className={`text-xs ${
                                            podCond.status === 'True' 
                                              ? 'text-green-600' 
                                              : podCond.status === 'False' 
                                              ? 'text-red-600' 
                                              : 'text-gray-600'
                                          }`}
                                        >
                                          {podCond.status}
                                        </span>
                                      </div>
                                      {displayMinDuration && (
                                        <div className="flex items-center gap-1">
                                          <ClockIcon className="h-3 w-3 text-gray-400" />
                                          <span className="text-xs text-gray-500 tabular-nums">
                                            {displayMinDuration}
                                          </span>
                                        </div>
                                      )}
                                    </div>
                                    );
                                  })}
                                </div>
                              );
                            })}
                          </div>
                        )}
                      </div>
                    ));
                  })()}
                </div>
              )}
            </div>
          )}

          {/* Target Namespaces */}
          <div>
            <div className="flex items-center space-x-2 mb-2">
              <GlobeAltIcon className="h-4 w-4 text-gray-500" />
              <span className="text-sm font-medium text-gray-700">Target Namespaces</span>
            </div>
            <div className="flex flex-wrap gap-1">
              {policy.spec.targetNamespaces.length > 0 ? (
                policy.spec.targetNamespaces.map((namespace) => (
                  <span 
                    key={namespace}
                    className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-white text-blue-600 border border-blue-300"
                  >
                    {namespace}
                  </span>
                ))
              ) : (
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-white text-blue-600 border border-blue-300">
                  Same as policy namespace
                </span>
              )}
            </div>
          </div>

          {/* LLM Configuration */}
          <div>
            <div className="flex items-center space-x-2 mb-2">
              <CogIcon className="h-4 w-4 text-gray-500" />
              <span className="text-sm font-medium text-gray-700">LLM Configuration</span>
            </div>
            <div className="bg-gray-50 rounded-lg p-3 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">Provider:</span>
                <div className="flex items-center space-x-2">
                  {getProviderIcon(policy.spec.llmConfig.provider)}
                  <span className={`inline-flex items-center px-2 py-0.5 text-xs font-medium rounded-full ${getProviderColor(policy.spec.llmConfig.provider)}`}>
                    {policy.spec.llmConfig.provider}
                  </span>
                </div>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">Model:</span>
                <span className="text-sm font-medium text-gray-900">
                  {policy.spec.llmConfig.model}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">Secret:</span>
                <div className="flex items-center space-x-1">
                  <KeyIcon className="h-3 w-3 text-gray-400" />
                  <span className="text-sm text-gray-700">
                    {policy.spec.llmConfig.apiKeySecretRef?.name || 'N/A'}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Metadata */}
          <div className="pt-3 border-t border-gray-200">
            <div className="flex items-center justify-between text-sm text-gray-500">
              <div className="flex items-center space-x-1">
                <CalendarIcon className="h-4 w-4" />
                <span>Created: {new Date(policy.metadata.creationTimestamp).toLocaleDateString()}</span>
              </div>
              <div>
                UID: {policy.metadata.uid?.substring(0, 8) || 'N/A'}...
              </div>
            </div>
          </div>
        </div>
      </CardBody>
    </Card>
  );
};

export default PolicyCard;