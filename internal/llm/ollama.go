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

package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

// OllamaAnalyzer implements DiagnosisAnalyzer for Ollama
type OllamaAnalyzer struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaAnalyzer creates a new Ollama analyzer instance
func NewOllamaAnalyzer(config diagnosisv1alpha1.LLMConfiguration, apiKey string) (DiagnosisAnalyzer, error) {
	// Note: Ollama typically doesn't require API keys for local deployments,
	// but we maintain the signature for consistency
	// Set default base URL if not provided
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Use default model if not specified
	model := config.Model
	if model == "" {
		model = "llama3.2"
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute, // Ollama can be slow for large models
	}

	return &OllamaAnalyzer{
		baseURL: baseURL,
		model:   model,
		client:  client,
	}, nil
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
	// Metrics
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`
}

// AnalyzePodDiagnosis performs LLM analysis on pod diagnostic data using Ollama
func (o *OllamaAnalyzer) AnalyzePodDiagnosis(ctx context.Context, collectedData diagnosisv1alpha1.CollectedData, triggerCondition *diagnosisv1alpha1.TriggerCondition, policy *diagnosisv1alpha1.DiagnosisPolicy) (diagnosisv1alpha1.DiagnosisAnalysis, error) {
	startTime := time.Now()

	// Build the prompt for analysis
	prompt := o.buildDiagnosisPrompt(collectedData, triggerCondition)

	// Create chat request
	chatRequest := OllamaChatRequest{
		Model: o.model,
		Messages: []OllamaMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,  // Disable streaming for simpler implementation
		Format: "json", // Request JSON format for structured output
		Options: map[string]interface{}{
			"temperature": 0.1, // Low temperature for consistent output
			"top_p":       0.9,
		},
	}

	// Call Ollama API
	analysis, err := o.callOllamaAPI(ctx, chatRequest)
	if err != nil {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("ollama analysis failed: %w", err)
	}

	// Add metadata
	analysis.Provider = "ollama"
	analysis.Model = o.model
	analysis.ProcessingTime = time.Since(startTime).String()

	return analysis, nil
}

// buildDiagnosisPrompt constructs the prompt for LLM analysis
func (o *OllamaAnalyzer) buildDiagnosisPrompt(collectedData diagnosisv1alpha1.CollectedData, triggerCondition *diagnosisv1alpha1.TriggerCondition) string {
	var promptParts []string

	// System-like instruction for Ollama
	promptParts = append(promptParts, "You are a Kubernetes expert assistant analyzing pod issues. You must respond ONLY with valid JSON, no additional text or explanations.")

	// Add trigger condition context
	if triggerCondition != nil {
		promptParts = append(promptParts, fmt.Sprintf("TRIGGER CONDITION: %s (detected at %s)",
			triggerCondition.Type, triggerCondition.DetectedAt.Format(time.RFC3339)))
	}

	// Add pod description
	promptParts = append(promptParts, "POD INFORMATION:")
	promptParts = append(promptParts, collectedData.PodDescription)

	// Add logs if available
	if collectedData.Logs != "" {
		promptParts = append(promptParts, fmt.Sprintf("POD LOGS (%d lines):", collectedData.LogLines))
		promptParts = append(promptParts, collectedData.Logs)
	}

	// Add analysis instructions with clearer JSON format requirement
	promptParts = append(promptParts, `
TASK: Analyze the Kubernetes pod data above and provide a diagnosis in the exact JSON format below. Do not include any text outside the JSON structure.

Required JSON format:
{
  "summary": "Brief 1-2 sentence summary of the main issue",
  "rootCause": "Detailed explanation of what is causing the problem",
  "recommendations": [
    "Specific action item 1",
    "Specific action item 2", 
    "Specific action item 3"
  ]
}

Focus on:
1. Identifying the root cause based on pod status, conditions, events, and logs
2. Providing actionable recommendations that developers can follow
3. Being specific about configuration issues, resource problems, or application errors
4. Suggesting both immediate fixes and preventive measures

Remember: Output ONLY valid JSON, nothing else.`)

	return strings.Join(promptParts, "\n\n")
}

// callOllamaAPI makes the HTTP request to Ollama API
func (o *OllamaAnalyzer) callOllamaAPI(ctx context.Context, request OllamaChatRequest) (diagnosisv1alpha1.DiagnosisAnalysis, error) {
	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/chat", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := o.client.Do(req)
	if err != nil {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaChatResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse the analysis from the message content
	return o.parseAnalysisResponse(ollamaResp.Message.Content), nil
}

// parseAnalysisResponse parses the LLM response into structured analysis
func (o *OllamaAnalyzer) parseAnalysisResponse(response string) diagnosisv1alpha1.DiagnosisAnalysis {
	// Clean the response - remove any potential markdown code blocks
	cleanResponse := strings.TrimSpace(response)

	// Try to find JSON content between curly braces
	startIdx := strings.Index(cleanResponse, "{")
	endIdx := strings.LastIndex(cleanResponse, "}")

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		cleanResponse = cleanResponse[startIdx : endIdx+1]
	}

	// Parse JSON response with flexible rootCause handling
	var parsed struct {
		Summary         string          `json:"summary"`
		RootCause       json.RawMessage `json:"rootCause"`
		Recommendations []string        `json:"recommendations"`
	}

	if err := json.Unmarshal([]byte(cleanResponse), &parsed); err != nil {
		// If JSON parsing fails, try to extract information manually
		return diagnosisv1alpha1.DiagnosisAnalysis{
			Summary:   "Failed to parse LLM response",
			RootCause: fmt.Sprintf("LLM response parsing error: %v. Raw response: %s", err, response),
			Recommendations: []string{
				"Check LLM model compatibility",
				"Verify prompt format",
				"Review pod logs manually",
			},
		}
	}

	// Handle rootCause - can be string or object
	var rootCauseStr string
	if len(parsed.RootCause) > 0 {
		// Try to parse as string first
		var strValue string
		if err := json.Unmarshal(parsed.RootCause, &strValue); err == nil {
			rootCauseStr = strValue
		} else {
			// If not a string, try to parse as object and extract description
			var objValue map[string]interface{}
			if err := json.Unmarshal(parsed.RootCause, &objValue); err == nil {
				// Extract meaningful information from the object
				if description, ok := objValue["description"].(string); ok {
					rootCauseStr = description
				} else if typeVal, ok := objValue["type"].(string); ok {
					rootCauseStr = typeVal
					if description, ok := objValue["description"].(string); ok {
						rootCauseStr += ": " + description
					}
				} else {
					// Fallback: convert entire object to string
					objBytes, _ := json.Marshal(objValue)
					rootCauseStr = string(objBytes)
				}
			} else {
				// Last resort: use raw message as string
				rootCauseStr = string(parsed.RootCause)
			}
		}
	}

	// Validate and clean up recommendations
	recommendations := parsed.Recommendations
	if len(recommendations) == 0 {
		recommendations = []string{"No specific recommendations provided by the model"}
	}

	return diagnosisv1alpha1.DiagnosisAnalysis{
		Summary:         parsed.Summary,
		RootCause:       rootCauseStr,
		Recommendations: recommendations,
	}
}
