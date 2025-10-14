import { apiClient } from '../api/client';
import type { Pod, DiagnosisPolicy, DiagnosisRequest } from '../types/api';

interface CreateManualRequestData {
  targetPod: {
    name: string;
    namespace: string;
  };
  policyRef: {
    name: string;
    namespace: string;
  };
  type: string;
}

export const diagnosisService = {
  // Get available namespaces
  async getNamespaces(): Promise<string[]> {
    const response = await apiClient.get('/api/v1/namespaces');
    return response.data.items.map((ns: any) => ns.metadata.name);
  },

  // Get pods in a namespace
  async getPods(namespace: string): Promise<Pod[]> {
    const response = await apiClient.get(`/api/v1/namespaces/${namespace}/pods`);
    return response.data.items;
  },

  // Search pods in a namespace with query
  async searchPods(namespace: string, query: string): Promise<Pod[]> {
    const response = await apiClient.get(`/api/v1/namespaces/${namespace}/pods`, {
      params: { search: query }
    });
    return response.data.items;
  },

  // Get diagnosis policies
  async getPolicies(): Promise<DiagnosisPolicy[]> {
    const response = await apiClient.get('/api/diagnosis/v1alpha1/policies');
    return response.data.items;
  },

  // Get manual diagnosis requests
  async getManualRequests(): Promise<DiagnosisRequest[]> {
    const response = await apiClient.get('/api/diagnosis/v1alpha1/requests', {
      params: { type: 'manual' }
    });
    return response.data.items.sort((a: DiagnosisRequest, b: DiagnosisRequest) => 
      new Date(b.metadata.creationTimestamp).getTime() - new Date(a.metadata.creationTimestamp).getTime()
    );
  },

  // Get specific diagnosis request
  async getRequest(requestId: string): Promise<DiagnosisRequest> {
    const response = await apiClient.get(`/api/diagnosis/v1alpha1/requests/${requestId}`);
    return response.data;
  },

  // Create manual diagnosis request
  async createManualRequest(data: CreateManualRequestData): Promise<DiagnosisRequest> {
    const requestBody = {
      apiVersion: 'diagnosis.apollo.io/v1alpha1',
      kind: 'DiagnosisRequest',
      metadata: {
        generateName: `${data.targetPod.name}-${data.type}-`,
        namespace: data.targetPod.namespace,
        labels: {
          'diagnosis.apollo.io/type': data.type,
          'diagnosis.apollo.io/target-pod': data.targetPod.name,
        },
      },
      spec: {
        targetPod: data.targetPod,
        policyRef: data.policyRef,
        type: data.type,
        manual: true,
      },
    };

    const response = await apiClient.post(
      `/api/diagnosis/v1alpha1/namespaces/${data.targetPod.namespace}/requests`,
      requestBody
    );
    return response.data;
  },

  // Delete diagnosis request
  async deleteRequest(requestId: string): Promise<void> {
    await apiClient.delete(`/api/diagnosis/v1alpha1/requests/${requestId}`);
  },

  // Cancel active diagnosis request
  async cancelRequest(requestId: string): Promise<DiagnosisRequest> {
    const response = await apiClient.patch(`/api/diagnosis/v1alpha1/requests/${requestId}`, {
      spec: {
        cancelled: true,
      },
    });
    return response.data;
  },
};