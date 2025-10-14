import { useState, useEffect, useCallback } from 'react';
import apiService from '../../../shared/services/apiService';
import type { DiagnosisPolicy } from '../../../shared/types/api';

interface UsePoliciesReturn {
  // Data
  policies: DiagnosisPolicy[];
  namespaces: string[];
  
  // Loading states
  loading: boolean;
  submitting: boolean;
  
  // Error states
  error: string | null;
  submitError: string | null;
  
  // Actions
  loadPolicies: () => Promise<void>;
  loadNamespaces: () => Promise<void>;
  createPolicy: (data: any) => Promise<DiagnosisPolicy | null>;
  updatePolicy: (policy: DiagnosisPolicy, data: any) => Promise<DiagnosisPolicy | null>;
  deletePolicy: (policy: DiagnosisPolicy) => Promise<void>;
  
  // Utility
  clearError: () => void;
  clearSubmitError: () => void;
}

export const usePolicies = (): UsePoliciesReturn => {
  // Data state
  const [policies, setPolicies] = useState<DiagnosisPolicy[]>([]);
  const [namespaces, setNamespaces] = useState<string[]>([]);

  // Loading states
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  // Error states
  const [error, setError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);

  // Load policies
  const loadPolicies = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiService.getPolicies();
      // Sort by creation timestamp (newest first)
      const sortedData = data.sort((a, b) => 
        new Date(b.metadata.creationTimestamp).getTime() - 
        new Date(a.metadata.creationTimestamp).getTime()
      );
      setPolicies(sortedData);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load policies';
      setError(message);
      console.error('Failed to load policies:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Load namespaces
  const loadNamespaces = useCallback(async () => {
    try {
      setError(null);
      const namespaces = await apiService.getNamespaces();
      const data = namespaces.map(ns => ns.metadata.name);
      setNamespaces(data);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load namespaces';
      setError(message);
      console.error('Failed to load namespaces:', err);
    }
  }, []);

  // Create policy
  const createPolicy = useCallback(async (data: any): Promise<DiagnosisPolicy | null> => {
    try {
      setSubmitting(true);
      setSubmitError(null);
      
      // Transform form data to API format
      const policyData = {
        metadata: {
          name: data.name,
          namespace: data.namespace,
        },
        spec: {
          targetNamespaces: data.targetNamespaces || [],
          triggerConditions: data.triggerConditions || [],
          llmConfig: {
            provider: data.llmProvider,
            model: data.llmModel,
            // Only include apiKeySecretRef for providers that require it
            ...(data.llmProvider === 'openai' && {
              apiKeySecretRef: {
                name: data.secretName,
                key: data.secretKey,
                namespace: data.namespace,
              }
            }),
            // Add baseURL for Ollama
            ...(data.llmProvider === 'ollama' && data.baseURL && { baseURL: data.baseURL })
          }
        }
      };
      
      const policy = await apiService.createPolicy(policyData);
      
      // Refresh the policies list
      await loadPolicies();
      
      return policy;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create policy';
      setSubmitError(message);
      console.error('Failed to create policy:', err);
      return null;
    } finally {
      setSubmitting(false);
    }
  }, [loadPolicies]);

  // Update policy
  const updatePolicy = useCallback(async (policy: DiagnosisPolicy, data: any): Promise<DiagnosisPolicy | null> => {
    try {
      setSubmitting(true);
      setSubmitError(null);
      
      // Transform form data to API format
      const updatedPolicyData: DiagnosisPolicy = {
        ...policy,
        metadata: {
          ...policy.metadata,
          name: data.name || policy.metadata.name,
          namespace: data.namespace || policy.metadata.namespace,
        },
        spec: {
          targetNamespaces: data.targetNamespaces !== undefined ? data.targetNamespaces : policy.spec.targetNamespaces || [],
          triggerConditions: data.triggerConditions !== undefined ? data.triggerConditions : policy.spec.triggerConditions || [],
          llmConfig: {
            provider: data.llmProvider || policy.spec.llmConfig.provider,
            model: data.llmModel || policy.spec.llmConfig.model,
            // Handle apiKeySecretRef based on provider
            ...((data.llmProvider || policy.spec.llmConfig.provider) === 'openai' && {
              apiKeySecretRef: {
                name: data.secretName || policy.spec.llmConfig.apiKeySecretRef?.name || '',
                key: data.secretKey || policy.spec.llmConfig.apiKeySecretRef?.key || 'api-key',
                namespace: data.namespace || policy.spec.llmConfig.apiKeySecretRef?.namespace || policy.metadata.namespace,
              }
            }),
            // Preserve or update baseURL for Ollama
            ...(policy.spec.llmConfig.baseURL && { baseURL: policy.spec.llmConfig.baseURL }),
            ...(data.llmProvider === 'ollama' && data.baseURL && { baseURL: data.baseURL })
          }
        }
      };
      
      const updatedPolicy = await apiService.updatePolicy(updatedPolicyData);
      
      // Refresh the entire policies list to ensure consistency
      await loadPolicies();
      
      return updatedPolicy;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update policy';
      setSubmitError(message);
      console.error('Failed to update policy:', err);
      return null;
    } finally {
      setSubmitting(false);
    }
  }, []);

  // Delete policy
  const deletePolicy = useCallback(async (policy: DiagnosisPolicy) => {
    try {
      setSubmitting(true);
      setError(null);
      
      await apiService.deletePolicy(policy.metadata.name, policy.metadata.namespace);
      
      // Remove the policy from local state
      setPolicies(prev => prev.filter(p => p.metadata.uid !== policy.metadata.uid));
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete policy';
      setError(message);
      console.error('Failed to delete policy:', err);
    } finally {
      setSubmitting(false);
    }
  }, []);

  // Clear errors
  const clearError = useCallback(() => setError(null), []);
  const clearSubmitError = useCallback(() => setSubmitError(null), []);

  // Initial data loading
  useEffect(() => {
    const loadInitialData = async () => {
      await Promise.all([
        loadPolicies(),
        loadNamespaces(),
      ]);
    };

    loadInitialData();
  }, [loadPolicies, loadNamespaces]);

  return {
    // Data
    policies,
    namespaces,
    
    // Loading states
    loading,
    submitting,
    
    // Error states
    error,
    submitError,
    
    // Actions
    loadPolicies,
    loadNamespaces,
    createPolicy,
    updatePolicy,
    deletePolicy,
    
    // Utility
    clearError,
    clearSubmitError,
  };
};