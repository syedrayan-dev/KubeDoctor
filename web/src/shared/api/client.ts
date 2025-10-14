import axios from 'axios';
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { API_BASE_URL, HTTP_STATUS } from '../constants/api';
import type { ApiResponse, PaginatedResponse } from '../types/api';

interface ApiError {
  message: string;
  code: string | number;
  details?: any;
}

// Create axios instance with default configuration
const createApiClient = (): AxiosInstance => {
  const client = axios.create({
    baseURL: API_BASE_URL,
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Request interceptor
  client.interceptors.request.use(
    (config) => {
      // Add auth token if available
      const token = localStorage.getItem('auth_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      
      // Log request in development
      if (import.meta.env.DEV) {
        console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`, config.data);
      }
      
      return config;
    },
    (error) => {
      return Promise.reject(error);
    }
  );

  // Response interceptor
  client.interceptors.response.use(
    (response: AxiosResponse) => {
      // Log response in development
      if (import.meta.env.DEV) {
        console.log(`[API] Response ${response.status}:`, response.data);
      }
      
      return response;
    },
    (error) => {
      // Handle common error cases
      const apiError: ApiError = {
        message: 'An error occurred',
        code: error.response?.status || 'UNKNOWN',
        details: error.response?.data,
      };

      if (error.response) {
        // Server responded with error status
        switch (error.response.status) {
          case HTTP_STATUS.BAD_REQUEST:
            apiError.message = 'Invalid request';
            break;
          case HTTP_STATUS.UNAUTHORIZED:
            apiError.message = 'Authentication required';
            // Redirect to login or clear auth token
            localStorage.removeItem('auth_token');
            break;
          case HTTP_STATUS.FORBIDDEN:
            apiError.message = 'Access denied';
            break;
          case HTTP_STATUS.NOT_FOUND:
            apiError.message = 'Resource not found';
            break;
          case HTTP_STATUS.INTERNAL_SERVER_ERROR:
            apiError.message = 'Server error';
            break;
          default:
            apiError.message = error.response.data?.message || 'Request failed';
        }
      } else if (error.request) {
        // Network error
        apiError.message = 'Network error - please check your connection';
        apiError.code = 'NETWORK_ERROR';
      } else {
        // Request setup error
        apiError.message = error.message || 'Request failed';
        apiError.code = 'REQUEST_ERROR';
      }

      if (import.meta.env.DEV) {
        console.error('[API] Error:', apiError);
      }

      return Promise.reject(apiError);
    }
  );

  return client;
};

// Create the API client instance
export const apiClient = createApiClient();

// Generic API methods
export class ApiService {
  private client: AxiosInstance;

  constructor(client: AxiosInstance) {
    this.client = client;
  }

  // GET request
  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<ApiResponse<T>>(url, config);
    return response.data.data;
  }

  // GET request with pagination
  async getList<T>(url: string, config?: AxiosRequestConfig): Promise<PaginatedResponse<T>> {
    const response = await this.client.get<PaginatedResponse<T>>(url, config);
    return response.data;
  }

  // POST request
  async post<T, D = any>(url: string, data?: D, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<ApiResponse<T>>(url, data, config);
    return response.data.data;
  }

  // PUT request
  async put<T, D = any>(url: string, data?: D, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<ApiResponse<T>>(url, data, config);
    return response.data.data;
  }

  // PATCH request
  async patch<T, D = any>(url: string, data?: D, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.patch<ApiResponse<T>>(url, data, config);
    return response.data.data;
  }

  // DELETE request
  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<ApiResponse<T>>(url, config);
    return response.data.data;
  }

  // Stream request for real-time data
  async stream(url: string): Promise<EventSource> {
    const fullUrl = `${this.client.defaults.baseURL}${url}`;
    return new EventSource(fullUrl);
  }
}

// Export the API service instance
export const apiService = new ApiService(apiClient);

// Helper function to build query parameters
export const buildQueryParams = (params: Record<string, any>): string => {
  const searchParams = new URLSearchParams();
  
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      if (Array.isArray(value)) {
        value.forEach(item => searchParams.append(key, String(item)));
      } else {
        searchParams.append(key, String(value));
      }
    }
  });
  
  return searchParams.toString();
};

// Helper function to create URL with query parameters
export const createUrlWithParams = (baseUrl: string, params?: Record<string, any>): string => {
  if (!params || Object.keys(params).length === 0) {
    return baseUrl;
  }
  
  const queryString = buildQueryParams(params);
  const separator = baseUrl.includes('?') ? '&' : '?';
  
  return `${baseUrl}${separator}${queryString}`;
};