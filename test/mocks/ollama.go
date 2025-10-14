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

package mocks

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// MockOllamaServer provides a mock Ollama API server for testing
type MockOllamaServer struct {
	server   *http.Server
	listener net.Listener
	URL      string
}

// OllamaChatRequest represents the request structure for Ollama chat API
type OllamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []OllamaMessage        `json:"messages"`
	Stream   bool                   `json:"stream"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// OllamaMessage represents a message in the chat
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaChatResponse represents the response from Ollama chat API
type OllamaChatResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   OllamaMessage `json:"message"`
	Done      bool          `json:"done"`
	// Metrics (optional for mocking)
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`
}

// NewMockOllamaServer creates and starts a new mock Ollama server
func NewMockOllamaServer() (*MockOllamaServer, error) {
	// Create a listener on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	mock := &MockOllamaServer{
		listener: listener,
		URL:      fmt.Sprintf("http://%s", listener.Addr().String()),
	}

	// Create HTTP server with mock handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat", mock.handleChat)

	mock.server = &http.Server{
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		if err := mock.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Mock Ollama server error: %v\n", err)
		}
	}()

	return mock, nil
}

// handleChat handles the /api/chat endpoint
func (m *MockOllamaServer) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request
	var req OllamaChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate mock diagnosis response based on the content
	analysisJSON := m.generateMockAnalysis(req.Messages)

	// Create Ollama response
	response := OllamaChatResponse{
		Model:     req.Model,
		CreatedAt: time.Now().Format(time.RFC3339),
		Message: OllamaMessage{
			Role:    "assistant",
			Content: analysisJSON,
		},
		Done:               true,
		TotalDuration:      1000000000, // 1 second in nanoseconds
		LoadDuration:       100000000,  // 100ms
		PromptEvalCount:    50,
		PromptEvalDuration: 200000000, // 200ms
		EvalCount:          150,
		EvalDuration:       700000000, // 700ms
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// generateMockAnalysis creates a mock diagnosis analysis based on input
func (m *MockOllamaServer) generateMockAnalysis(messages []OllamaMessage) string {
	// Analyze the prompt to determine what kind of issue is being diagnosed
	var content string
	for _, msg := range messages {
		if msg.Role == "user" {
			content = msg.Content
			break
		}
	}

	// Return different diagnoses based on content patterns
	if containsAny(content, []string{"ImagePullBackOff", "image", "pull"}) {
		return `{
  "summary": "Pod is failing to pull the specified container image from the registry.",
  "rootCause": "The container image specified in the pod specification cannot be pulled. 
    This could be due to the image not existing, incorrect image name/tag, 
    registry authentication issues, or network connectivity problems.",
  "recommendations": [
    "Verify the image name and tag are correct in the pod specification",
    "Check if the image exists in the specified registry",
    "Ensure proper registry authentication credentials are configured",
    "Verify network connectivity to the container registry"
  ]
}`
	}

	if containsAny(content, []string{"CrashLoopBackOff", "crash", "exit"}) {
		return `{
  "summary": "Pod is crashing repeatedly and Kubernetes is backing off restart attempts.",
  "rootCause": "The container is exiting with a non-zero exit code repeatedly. 
    This indicates an application error, misconfiguration, or missing dependencies 
    that prevent the application from starting successfully.",
  "recommendations": [
    "Check application logs for specific error messages",
    "Verify application configuration and environment variables",
    "Ensure all required dependencies and services are available",
    "Review resource limits and requests for the container"
  ]
}`
	}

	if containsAny(content, []string{"Pending", "pending", "schedule"}) {
		return `{
  "summary": "Pod is stuck in Pending state and cannot be scheduled to any node.",
  "rootCause": "The Kubernetes scheduler cannot find a suitable node to place the pod. 
    This is typically due to insufficient resources, node selector constraints, 
    or scheduling policies that cannot be satisfied.",
  "recommendations": [
    "Check cluster resource availability (CPU, memory, disk)",
    "Verify node selectors and affinity rules are correct",
    "Review pod resource requests and limits",
    "Check for tainted nodes and corresponding tolerations"
  ]
}`
	}

	// Default generic analysis
	return `{
  "summary": "Pod is experiencing issues that require investigation.",
  "rootCause": "Based on the provided pod information, there appears to be a configuration 
    or runtime issue preventing the pod from running successfully.",
  "recommendations": [
    "Review pod specification for configuration errors",
    "Check pod events and logs for specific error messages",
    "Verify resource availability in the cluster",
    "Ensure all dependencies and services are properly configured"
  ]
}`
}

// containsAny checks if the text contains any of the given substrings (case-insensitive)
func containsAny(text string, substrings []string) bool {
	textLower := strings.ToLower(text)
	for _, substr := range substrings {
		if strings.Contains(textLower, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// Stop gracefully stops the mock server
func (m *MockOllamaServer) Stop() error {
	if m.server != nil {
		return m.server.Close()
	}
	return nil
}
