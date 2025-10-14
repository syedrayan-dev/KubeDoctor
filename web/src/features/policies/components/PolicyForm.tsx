import React, { useState, useEffect } from 'react';
import { 
  PlusIcon,
  XMarkIcon,
  InformationCircleIcon,
  KeyIcon
} from '@heroicons/react/24/outline';
import Card, { CardHeader, CardTitle, CardBody } from '../../../shared/components/ui/Card';
import Button from '../../../shared/components/ui/Button';
import Badge from '../../../shared/components/ui/Badge';
import Select from '../../../shared/components/ui/Select';
import type { DiagnosisPolicy } from '../../../shared/types/api';

interface PodPhaseConfig {
  enabled: boolean;
  minDuration?: string;
  conditions: Array<{
    name: string;
    status: string;
    minDuration?: string;
  }>;
}

interface PolicyFormData {
  name: string;
  namespace: string;
  targetNamespaces: string[];
  phases: {
    Failed: PodPhaseConfig;
    Pending: PodPhaseConfig;
    Unknown: PodPhaseConfig;
    Running: PodPhaseConfig;
  };
  llmProvider: string;
  llmModel: string;
  secretName: string;
  secretKey: string;
  baseURL?: string;
}

interface PolicyFormSubmitData extends Omit<PolicyFormData, 'phases'> {
  triggerConditions: Array<{
    type: string;
    minDuration?: string;
    conditions?: Array<{
      name: string;
      status: string;
      minDuration?: string;
    }>;
  }>;
}

interface PolicyFormProps {
  policy?: DiagnosisPolicy;
  namespaces: string[];
  onSubmit: (data: PolicyFormSubmitData) => void;
  onCancel: () => void;
  loading?: boolean;
}

const PolicyForm: React.FC<PolicyFormProps> = ({
  policy,
  namespaces,
  onSubmit,
  onCancel,
  loading = false,
}) => {
  const [formData, setFormData] = useState<PolicyFormData>({
    name: '',
    namespace: '',
    targetNamespaces: [],
    phases: {
      Failed: { enabled: true, minDuration: '0s', conditions: [] },
      Pending: { enabled: false, minDuration: '30s', conditions: [] },
      Unknown: { enabled: false, minDuration: '5m', conditions: [] },
      Running: { enabled: false, minDuration: '5m', conditions: [] }
    },
    llmProvider: 'openai',
    llmModel: 'gpt-4o',
    secretName: '',
    secretKey: '',
    baseURL: '',
  });

  const [newNamespace, setNewNamespace] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (policy) {
      // Convert triggerConditions to phases format
      const phases: {
        Failed: PodPhaseConfig;
        Pending: PodPhaseConfig;
        Unknown: PodPhaseConfig;
        Running: PodPhaseConfig;
      } = {
        Failed: { enabled: false, minDuration: undefined, conditions: [] },
        Pending: { enabled: false, minDuration: undefined, conditions: [] },
        Unknown: { enabled: false, minDuration: undefined, conditions: [] },
        Running: { enabled: false, minDuration: undefined, conditions: [] }
      };

      // Convert existing triggerConditions to the new phases format
      policy.spec.triggerConditions?.forEach(tc => {
        const phaseKey = tc.type as keyof typeof phases;
        if (phases[phaseKey]) {
          phases[phaseKey] = {
            enabled: true,
            minDuration: tc.minDuration,
            conditions: tc.conditions?.map(cond => ({
              name: cond.name,
              status: cond.status,
              minDuration: cond.minDuration
            })) || []
          };
        }
      });

      setFormData({
        name: policy.metadata.name,
        namespace: policy.metadata.namespace,
        targetNamespaces: policy.spec.targetNamespaces || [],
        phases,
        llmProvider: policy.spec.llmConfig.provider,
        llmModel: policy.spec.llmConfig.model,
        secretName: policy.spec.llmConfig.apiKeySecretRef?.name || '',
        secretKey: policy.spec.llmConfig.apiKeySecretRef?.key || '',
        baseURL: policy.spec.llmConfig.baseURL || '',
      });
    }
  }, [policy]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Policy name is required';
    } else if (!/^[a-z0-9-]+$/.test(formData.name)) {
      newErrors.name = 'Policy name must be lowercase letters, numbers, and hyphens only';
    }

    if (!formData.namespace) {
      newErrors.namespace = 'Namespace is required';
    }

    if (!formData.llmProvider) {
      newErrors.llmProvider = 'LLM provider is required';
    }

    if (!formData.llmModel.trim()) {
      newErrors.llmModel = 'LLM model is required';
    }

    // Secret is only required for OpenAI
    if (formData.llmProvider === 'openai') {
      if (!formData.secretName.trim()) {
        newErrors.secretName = 'Secret name is required for OpenAI';
      }

      if (!formData.secretKey.trim()) {
        newErrors.secretKey = 'Secret key is required for OpenAI';
      }
    }

    // Check if at least one phase is enabled
    const enabledPhases = Object.values(formData.phases).filter(phase => phase.enabled);
    if (enabledPhases.length === 0) {
      newErrors.phases = 'At least one trigger condition phase must be enabled';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      // Convert phases back to triggerConditions format
      const triggerConditions = Object.entries(formData.phases)
        .filter(([_, config]) => config.enabled)
        .map(([phase, config]) => ({
          type: phase,
          minDuration: config.minDuration,
          conditions: config.conditions.length > 0 ? config.conditions : undefined
        }));

      const { phases, ...formDataWithoutPhases } = formData;
      const submitData: PolicyFormSubmitData = {
        ...formDataWithoutPhases,
        triggerConditions
      };
      
      onSubmit(submitData);
    }
  };

  const handleAddNamespace = () => {
    if (newNamespace && !formData.targetNamespaces.includes(newNamespace)) {
      setFormData(prev => ({
        ...prev,
        targetNamespaces: [...prev.targetNamespaces, newNamespace]
      }));
      setNewNamespace('');
    }
  };

  const handleRemoveNamespace = (namespace: string) => {
    setFormData(prev => ({
      ...prev,
      targetNamespaces: prev.targetNamespaces.filter(ns => ns !== namespace)
    }));
  };

  const llmProviders = [
    { value: 'openai', label: 'OpenAI' },
    { value: 'ollama', label: 'Ollama' },
  ];

  const modelsByProvider: Record<string, string[]> = {
    openai: [
      'gpt-4o',
      'gpt-4o-mini', 
      'gpt-4',
      'gpt-4-turbo',
      'gpt-3.5-turbo',
    ],
    ollama: [
      'llama3.2',
      'llama3.1',
      'llama3',
      'llama2',
      'mistral',
      'mixtral',
      'gemma',
      'codellama',
    ]
  };

  const currentModels = modelsByProvider[formData.llmProvider] || [];

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {policy ? 'Edit Policy' : 'Create New Policy'}
        </CardTitle>
      </CardHeader>
      <CardBody>
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Basic Information */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Policy Name *
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                disabled={!!policy || loading}
                className={`block w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500 ${
                  errors.name ? 'border-red-300' : 'border-gray-300'
                } ${(!!policy || loading) ? 'bg-gray-50 cursor-not-allowed' : ''}`}
                placeholder="my-diagnosis-policy"
              />
              {errors.name && (
                <p className="mt-1 text-sm text-red-600">{errors.name}</p>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Namespace *
              </label>
              {!!policy ? (
                <input
                  type="text"
                  value={formData.namespace}
                  disabled
                  className="block w-full px-3 py-2 border border-gray-300 rounded-lg text-sm bg-gray-50 cursor-not-allowed"
                />
              ) : (
                <>
                  <Select
                    value={formData.namespace}
                    onChange={(value) => setFormData(prev => ({ ...prev, namespace: value }))}
                    disabled={loading}
                    options={[
                      { value: '', label: 'Select namespace' },
                      ...namespaces.map((ns) => ({ value: ns, label: ns }))
                    ]}
                    placeholder="Select namespace"
                  />
                  {errors.namespace && (
                    <p className="mt-1 text-sm text-red-600">{errors.namespace}</p>
                  )}
                </>
              )}
            </div>
          </div>

          {/* Target Namespaces */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Target Namespaces
            </label>
            <div className="space-y-3">
              <div className="flex space-x-2">
                <Select
                  value={newNamespace}
                  onChange={setNewNamespace}
                  disabled={loading}
                  options={[
                    { value: '', label: 'Select namespace to add' },
                    ...namespaces
                      .filter(ns => !formData.targetNamespaces.includes(ns))
                      .map((ns) => ({ value: ns, label: ns }))
                  ]}
                  placeholder="Select namespace to add"
                  className="flex-1"
                />
                <Button
                  type="button"
                  onClick={handleAddNamespace}
                  disabled={!newNamespace || loading}
                  size="sm"
                >
                  <PlusIcon className="h-4 w-4 mr-1" />
                  Add
                </Button>
              </div>

              {formData.targetNamespaces.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {formData.targetNamespaces.map((namespace) => (
                    <div key={namespace} className="flex items-center space-x-1">
                      <Badge variant="outline">
                        {namespace}
                      </Badge>
                      <button
                        type="button"
                        onClick={() => handleRemoveNamespace(namespace)}
                        disabled={loading}
                        className="text-red-500 hover:text-red-700 disabled:opacity-50"
                      >
                        <XMarkIcon className="h-3 w-3" />
                      </button>
                    </div>
                  ))}
                </div>
              )}

              <div className="flex items-start space-x-2 text-sm text-gray-600 bg-blue-50 p-3 rounded-lg">
                <InformationCircleIcon className="h-4 w-4 text-blue-500 mt-0.5 flex-shrink-0" />
                <p>
                  If no target namespaces are specified, the policy will only apply to the same namespace where it's created.
                </p>
              </div>
            </div>
          </div>

          {/* Trigger Conditions - Phase Based */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium text-gray-900">Trigger Conditions</h3>
            
            {errors.phases && (
              <p className="text-sm text-red-600">{errors.phases}</p>
            )}
            
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              {Object.entries(formData.phases).map(([phase, config]) => {
                const phaseColors = {
                  Failed: 'border-red-200 bg-red-50',
                  Pending: 'border-yellow-200 bg-yellow-50', 
                  Unknown: 'border-gray-200 bg-gray-50',
                  Running: 'border-blue-200 bg-blue-50'
                };

                const phaseTextColors = {
                  Failed: 'text-red-800',
                  Pending: 'text-yellow-800',
                  Unknown: 'text-gray-800', 
                  Running: 'text-blue-800'
                };

                return (
                  <div key={phase} className={`border rounded-lg p-4 ${config.enabled ? phaseColors[phase as keyof typeof phaseColors] : 'border-gray-200 bg-gray-50'}`}>
                    {/* Phase Header with Toggle */}
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center space-x-3">
                        <input
                          type="checkbox"
                          checked={config.enabled}
                          onChange={(e) => {
                            setFormData(prev => ({
                              ...prev,
                              phases: {
                                ...prev.phases,
                                [phase]: {
                                  ...prev.phases[phase as keyof typeof prev.phases],
                                  enabled: e.target.checked
                                }
                              }
                            }));
                          }}
                          disabled={loading}
                          className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                        />
                        <h4 className={`text-sm font-medium ${config.enabled ? phaseTextColors[phase as keyof typeof phaseTextColors] : 'text-gray-500'}`}>
                          {phase} Phase
                        </h4>
                      </div>
                    </div>

                    {/* Phase Configuration */}
                    {config.enabled && (
                      <div className="space-y-3">
                        {/* Min Duration */}
                        <div>
                          <label className="block text-xs font-medium text-gray-700 mb-1">
                            Min Duration
                          </label>
                          <input
                            type="text"
                            value={config.minDuration || ''}
                            onChange={(e) => {
                              setFormData(prev => ({
                                ...prev,
                                phases: {
                                  ...prev.phases,
                                  [phase]: {
                                    ...prev.phases[phase as keyof typeof prev.phases],
                                    minDuration: e.target.value || undefined
                                  }
                                }
                              }));
                            }}
                            disabled={loading}
                            className="block w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                            placeholder="5m"
                          />
                        </div>

                        {/* Pod Conditions */}
                        <div className="bg-white rounded p-3 border border-gray-200">
                          <div className="flex items-center justify-between mb-2">
                            <h5 className="text-xs font-medium text-gray-700">Pod Conditions</h5>
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              onClick={() => {
                                setFormData(prev => ({
                                  ...prev,
                                  phases: {
                                    ...prev.phases,
                                    [phase]: {
                                      ...prev.phases[phase as keyof typeof prev.phases],
                                      conditions: [
                                        ...prev.phases[phase as keyof typeof prev.phases].conditions,
                                        { name: 'Ready', status: 'False', minDuration: undefined }
                                      ]
                                    }
                                  }
                                }));
                              }}
                              disabled={loading}
                              className="text-xs"
                            >
                              <PlusIcon className="h-3 w-3 mr-1" />
                              Add
                            </Button>
                          </div>

                          {config.conditions.length > 0 ? (
                            <div className="space-y-2">
                              {config.conditions.map((condition, condIndex) => (
                                <div key={condIndex} className="bg-gray-50 rounded p-2">
                                  <div className="grid grid-cols-3 gap-2 mb-2">
                                    <Select
                                      value={condition.name}
                                      onChange={(value) => {
                                        setFormData(prev => ({
                                          ...prev,
                                          phases: {
                                            ...prev.phases,
                                            [phase]: {
                                              ...prev.phases[phase as keyof typeof prev.phases],
                                              conditions: prev.phases[phase as keyof typeof prev.phases].conditions.map((c, i) =>
                                                i === condIndex ? { ...c, name: value } : c
                                              )
                                            }
                                          }
                                        }));
                                      }}
                                      disabled={loading}
                                      options={[
                                        { value: 'Ready', label: 'Ready' },
                                        { value: 'ContainersReady', label: 'ContainersReady' },
                                        { value: 'PodScheduled', label: 'PodScheduled' },
                                        { value: 'Initialized', label: 'Initialized' },
                                        { value: 'PodReadyToStartContainers', label: 'PodReadyToStartContainers' },
                                        { value: 'DisruptionTarget', label: 'DisruptionTarget' }
                                      ]}
                                      className="text-xs"
                                    />
                                    
                                    <Select
                                      value={condition.status}
                                      onChange={(value) => {
                                        setFormData(prev => ({
                                          ...prev,
                                          phases: {
                                            ...prev.phases,
                                            [phase]: {
                                              ...prev.phases[phase as keyof typeof prev.phases],
                                              conditions: prev.phases[phase as keyof typeof prev.phases].conditions.map((c, i) =>
                                                i === condIndex ? { ...c, status: value } : c
                                              )
                                            }
                                          }
                                        }));
                                      }}
                                      disabled={loading}
                                      options={[
                                        { value: 'True', label: 'True' },
                                        { value: 'False', label: 'False' },
                                        { value: 'Unknown', label: 'Unknown' }
                                      ]}
                                      className="text-xs"
                                    />
                                    
                                    <div className="flex items-center space-x-1">
                                      <input
                                        type="text"
                                        value={condition.minDuration || ''}
                                        onChange={(e) => {
                                          setFormData(prev => ({
                                            ...prev,
                                            phases: {
                                              ...prev.phases,
                                              [phase]: {
                                                ...prev.phases[phase as keyof typeof prev.phases],
                                                conditions: prev.phases[phase as keyof typeof prev.phases].conditions.map((c, i) =>
                                                  i === condIndex ? { ...c, minDuration: e.target.value || undefined } : c
                                                )
                                              }
                                            }
                                          }));
                                        }}
                                        disabled={loading}
                                        className="flex-1 text-xs px-1 py-1 border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-primary-500"
                                        placeholder={config.minDuration || '5m'}
                                      />
                                      <Button
                                        type="button"
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => {
                                          setFormData(prev => ({
                                            ...prev,
                                            phases: {
                                              ...prev.phases,
                                              [phase]: {
                                                ...prev.phases[phase as keyof typeof prev.phases],
                                                conditions: prev.phases[phase as keyof typeof prev.phases].conditions.filter((_, i) => i !== condIndex)
                                              }
                                            }
                                          }));
                                        }}
                                        disabled={loading}
                                        className="text-red-600 hover:text-red-700 p-0.5"
                                      >
                                        <XMarkIcon className="h-3 w-3" />
                                      </Button>
                                    </div>
                                  </div>
                                </div>
                              ))}
                            </div>
                          ) : (
                            <p className="text-xs text-gray-500 italic">
                              No conditions. Triggers on any {phase} pod.
                            </p>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
            
            <div className="flex items-start space-x-2 text-sm text-gray-600 bg-blue-50 p-3 rounded-lg">
              <InformationCircleIcon className="h-4 w-4 text-blue-500 mt-0.5 flex-shrink-0" />
              <div className="space-y-1">
                <p>
                  Enable phases to monitor specific pod states. Multiple enabled phases work as OR logic.
                </p>
                <p className="text-xs">
                  • Without conditions: Triggers when pod enters the phase<br/>
                  • With conditions: Triggers when pod is in the phase AND conditions are met
                </p>
              </div>
            </div>
          </div>

          {/* LLM Configuration */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium text-gray-900">LLM Configuration</h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Provider *
                </label>
                <Select
                  value={formData.llmProvider}
                  onChange={(newProvider) => {
                    setFormData(prev => ({ 
                      ...prev, 
                      llmProvider: newProvider,
                      llmModel: '' // Reset model when provider changes
                    }))
                  }}
                  disabled={loading}
                  options={llmProviders}
                  placeholder="Select provider"
                />
                {errors.llmProvider && (
                  <p className="mt-1 text-sm text-red-600">{errors.llmProvider}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Model *
                </label>
                <Select
                  value={formData.llmModel}
                  onChange={(value) => setFormData(prev => ({ ...prev, llmModel: value }))}
                  disabled={loading}
                  options={[
                    { value: '', label: 'Select model' },
                    ...currentModels.map((model) => ({ value: model, label: model }))
                  ]}
                  placeholder="Select model"
                />
                {errors.llmModel && (
                  <p className="mt-1 text-sm text-red-600">{errors.llmModel}</p>
                )}
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Secret Name {formData.llmProvider === 'openai' ? '*' : '(Optional)'}
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <KeyIcon className="h-4 w-4 text-gray-400" />
                  </div>
                  <input
                    type="text"
                    value={formData.secretName}
                    onChange={(e) => setFormData(prev => ({ ...prev, secretName: e.target.value }))}
                    disabled={loading}
                    className={`block w-full pl-10 pr-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500 ${
                      errors.secretName ? 'border-red-300' : 'border-gray-300'
                    }`}
                    placeholder="openai-secret"
                  />
                </div>
                {errors.secretName && (
                  <p className="mt-1 text-sm text-red-600">{errors.secretName}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Secret Key {formData.llmProvider === 'openai' ? '*' : '(Optional)'}
                </label>
                <input
                  type="text"
                  value={formData.secretKey}
                  onChange={(e) => setFormData(prev => ({ ...prev, secretKey: e.target.value }))}
                  disabled={loading}
                  className={`block w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500 ${
                    errors.secretKey ? 'border-red-300' : 'border-gray-300'
                  }`}
                  placeholder="api-key"
                />
                {errors.secretKey && (
                  <p className="mt-1 text-sm text-red-600">{errors.secretKey}</p>
                )}
              </div>
            </div>

            {formData.llmProvider === 'ollama' && (
              <div className="space-y-3">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Base URL
                  </label>
                  <input
                    type="url"
                    value={formData.baseURL || ''}
                    onChange={(e) => setFormData(prev => ({ ...prev, baseURL: e.target.value }))}
                    disabled={loading}
                    className="block w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                    placeholder="http://localhost:11434"
                  />
                  <p className="mt-1 text-xs text-gray-500">
                    Optional. Custom Ollama server URL. Leave empty to use default.
                  </p>
                </div>
                
                <div className="flex items-start space-x-2 text-sm text-gray-600 bg-blue-50 p-3 rounded-lg">
                  <InformationCircleIcon className="h-4 w-4 text-blue-500 mt-0.5 flex-shrink-0" />
                  <p>
                    Ollama doesn't require API keys. The secret fields above are optional and can be left empty.
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end space-x-3 pt-6 border-t border-gray-200">
            <Button
              type="button"
              variant="outline"
              onClick={onCancel}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={loading}
              disabled={loading}
            >
              {policy ? 'Update Policy' : 'Create Policy'}
            </Button>
          </div>
        </form>
      </CardBody>
    </Card>
  );
};

export default PolicyForm;