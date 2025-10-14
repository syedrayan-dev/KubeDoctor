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
	"fmt"
	"strings"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

// NewAnalyzer creates a new DiagnosisAnalyzer based on the LLM configuration
func NewAnalyzer(llmConfig diagnosisv1alpha1.LLMConfiguration, apiKey string) (DiagnosisAnalyzer, error) {
	provider := strings.ToLower(llmConfig.Provider)

	switch provider {
	case "openai":
		return NewOpenAIAnalyzer(llmConfig, apiKey)
	case "ollama":
		return NewOllamaAnalyzer(llmConfig, apiKey)
	// TODO: Implement additional LLM providers
	// case "azure-openai":
	//     return NewAzureOpenAIAnalyzer(llmConfig, apiKey)
	// case "anthropic":
	//     return NewAnthropicAnalyzer(llmConfig, apiKey)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (supported: 'openai', 'ollama')", llmConfig.Provider)
	}
}

// GetSupportedProviders returns a list of supported LLM providers
func GetSupportedProviders() []string {
	return []string{"openai", "ollama"}
	// TODO: Add more providers as they are implemented
	// return []string{"openai", "azure-openai", "anthropic"}
}
