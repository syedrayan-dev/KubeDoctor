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
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/yth01/apollo/internal/api/utils"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	client client.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(k8sClient client.Client) *HealthHandler {
	return &HealthHandler{
		client: k8sClient,
	}
}

// HandleHealth returns server health status
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	utils.WriteSuccessResponse(w, response)
}
