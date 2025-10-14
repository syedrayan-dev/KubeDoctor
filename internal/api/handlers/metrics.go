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

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
	"github.com/yth01/apollo/internal/api/utils"
)

// MetricsHandler handles dashboard metrics endpoints
type MetricsHandler struct {
	client client.Client
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(k8sClient client.Client) *MetricsHandler {
	return &MetricsHandler{
		client: k8sClient,
	}
}

// HandleMetrics returns dashboard metrics
func (h *MetricsHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())

	// Get reports
	var reports diagnosisv1alpha1.DiagnosisReportList
	if err := h.client.List(r.Context(), &reports); err != nil {
		logger.Error(err, "failed to list diagnosis reports for metrics")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to get metrics")
		return
	}

	// Get requests
	var requests diagnosisv1alpha1.DiagnosisRequestList
	if err := h.client.List(r.Context(), &requests); err != nil {
		logger.Error(err, "failed to list diagnosis requests for metrics")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to get metrics")
		return
	}

	// Calculate metrics
	activeRequests := 0
	problemDistribution := map[string]int{
		"Failed":  0,
		"Pending": 0,
		"Running": 0,
		"Unknown": 0,
	}

	for _, req := range requests.Items {
		if req.Status.Phase == diagnosisv1alpha1.DiagnosisRequestPhaseInProgress ||
			req.Status.Phase == diagnosisv1alpha1.DiagnosisRequestPhasePending {
			activeRequests++
		}
	}

	for _, report := range reports.Items {
		problemType := report.Spec.TriggerCondition.Type
		if count, exists := problemDistribution[problemType]; exists {
			problemDistribution[problemType] = count + 1
		}
	}

	// Convert to slice for frontend
	problemDistributionSlice := make([]map[string]interface{}, 0, len(problemDistribution))
	for name, value := range problemDistribution {
		problemDistributionSlice = append(problemDistributionSlice, map[string]interface{}{
			"name":  name,
			"value": value,
		})
	}

	// Get recent reports (last 5, sorted by creation time)
	recentReports := reports.Items
	// Sort by creation timestamp (newest first)
	sort.Slice(recentReports, func(i, j int) bool {
		return recentReports[i].CreationTimestamp.After(recentReports[j].CreationTimestamp.Time)
	})
	// Take the first 5 (most recent)
	if len(recentReports) > 5 {
		recentReports = recentReports[:5]
	}

	response := map[string]interface{}{
		"totalReports":        len(reports.Items),
		"totalRequests":       len(requests.Items),
		"activeRequests":      activeRequests,
		"recentReports":       recentReports,
		"problemDistribution": problemDistributionSlice,
	}

	utils.WriteSuccessResponse(w, response)
}
