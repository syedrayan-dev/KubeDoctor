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
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

var _ = Describe("Ollama Analyzer", func() {
	var (
		validConfig = diagnosisv1alpha1.LLMConfiguration{
			Provider: "ollama",
		}

		testCollectedData = diagnosisv1alpha1.CollectedData{
			PodDescription: "Pod test-pod is in Failed state",
			Logs:           "Error: container failed to start",
			LogLines:       1,
			CollectionTime: metav1.Now(),
		}

		testTriggerCondition = &diagnosisv1alpha1.TriggerCondition{
			Type:       "Failed",
			DetectedAt: metav1.Now(),
		}
	)

	Context("When creating analyzer", func() {
		It("should create analyzer successfully", func() {
			analyzer, err := NewOllamaAnalyzer(validConfig, "")

			Expect(err).NotTo(HaveOccurred())
			Expect(analyzer).NotTo(BeNil())
		})
	})

	Context("When analyzing pod diagnosis", func() {
		var (
			analyzer   *OllamaAnalyzer
			mockServer *httptest.Server
		)

		BeforeEach(func() {
			// Create mock Ollama API server
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify Ollama-specific request format
				Expect(r.Method).To(Equal("POST"))
				Expect(r.URL.Path).To(Equal("/api/chat"))

				// Verify request body structure
				var reqBody OllamaChatRequest
				_ = json.NewDecoder(r.Body).Decode(&reqBody)
				Expect(reqBody.Stream).To(BeFalse())
				Expect(reqBody.Format).To(Equal("json"))
				Expect(reqBody.Options).To(HaveKey("temperature"))

				// Mock Ollama response
				response := map[string]interface{}{
					"model":      "llama3.2",
					"created_at": "2025-01-10T12:00:00Z",
					"message": map[string]interface{}{
						"role": "assistant",
						"content": `{
							"summary": "Pod analysis completed",
							"rootCause": "Mock diagnosis result",
							"recommendations": ["Mock recommendation 1", "Mock recommendation 2"]
						}`,
					},
					"done":           true,
					"total_duration": 2500000000,
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			}))

			// Create analyzer with mock server URL
			config := validConfig
			config.BaseURL = mockServer.URL

			analyzerInterface, err := NewOllamaAnalyzer(config, "")
			Expect(err).NotTo(HaveOccurred())
			var ok bool
			analyzer, ok = analyzerInterface.(*OllamaAnalyzer)
			Expect(ok).To(BeTrue())
		})

		AfterEach(func() {
			if mockServer != nil {
				mockServer.Close()
			}
		})

		It("should complete end-to-end analysis successfully", func() {
			ctx := context.Background()
			policy := &diagnosisv1alpha1.DiagnosisPolicy{}

			analysis, err := analyzer.AnalyzePodDiagnosis(ctx, testCollectedData, testTriggerCondition, policy)

			Expect(err).NotTo(HaveOccurred())
			Expect(analysis.Provider).To(Equal("ollama"))
			Expect(analysis.Model).To(Equal("llama3.2"))
			Expect(analysis.ProcessingTime).NotTo(BeEmpty())
			Expect(analysis.Summary).NotTo(BeEmpty())
			Expect(analysis.RootCause).NotTo(BeEmpty())
			Expect(analysis.Recommendations).NotTo(BeEmpty())
		})

		It("should handle invalid LLM response gracefully", func() {
			// Create mock server that returns invalid response
			invalidMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"model": "llama3.2",
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "This is not valid JSON",
					},
					"done": true,
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			}))
			defer invalidMockServer.Close()

			// Create analyzer with invalid mock server
			config := validConfig
			config.BaseURL = invalidMockServer.URL

			analyzerInterface, err := NewOllamaAnalyzer(config, "")
			Expect(err).NotTo(HaveOccurred())
			invalidAnalyzer, ok := analyzerInterface.(*OllamaAnalyzer)
			Expect(ok).To(BeTrue())

			ctx := context.Background()
			policy := &diagnosisv1alpha1.DiagnosisPolicy{}

			analysis, err := invalidAnalyzer.AnalyzePodDiagnosis(ctx, testCollectedData, testTriggerCondition, policy)

			Expect(err).NotTo(HaveOccurred())
			Expect(analysis.Summary).To(Equal("Failed to parse LLM response"))
			Expect(analysis.RootCause).To(ContainSubstring("LLM response parsing error"))
			Expect(analysis.Recommendations).To(HaveLen(3))
		})

	})
})
