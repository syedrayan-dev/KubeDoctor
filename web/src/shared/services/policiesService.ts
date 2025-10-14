import { apiClient } from '../api/client';
import type { DiagnosisPolicy } from '../types/api';

interface CreatePolicyData {
  name: string;
  namespace: string;
  targetNamespaces: string[];
  llmProvider: string;
  llmModel: string;
  secretName: string;
  secretKey: string;
}

export const policiesService = {
  // Get available namespaces
  async getNamespaces(): Promise<string[]> {
    const response = await apiClient.get('/api/v1/namespaces');
    return response.data.items.map((ns: any) => ns.metadata.name);
  },

  // Get diagnosis policies
  async getPolicies(): Promise<DiagnosisPolicy[]> {
    const response = await apiClient.get('/api/diagnosis/v1alpha1/policies');
    return response.data.items;
  },

  // Get specific policy
  async getPolicy(name: string, namespace: string): Promise<DiagnosisPolicy> {
    const response = await apiClient.get(`/api/diagnosis/v1alpha1/namespaces/${namespace}/policies/${name}`);
    return response.data;
  },

  // Create diagnosis policy
  async createPolicy(data: CreatePolicyData): Promise<DiagnosisPolicy> {
    const policyBody = {
      apiVersion: 'diagnosis.apollo.io/v1alpha1',
      kind: 'DiagnosisPolicy',
      metadata: {
        name: data.name,
        namespace: data.namespace,
        labels: {
          'diagnosis.apollo.io/managed-by': 'web-ui',
        },
      },
      spec: {
        targetNamespaces: data.targetNamespaces,
        llmConfig: {
          provider: data.llmProvider,
          model: data.llmModel,
          secretRef: {
            name: data.secretName,
            key: data.secretKey,
          },
        },
      },
    };

    const response = await apiClient.post(
      `/api/diagnosis/v1alpha1/namespaces/${data.namespace}/policies`,
      policyBody
    );
    return response.data;
  },

  // Update diagnosis policy
  async updatePolicy(name: string, namespace: string, data: Partial<CreatePolicyData>): Promise<DiagnosisPolicy> {
    const updates: any = {};

    if (data.targetNamespaces !== undefined) {
      updates.spec = { ...updates.spec, targetNamespaces: data.targetNamespaces };
    }

    if (data.llmProvider || data.llmModel || data.secretName || data.secretKey) {
      updates.spec = {
        ...updates.spec,
        llmConfig: {
          ...(data.llmProvider && { provider: data.llmProvider }),
          ...(data.llmModel && { model: data.llmModel }),
          ...(data.secretName || data.secretKey) && {
            secretRef: {
              ...(data.secretName && { name: data.secretName }),
              ...(data.secretKey && { key: data.secretKey }),
            },
          },
        },
      };
    }

    const response = await apiClient.patch(
      `/api/diagnosis/v1alpha1/namespaces/${namespace}/policies/${name}`,
      updates
    );
    return response.data;
  },

  // Delete diagnosis policy
  async deletePolicy(name: string, namespace: string): Promise<void> {
    await apiClient.delete(`/api/diagnosis/v1alpha1/namespaces/${namespace}/policies/${name}`);
  },

  // Validate policy configuration
  async validatePolicy(data: CreatePolicyData): Promise<{ valid: boolean; errors: string[] }> {
    try {
      const response = await apiClient.post('/api/diagnosis/v1alpha1/policies/validate', data);
      return response.data;
    } catch (err: any) {
      // If validation endpoint doesn't exist, do basic client-side validation
      const errors: string[] = [];

      if (!data.name || !/^[a-z0-9-]+$/.test(data.name)) {
        errors.push('Policy name must be lowercase letters, numbers, and hyphens only');
      }

      if (!data.namespace) {
        errors.push('Namespace is required');
      }

      if (!data.llmProvider) {
        errors.push('LLM provider is required');
      }

      if (!data.llmModel) {
        errors.push('LLM model is required');
      }

      if (!data.secretName) {
        errors.push('Secret name is required');
      }

      if (!data.secretKey) {
        errors.push('Secret key is required');
      }

      return {
        valid: errors.length === 0,
        errors,
      };
    }
  },
};