/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
	"github.com/yth01/apollo/internal/api/utils"
)

// RequestsHandler handles DiagnosisRequest CRUD operations
type RequestsHandler struct {
	client client.Client
}

// NewRequestsHandler creates a new requests handler
func NewRequestsHandler(k8sClient client.Client) *RequestsHandler {
	return &RequestsHandler{
		client: k8sClient,
	}
}

// HandleAllRequests returns list of all diagnosis requests cluster-wide
func (h *RequestsHandler) HandleAllRequests(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())

	var requests diagnosisv1alpha1.DiagnosisRequestList
	if err := h.client.List(r.Context(), &requests); err != nil {
		logger.Error(err, "failed to list all diagnosis requests")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to list diagnosis requests")
		return
	}

	// Support filtering by type
	requestType := r.URL.Query().Get("type")
	if requestType != "" {
		var filteredRequests []diagnosisv1alpha1.DiagnosisRequest
		for _, req := range requests.Items {
			if string(req.Spec.Type) == requestType {
				filteredRequests = append(filteredRequests, req)
			}
		}
		requests.Items = filteredRequests
	}

	// Support filtering by status
	status := r.URL.Query().Get("status")
	if status != "" {
		var filteredRequests []diagnosisv1alpha1.DiagnosisRequest
		for _, req := range requests.Items {
			if string(req.Status.Phase) == status {
				filteredRequests = append(filteredRequests, req)
			}
		}
		requests.Items = filteredRequests
	}

	// Sort requests by creation time (newest first)
	sort.Slice(requests.Items, func(i, j int) bool {
		return requests.Items[i].CreationTimestamp.After(requests.Items[j].CreationTimestamp.Time)
	})

	utils.WriteSuccessResponse(w, map[string]interface{}{
		"items": requests.Items,
	})
}

// HandleRequests returns list of diagnosis requests in a namespace
func (h *RequestsHandler) HandleRequests(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	var requests diagnosisv1alpha1.DiagnosisRequestList
	if err := h.client.List(r.Context(), &requests, client.InNamespace(namespace)); err != nil {
		logger.Error(err, "failed to list diagnosis requests", "namespace", namespace)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to list diagnosis requests")
		return
	}

	// Support filtering by type
	requestType := r.URL.Query().Get("type")
	if requestType != "" {
		var filteredRequests []diagnosisv1alpha1.DiagnosisRequest
		for _, req := range requests.Items {
			if string(req.Spec.Type) == requestType {
				filteredRequests = append(filteredRequests, req)
			}
		}
		requests.Items = filteredRequests
	}

	// Sort requests by creation time (newest first)
	sort.Slice(requests.Items, func(i, j int) bool {
		return requests.Items[i].CreationTimestamp.After(requests.Items[j].CreationTimestamp.Time)
	})

	utils.WriteSuccessResponse(w, map[string]interface{}{
		"items": requests.Items,
	})
}

// HandleGetRequest returns a specific diagnosis request
func (h *RequestsHandler) HandleGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	requestName := vars["name"]

	request := &diagnosisv1alpha1.DiagnosisRequest{}
	if err := h.client.Get(r.Context(), client.ObjectKey{
		Name:      requestName,
		Namespace: namespace,
	}, request); err != nil {
		logger.Error(err, "failed to get request", "namespace", namespace, "name", requestName)
		if client.IgnoreNotFound(err) == nil {
			utils.WriteErrorResponse(w, http.StatusNotFound, err, "Request not found")
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to get request")
		}
		return
	}

	utils.WriteSuccessResponse(w, request)
}

// CreateRequestBody represents the request body for creating a diagnosis request
type CreateRequestBody struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		GenerateName string            `json:"generateName"`
		Namespace    string            `json:"namespace"`
		Labels       map[string]string `json:"labels"`
	} `json:"metadata"`
	Spec struct {
		TargetPod struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"targetPod"`
		PolicyRef struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"policyRef"`
		Type string `json:"type"`
	} `json:"spec"`
}

// HandleCreateRequest creates a new diagnosis request
func (h *RequestsHandler) HandleCreateRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	var requestBody CreateRequestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		logger.Error(err, "failed to decode request body")
		utils.WriteErrorResponse(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Generate unique name
	timestamp := time.Now().Unix()
	problemType := strings.ToLower(requestBody.Spec.Type)
	requestName := fmt.Sprintf("%s-%s-%d", requestBody.Spec.TargetPod.Name, problemType, timestamp)

	// Create DiagnosisRequest
	diagnosisRequest := &diagnosisv1alpha1.DiagnosisRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      requestName,
			Namespace: namespace,
			Labels:    requestBody.Metadata.Labels,
		},
		Spec: diagnosisv1alpha1.DiagnosisRequestSpec{
			Type: diagnosisv1alpha1.DiagnosisRequestType(requestBody.Spec.Type),
			TargetPod: diagnosisv1alpha1.TargetPodReference{
				Name:      requestBody.Spec.TargetPod.Name,
				Namespace: requestBody.Spec.TargetPod.Namespace,
			},
			PolicyRef: diagnosisv1alpha1.PolicyReference{
				Name:      requestBody.Spec.PolicyRef.Name,
				Namespace: requestBody.Spec.PolicyRef.Namespace,
			},
		},
	}

	if err := h.client.Create(r.Context(), diagnosisRequest); err != nil {
		logger.Error(err, "failed to create diagnosis request", "namespace", namespace, "name", requestName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to create diagnosis request")
		return
	}

	logger.Info("Created diagnosis request", "namespace", namespace, "name", requestName)
	utils.WriteCreatedResponse(w, diagnosisRequest)
}

// HandleDeleteRequest deletes a diagnosis request
func (h *RequestsHandler) HandleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	requestName := vars["name"]

	request := &diagnosisv1alpha1.DiagnosisRequest{}
	if err := h.client.Get(r.Context(), client.ObjectKey{
		Name:      requestName,
		Namespace: namespace,
	}, request); err != nil {
		logger.Error(err, "failed to get request for deletion", "namespace", namespace, "name", requestName)
		if client.IgnoreNotFound(err) == nil {
			utils.WriteErrorResponse(w, http.StatusNotFound, err, "Request not found")
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to get request")
		}
		return
	}

	if err := h.client.Delete(r.Context(), request); err != nil {
		logger.Error(err, "failed to delete request", "namespace", namespace, "name", requestName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to delete request")
		return
	}

	utils.WriteNoContentResponse(w)
}

// HandleDeleteRequestLegacy handles legacy delete by ID (searches across all namespaces)
func (h *RequestsHandler) HandleDeleteRequestLegacy(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	requestID := vars["id"]

	// Find the request across all namespaces
	var requests diagnosisv1alpha1.DiagnosisRequestList
	if err := h.client.List(r.Context(), &requests); err != nil {
		logger.Error(err, "failed to list diagnosis requests for deletion", "requestID", requestID)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to find request")
		return
	}

	for _, req := range requests.Items {
		if req.Name == requestID {
			if err := h.client.Delete(r.Context(), &req); err != nil {
				logger.Error(err, "failed to delete request", "requestID", requestID, "namespace", req.Namespace)
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to delete request")
				return
			}
			utils.WriteNoContentResponse(w)
			return
		}
	}

	utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("request not found"), "Request not found")
}
