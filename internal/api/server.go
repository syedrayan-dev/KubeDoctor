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

package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/yth01/apollo/internal/api/handlers"
	"github.com/yth01/apollo/internal/api/middleware"
)

// Server wraps the HTTP server and Kubernetes client
type Server struct {
	client client.Client
	router *mux.Router

	// Handler instances
	kubernetesHandler *handlers.KubernetesHandler
	policiesHandler   *handlers.PoliciesHandler
	requestsHandler   *handlers.RequestsHandler
	reportsHandler    *handlers.ReportsHandler
	metricsHandler    *handlers.MetricsHandler
	healthHandler     *handlers.HealthHandler
}

// NewServer creates a new API server
func NewServer(k8sClient client.Client) *Server {
	s := &Server{
		client: k8sClient,
		router: mux.NewRouter(),

		// Initialize handlers
		kubernetesHandler: handlers.NewKubernetesHandler(k8sClient),
		policiesHandler:   handlers.NewPoliciesHandler(k8sClient),
		requestsHandler:   handlers.NewRequestsHandler(k8sClient),
		reportsHandler:    handlers.NewReportsHandler(k8sClient),
		metricsHandler:    handlers.NewMetricsHandler(k8sClient),
		healthHandler:     handlers.NewHealthHandler(k8sClient),
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// Apply CORS middleware
	s.router.Use(middleware.CORSMiddleware)

	// API routes
	api := s.router.PathPrefix("/api").Subrouter()

	// Standard Kubernetes APIs
	v1 := api.PathPrefix("/v1").Subrouter()
	v1.HandleFunc("/namespaces", s.kubernetesHandler.HandleNamespaces).Methods("GET")
	v1.HandleFunc("/namespaces/{namespace}/pods", s.kubernetesHandler.HandlePods).Methods("GET")

	// Diagnosis APIs
	diagnosisAPI := api.PathPrefix("/diagnosis/v1alpha1").Subrouter()

	// Policies (cluster-wide and namespace-scoped)
	diagnosisAPI.HandleFunc("/policies", s.policiesHandler.HandleAllPolicies).Methods("GET")
	diagnosisAPI.HandleFunc("/namespaces/{namespace}/policies", s.policiesHandler.HandlePolicies).Methods("GET")
	diagnosisAPI.HandleFunc("/namespaces/{namespace}/policies", s.policiesHandler.HandleCreatePolicy).Methods("POST")
	diagnosisAPI.HandleFunc("/namespaces/{namespace}/policies/{name}", s.policiesHandler.HandleGetPolicy).Methods("GET")
	diagnosisAPI.HandleFunc("/namespaces/{namespace}/policies/{name}", s.policiesHandler.HandleUpdatePolicy).Methods("PUT")
	diagnosisAPI.HandleFunc("/namespaces/{namespace}/policies/{name}", s.policiesHandler.HandleDeletePolicy).Methods("DELETE")

	// Requests (cluster-wide and namespace-scoped)
	diagnosisAPI.HandleFunc("/requests", s.requestsHandler.HandleAllRequests).Methods("GET")
	diagnosisAPI.HandleFunc("/requests/{id}", s.requestsHandler.HandleDeleteRequestLegacy).Methods("DELETE")
	diagnosisAPI.HandleFunc("/namespaces/{namespace}/requests", s.requestsHandler.HandleRequests).Methods("GET")
	diagnosisAPI.HandleFunc("/namespaces/{namespace}/requests", s.requestsHandler.HandleCreateRequest).Methods("POST")
	diagnosisAPI.HandleFunc("/namespaces/{namespace}/requests/{name}", s.requestsHandler.HandleGetRequest).Methods("GET")
	diagnosisAPI.HandleFunc("/namespaces/{namespace}/requests/{name}", s.requestsHandler.HandleDeleteRequest).Methods("DELETE")

	// Reports (cluster-wide)
	diagnosisAPI.HandleFunc("/reports", s.reportsHandler.HandleReports).Methods("GET")
	diagnosisAPI.HandleFunc("/reports/{name}", s.reportsHandler.HandleGetReport).Methods("GET")

	// Metrics
	diagnosisAPI.HandleFunc("/metrics", s.metricsHandler.HandleMetrics).Methods("GET")

	// Health check
	s.router.HandleFunc("/health", s.healthHandler.HandleHealth).Methods("GET")
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Start starts the HTTP server on the specified port
func (s *Server) Start(ctx context.Context, port int) error {
	logger := log.FromContext(ctx)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s,
	}

	logger.Info("Starting API server", "port", port)

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "API server failed to start")
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	logger.Info("Shutting down API server")
	return server.Shutdown(context.Background())
}
