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
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
	"github.com/yth01/apollo/internal/api/utils"
)

// PoliciesHandler handles DiagnosisPolicy CRUD operations
type PoliciesHandler struct {
	client client.Client
}

// NewPoliciesHandler creates a new policies handler
func NewPoliciesHandler(k8sClient client.Client) *PoliciesHandler {
	return &PoliciesHandler{
		client: k8sClient,
	}
}

// HandleAllPolicies returns list of all diagnosis policies cluster-wide
func (h *PoliciesHandler) HandleAllPolicies(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())

	var policies diagnosisv1alpha1.DiagnosisPolicyList
	if err := h.client.List(r.Context(), &policies); err != nil {
		logger.Error(err, "failed to list all diagnosis policies")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to list diagnosis policies")
		return
	}

	// Sort policies by namespace then name for consistent ordering
	sort.Slice(policies.Items, func(i, j int) bool {
		if policies.Items[i].Namespace != policies.Items[j].Namespace {
			return policies.Items[i].Namespace < policies.Items[j].Namespace
		}
		return policies.Items[i].Name < policies.Items[j].Name
	})

	utils.WriteSuccessResponse(w, map[string]interface{}{
		"items": policies.Items,
	})
}

// HandlePolicies returns list of diagnosis policies in a namespace
func (h *PoliciesHandler) HandlePolicies(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	var policies diagnosisv1alpha1.DiagnosisPolicyList
	if err := h.client.List(r.Context(), &policies, client.InNamespace(namespace)); err != nil {
		logger.Error(err, "failed to list diagnosis policies", "namespace", namespace)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to list diagnosis policies")
		return
	}

	// Sort policies by name for consistent ordering
	sort.Slice(policies.Items, func(i, j int) bool {
		return policies.Items[i].Name < policies.Items[j].Name
	})

	utils.WriteSuccessResponse(w, map[string]interface{}{
		"items": policies.Items,
	})
}

// HandleGetPolicy returns a specific diagnosis policy
func (h *PoliciesHandler) HandleGetPolicy(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	policyName := vars["name"]

	policy := &diagnosisv1alpha1.DiagnosisPolicy{}
	if err := h.client.Get(r.Context(), client.ObjectKey{
		Name:      policyName,
		Namespace: namespace,
	}, policy); err != nil {
		logger.Error(err, "failed to get policy", "namespace", namespace, "name", policyName)
		if client.IgnoreNotFound(err) == nil {
			utils.WriteErrorResponse(w, http.StatusNotFound, err, "Policy not found")
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to get policy")
		}
		return
	}

	utils.WriteSuccessResponse(w, policy)
}

// HandleCreatePolicy creates a new diagnosis policy
func (h *PoliciesHandler) HandleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	var policy diagnosisv1alpha1.DiagnosisPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		logger.Error(err, "failed to decode policy request")
		utils.WriteErrorResponse(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Ensure namespace matches URL parameter
	policy.Namespace = namespace

	// Set metadata if not provided
	if policy.CreationTimestamp.IsZero() {
		policy.CreationTimestamp = metav1.Now()
	}

	if err := h.client.Create(r.Context(), &policy); err != nil {
		logger.Error(err, "failed to create policy", "namespace", namespace, "name", policy.Name)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to create policy")
		return
	}

	utils.WriteCreatedResponse(w, policy)
}

// HandleUpdatePolicy updates an existing diagnosis policy
func (h *PoliciesHandler) HandleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	policyName := vars["name"]

	// Get existing policy
	existingPolicy := &diagnosisv1alpha1.DiagnosisPolicy{}
	if err := h.client.Get(r.Context(), client.ObjectKey{
		Name:      policyName,
		Namespace: namespace,
	}, existingPolicy); err != nil {
		logger.Error(err, "failed to get existing policy", "namespace", namespace, "name", policyName)
		if client.IgnoreNotFound(err) == nil {
			utils.WriteErrorResponse(w, http.StatusNotFound, err, "Policy not found")
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to get policy")
		}
		return
	}

	// Decode update request
	var updatedPolicy diagnosisv1alpha1.DiagnosisPolicy
	if err := json.NewDecoder(r.Body).Decode(&updatedPolicy); err != nil {
		logger.Error(err, "failed to decode policy update request")
		utils.WriteErrorResponse(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Preserve metadata and ensure namespace/name consistency
	updatedPolicy.ObjectMeta = existingPolicy.ObjectMeta
	updatedPolicy.Namespace = namespace
	updatedPolicy.Name = policyName

	if err := h.client.Update(r.Context(), &updatedPolicy); err != nil {
		logger.Error(err, "failed to update policy", "namespace", namespace, "name", policyName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to update policy")
		return
	}

	utils.WriteSuccessResponse(w, updatedPolicy)
}

// HandleDeletePolicy deletes a diagnosis policy
func (h *PoliciesHandler) HandleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	policyName := vars["name"]

	policy := &diagnosisv1alpha1.DiagnosisPolicy{}
	if err := h.client.Get(r.Context(), client.ObjectKey{
		Name:      policyName,
		Namespace: namespace,
	}, policy); err != nil {
		logger.Error(err, "failed to get policy for deletion", "namespace", namespace, "name", policyName)
		if client.IgnoreNotFound(err) == nil {
			utils.WriteErrorResponse(w, http.StatusNotFound, err, "Policy not found")
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to get policy")
		}
		return
	}

	if err := h.client.Delete(r.Context(), policy); err != nil {
		logger.Error(err, "failed to delete policy", "namespace", namespace, "name", policyName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to delete policy")
		return
	}

	utils.WriteNoContentResponse(w)
}
