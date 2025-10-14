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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

var _ = Describe("Policies Handler", func() {
	var (
		handler *PoliciesHandler
		ctx     context.Context
	)

	BeforeEach(func() {
		handler = NewPoliciesHandler(k8sClient)
		ctx = context.Background()

		// Create 2 test policies
		for i := 1; i <= 2; i++ {
			testPolicy := &diagnosisv1alpha1.DiagnosisPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("test-policy-%d", i),
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisPolicySpec{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": fmt.Sprintf("test-%d", i)},
					},
					TriggerConditions: []diagnosisv1alpha1.PodConditionTrigger{
						{Type: "Failed"},
					},
					LLMConfig: diagnosisv1alpha1.LLMConfiguration{
						Provider: "ollama",
						Model:    "test-model",
					},
				},
			}
			Expect(k8sClient.Create(ctx, testPolicy)).To(Succeed())
		}
	})

	AfterEach(func() {
		// Clean up test policies and wait for deletion to complete
		Eventually(func() error {
			var policies diagnosisv1alpha1.DiagnosisPolicyList
			if err := k8sClient.List(ctx, &policies); err != nil {
				return err
			}
			for _, policy := range policies.Items {
				if err := k8sClient.Delete(ctx, &policy); err != nil {
					return err
				}
			}
			return nil
		}).Should(Succeed())

		// Verify all policies are actually deleted
		Eventually(func() int {
			var policies diagnosisv1alpha1.DiagnosisPolicyList
			_ = k8sClient.List(ctx, &policies)
			return len(policies.Items)
		}).Should(BeZero())
	})

	Context("When listing all policies", func() {
		It("should return all policies successfully", func() {
			req := httptest.NewRequest("GET", "/api/diagnosis/v1alpha1/policies", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.HandleAllPolicies(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveKey("items"))
			items, ok := response["items"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(items).To(HaveLen(2))
		})
	})

	Context("When listing policies in namespace", func() {
		It("should return namespace policies successfully", func() {
			req := httptest.NewRequest("GET", "/api/diagnosis/v1alpha1/namespaces/default/policies", nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{"namespace": "default"})
			rec := httptest.NewRecorder()

			handler.HandlePolicies(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveKey("items"))
			items, ok := response["items"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(items).To(HaveLen(2))
		})
	})

	Context("When getting specific policy", func() {
		It("should return specific policy successfully", func() {
			req := httptest.NewRequest("GET", "/api/diagnosis/v1alpha1/namespaces/default/policies/test-policy-1", nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{
				"namespace": "default",
				"name":      "test-policy-1",
			})
			rec := httptest.NewRecorder()

			handler.HandleGetPolicy(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response diagnosisv1alpha1.DiagnosisPolicy
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Name).To(Equal("test-policy-1"))
			Expect(response.Spec.LLMConfig.Provider).NotTo(BeEmpty())
		})
	})

	Context("When creating policy", func() {
		It("should create policy successfully", func() {
			newPolicy := diagnosisv1alpha1.DiagnosisPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "new-test-policy",
				},
				Spec: diagnosisv1alpha1.DiagnosisPolicySpec{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "new-test"},
					},
					TriggerConditions: []diagnosisv1alpha1.PodConditionTrigger{
						{Type: "Failed"},
					},
					LLMConfig: diagnosisv1alpha1.LLMConfiguration{
						Provider: "openai",
						Model:    "gpt-4",
					},
				},
			}

			bodyBytes, err := json.Marshal(newPolicy)
			Expect(err).NotTo(HaveOccurred())

			req := httptest.NewRequest("POST", "/api/diagnosis/v1alpha1/namespaces/default/policies", bytes.NewReader(bodyBytes))
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{"namespace": "default"})
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleCreatePolicy(rec, req)

			Expect(rec.Code).To(Equal(http.StatusCreated))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response diagnosisv1alpha1.DiagnosisPolicy
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Name).To(Equal("new-test-policy"))
			Expect(response.Namespace).To(Equal("default"))
		})
	})

	Context("When updating policy", func() {
		It("should update policy successfully", func() {
			updatePolicy := diagnosisv1alpha1.DiagnosisPolicy{
				Spec: diagnosisv1alpha1.DiagnosisPolicySpec{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "updated-test"},
					},
					TriggerConditions: []diagnosisv1alpha1.PodConditionTrigger{
						{Type: "Failed"},
					},
					LLMConfig: diagnosisv1alpha1.LLMConfiguration{
						Provider: "openai",
						Model:    "gpt-4",
					},
				},
			}

			bodyBytes, err := json.Marshal(updatePolicy)
			Expect(err).NotTo(HaveOccurred())

			req := httptest.NewRequest("PUT", "/api/diagnosis/v1alpha1/namespaces/default/policies/test-policy-1", bytes.NewReader(bodyBytes))
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{
				"namespace": "default",
				"name":      "test-policy-1",
			})
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleUpdatePolicy(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response diagnosisv1alpha1.DiagnosisPolicy
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Name).To(Equal("test-policy-1"))
			Expect(response.Spec.LLMConfig.Provider).To(Equal("openai"))
		})
	})

	Context("When deleting policy", func() {
		It("should delete policy successfully", func() {
			req := httptest.NewRequest("DELETE", "/api/diagnosis/v1alpha1/namespaces/default/policies/test-policy-1", nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{
				"namespace": "default",
				"name":      "test-policy-1",
			})
			rec := httptest.NewRecorder()

			handler.HandleDeletePolicy(rec, req)

			Expect(rec.Code).To(Equal(http.StatusNoContent))

			// Verify policy is deleted
			var policies diagnosisv1alpha1.DiagnosisPolicyList
			err := k8sClient.List(ctx, &policies)
			Expect(err).NotTo(HaveOccurred())
			Expect(policies.Items).To(HaveLen(1)) // Only 1 policy should remain
		})
	})
})
