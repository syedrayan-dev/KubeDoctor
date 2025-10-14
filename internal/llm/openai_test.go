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

var _ = Describe("OpenAI Analyzer", func() {
	var (
		validConfig = diagnosisv1alpha1.LLMConfiguration{
			Provider: "openai",
			Model:    "gpt-4",
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
		It("should create analyzer successfully with valid API key", func() {
			analyzer, err := NewOpenAIAnalyzer(validConfig, "test-api-key")

			Expect(err).NotTo(HaveOccurred())
			Expect(analyzer).NotTo(BeNil())
		})

		It("should return error when API key is missing", func() {
			analyzer, err := NewOpenAIAnalyzer(validConfig, "")

			Expect(err).To(HaveOccurred())
			Expect(analyzer).To(BeNil())
		})
	})

	Context("When analyzing pod diagnosis", func() {
		var (
			analyzer   *OpenAIAnalyzer
			mockServer *httptest.Server
		)

		BeforeEach(func() {
			// Create mock OpenAI API server
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Mock successful response
				response := map[string]interface{}{
					"choices": []map[string]interface{}{
						{
							"message": map[string]interface{}{
								"content": `{
									"summary": "Pod analysis completed",
									"rootCause": "Mock diagnosis result",
									"recommendations": ["Mock recommendation 1", "Mock recommendation 2"]
								}`,
							},
						},
					},
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			}))

			// Create analyzer with mock server URL
			config := validConfig
			config.BaseURL = mockServer.URL + "/v1"

			analyzerInterface, err := NewOpenAIAnalyzer(config, "test-api-key")
			Expect(err).NotTo(HaveOccurred())
			var ok bool
			analyzer, ok = analyzerInterface.(*OpenAIAnalyzer)
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
			Expect(analysis.Provider).To(Equal("openai"))
			Expect(analysis.Model).To(Equal("gpt-4"))
			Expect(analysis.ProcessingTime).NotTo(BeEmpty())
			Expect(analysis.Summary).NotTo(BeEmpty())
			Expect(analysis.RootCause).NotTo(BeEmpty())
			Expect(analysis.Recommendations).NotTo(BeEmpty())
		})

		It("should handle invalid LLM response gracefully", func() {
			// Create mock server that returns invalid response
			invalidMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"choices": []map[string]interface{}{
						{
							"message": map[string]interface{}{
								"content": "This is not valid JSON",
							},
						},
					},
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			}))
			defer invalidMockServer.Close()

			// Create analyzer with invalid mock server
			config := validConfig
			config.BaseURL = invalidMockServer.URL + "/v1"

			analyzerInterface, err := NewOpenAIAnalyzer(config, "test-api-key")
			Expect(err).NotTo(HaveOccurred())
			invalidAnalyzer, ok := analyzerInterface.(*OpenAIAnalyzer)
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
