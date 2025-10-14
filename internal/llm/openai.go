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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

// OpenAIAnalyzer implements DiagnosisAnalyzer for OpenAI
type OpenAIAnalyzer struct {
	client *openai.Client
	config diagnosisv1alpha1.LLMConfiguration
}

// NewOpenAIAnalyzer creates a new OpenAI analyzer instance
func NewOpenAIAnalyzer(config diagnosisv1alpha1.LLMConfiguration, apiKey string) (DiagnosisAnalyzer, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for OpenAI provider")
	}

	var clientOpts []option.RequestOption
	clientOpts = append(clientOpts, option.WithAPIKey(apiKey))

	// Use custom base URL if provided
	if config.BaseURL != "" {
		clientOpts = append(clientOpts, option.WithBaseURL(config.BaseURL))
	}

	client := openai.NewClient(clientOpts...)

	return &OpenAIAnalyzer{
		client: &client,
		config: config,
	}, nil
}

// AnalyzePodDiagnosis performs LLM analysis on pod diagnostic data
func (o *OpenAIAnalyzer) AnalyzePodDiagnosis(ctx context.Context, collectedData diagnosisv1alpha1.CollectedData, triggerCondition *diagnosisv1alpha1.TriggerCondition, policy *diagnosisv1alpha1.DiagnosisPolicy) (diagnosisv1alpha1.DiagnosisAnalysis, error) {
	startTime := time.Now()

	// Determine model from configuration
	model := o.config.Model
	if model == "" {
		model = "gpt-4"
	}

	// Build the prompt for LLM analysis
	prompt := o.buildDiagnosisPrompt(collectedData, triggerCondition)

	// Perform OpenAI API call
	analysis, err := o.callOpenAIAPI(ctx, prompt, model)
	if err != nil {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("OpenAI analysis failed: %w", err)
	}

	// Add metadata
	analysis.Provider = "openai"
	analysis.Model = model
	analysis.ProcessingTime = time.Since(startTime).String()

	return analysis, nil
}

// buildDiagnosisPrompt constructs the prompt for LLM analysis
func (o *OpenAIAnalyzer) buildDiagnosisPrompt(collectedData diagnosisv1alpha1.CollectedData, triggerCondition *diagnosisv1alpha1.TriggerCondition) string {
	var promptParts []string

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

	// Add analysis instructions
	promptParts = append(promptParts, `
ANALYSIS INSTRUCTIONS:
Please analyze the above Kubernetes pod data and provide a structured diagnosis in the following JSON format:

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

Return ONLY the JSON response, no additional text.`)

	return strings.Join(promptParts, "\n\n")
}

// parseAnalysisResponse parses the LLM response into structured analysis
func (o *OpenAIAnalyzer) parseAnalysisResponse(response string) diagnosisv1alpha1.DiagnosisAnalysis {
	// Clean the response (remove any markdown code blocks)
	cleanResponse := strings.TrimSpace(response)
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimPrefix(cleanResponse, "```")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	// Parse JSON response with flexible rootCause handling
	var parsed struct {
		Summary         string          `json:"summary"`
		RootCause       json.RawMessage `json:"rootCause"`
		Recommendations []string        `json:"recommendations"`
	}

	if err := json.Unmarshal([]byte(cleanResponse), &parsed); err != nil {
		// If JSON parsing fails, create a fallback response
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

	// Validate parsed content
	if parsed.Summary == "" {
		parsed.Summary = "Pod diagnosis analysis completed"
	}
	if rootCauseStr == "" {
		rootCauseStr = "Unable to determine root cause from available data"
	}
	if len(parsed.Recommendations) == 0 {
		parsed.Recommendations = []string{"Review pod configuration and logs for issues"}
	}

	return diagnosisv1alpha1.DiagnosisAnalysis{
		Summary:         parsed.Summary,
		RootCause:       rootCauseStr,
		Recommendations: parsed.Recommendations,
	}
}

// callOpenAIAPI makes the actual API call to OpenAI
func (o *OpenAIAnalyzer) callOpenAIAPI(ctx context.Context, prompt, model string) (diagnosisv1alpha1.DiagnosisAnalysis, error) {
	// Create chat completion request using v2 SDK
	chatCompletion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("You are an expert Kubernetes troubleshooter. Analyze the provided pod data and provide a comprehensive diagnosis with clear recommendations."),
			openai.UserMessage(prompt),
		},
		Model: model,
	})

	if err != nil {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	// Parse the response
	if len(chatCompletion.Choices) == 0 {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("no response from OpenAI")
	}

	responseText := chatCompletion.Choices[0].Message.Content
	analysis := o.parseAnalysisResponse(responseText)

	return analysis, nil
}
