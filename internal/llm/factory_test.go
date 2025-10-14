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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

var _ = Describe("LLM Factory", func() {
	Context("When creating analyzer", func() {
		It("should create analyzer successfully for supported providers", func() {
			testCases := []struct {
				provider string
				apiKey   string
			}{
				{"openai", "test-api-key"},
				{"ollama", ""},
			}

			for _, tc := range testCases {
				config := diagnosisv1alpha1.LLMConfiguration{
					Provider: tc.provider,
					Model:    "test-model",
				}
				analyzer, err := NewAnalyzer(config, tc.apiKey)

				Expect(err).NotTo(HaveOccurred())
				Expect(analyzer).NotTo(BeNil())
			}
		})

		It("should handle unsupported provider gracefully", func() {
			config := diagnosisv1alpha1.LLMConfiguration{
				Provider: "unsupported-provider",
				Model:    "some-model",
			}

			analyzer, err := NewAnalyzer(config, "test-key")

			Expect(err).To(HaveOccurred())
			Expect(analyzer).To(BeNil())
		})

		It("should require API key for OpenAI provider", func() {
			config := diagnosisv1alpha1.LLMConfiguration{
				Provider: "openai",
				Model:    "gpt-4",
			}

			analyzer, err := NewAnalyzer(config, "")

			Expect(err).To(HaveOccurred())
			Expect(analyzer).To(BeNil())
		})
	})
})
