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
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/yth01/apollo/internal/api/utils"
)

// KubernetesHandler handles standard Kubernetes API endpoints
type KubernetesHandler struct {
	client client.Client
}

// NewKubernetesHandler creates a new Kubernetes handler
func NewKubernetesHandler(k8sClient client.Client) *KubernetesHandler {
	return &KubernetesHandler{
		client: k8sClient,
	}
}

// HandleNamespaces returns list of namespaces
func (h *KubernetesHandler) HandleNamespaces(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())

	var namespaces corev1.NamespaceList
	if err := h.client.List(r.Context(), &namespaces); err != nil {
		logger.Error(err, "failed to list namespaces")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to list namespaces")
		return
	}

	// Sort namespaces by name for consistent ordering
	sort.Slice(namespaces.Items, func(i, j int) bool {
		return namespaces.Items[i].Name < namespaces.Items[j].Name
	})

	utils.WriteSuccessResponse(w, namespaces)
}

// HandlePods returns list of pods in a namespace
func (h *KubernetesHandler) HandlePods(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	var pods corev1.PodList
	if err := h.client.List(r.Context(), &pods, client.InNamespace(namespace)); err != nil {
		logger.Error(err, "failed to list pods", "namespace", namespace)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to list pods")
		return
	}

	// Sort pods by name for consistent ordering
	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[i].Name < pods.Items[j].Name
	})

	utils.WriteSuccessResponse(w, pods)
}
