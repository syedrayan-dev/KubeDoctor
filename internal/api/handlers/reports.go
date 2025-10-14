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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
	"github.com/yth01/apollo/internal/api/utils"
)

// ReportsHandler handles DiagnosisReport operations
type ReportsHandler struct {
	client client.Client
}

// NewReportsHandler creates a new reports handler
func NewReportsHandler(k8sClient client.Client) *ReportsHandler {
	return &ReportsHandler{
		client: k8sClient,
	}
}

// HandleReports returns list of diagnosis reports
func (h *ReportsHandler) HandleReports(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())

	var reports diagnosisv1alpha1.DiagnosisReportList
	listOpts := []client.ListOption{}

	// Support namespace filtering
	namespace := r.URL.Query().Get("namespace")
	if namespace != "" {
		listOpts = append(listOpts, client.InNamespace(namespace))
	}

	if err := h.client.List(r.Context(), &reports, listOpts...); err != nil {
		logger.Error(err, "failed to list diagnosis reports")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to list diagnosis reports")
		return
	}

	// Filter by type if specified
	problemType := r.URL.Query().Get("type")
	filteredReports := make([]diagnosisv1alpha1.DiagnosisReport, 0)

	if problemType != "" {
		for _, report := range reports.Items {
			if report.Spec.TriggerCondition.Type == problemType {
				filteredReports = append(filteredReports, report)
			}
		}
	} else {
		filteredReports = reports.Items
	}

	// Sort reports by creation time (newest first)
	sort.Slice(filteredReports, func(i, j int) bool {
		return filteredReports[i].CreationTimestamp.After(filteredReports[j].CreationTimestamp.Time)
	})

	utils.WriteSuccessResponse(w, map[string]interface{}{
		"items":      filteredReports,
		"totalCount": len(filteredReports),
	})
}

// HandleGetReport returns a specific diagnosis report
func (h *ReportsHandler) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	vars := mux.Vars(r)
	reportName := vars["name"]

	if reportName == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil, "Report name is required")
		return
	}

	logger.Info("Fetching report", "reportName", reportName)

	// Find the report across all namespaces (reports are cluster-wide)
	var reports diagnosisv1alpha1.DiagnosisReportList
	if err := h.client.List(r.Context(), &reports); err != nil {
		logger.Error(err, "failed to list diagnosis reports", "reportName", reportName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err, "Failed to get diagnosis report")
		return
	}

	for _, report := range reports.Items {
		if report.Name == reportName {
			logger.Info("Found report", "reportName", reportName, "namespace", report.Namespace)
			utils.WriteSuccessResponse(w, report)
			return
		}
	}

	logger.Info("Report not found", "reportName", reportName, "totalReports", len(reports.Items))
	utils.WriteErrorResponse(w, http.StatusNotFound, nil, "Report not found")
}
