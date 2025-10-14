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

var _ = Describe("Requests Handler", func() {
	var (
		handler *RequestsHandler
		ctx     context.Context
	)

	BeforeEach(func() {
		handler = NewRequestsHandler(k8sClient)
		ctx = context.Background()

		// Create 3 test requests
		for i := 1; i <= 3; i++ {
			testRequest := &diagnosisv1alpha1.DiagnosisRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("test-request-%d", i),
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisRequestSpec{
					Type: diagnosisv1alpha1.DiagnosisRequestTypeAutomatic,
					TargetPod: diagnosisv1alpha1.TargetPodReference{
						Name:      fmt.Sprintf("test-pod-%d", i),
						Namespace: "default",
					},
					PolicyRef: diagnosisv1alpha1.PolicyReference{
						Name:      "test-policy",
						Namespace: "default",
					},
					TriggerCondition: &diagnosisv1alpha1.TriggerCondition{
						Type:       "Failed",
						DetectedAt: metav1.Now(),
					},
				},
			}
			Expect(k8sClient.Create(ctx, testRequest)).To(Succeed())
		}
	})

	AfterEach(func() {
		// Clean up test requests and wait for deletion to complete
		Eventually(func() error {
			var requests diagnosisv1alpha1.DiagnosisRequestList
			if err := k8sClient.List(ctx, &requests); err != nil {
				return err
			}
			for _, request := range requests.Items {
				if err := k8sClient.Delete(ctx, &request); err != nil {
					return err
				}
			}
			return nil
		}).Should(Succeed())

		// Verify all requests are actually deleted
		Eventually(func() int {
			var requests diagnosisv1alpha1.DiagnosisRequestList
			_ = k8sClient.List(ctx, &requests)
			return len(requests.Items)
		}).Should(BeZero())
	})

	Context("When listing all requests", func() {
		It("should return all requests successfully", func() {
			req := httptest.NewRequest("GET", "/api/diagnosis/v1alpha1/requests", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.HandleAllRequests(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveKey("items"))
			items, ok := response["items"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(items).To(HaveLen(3))
		})
	})

	Context("When listing requests in namespace", func() {
		It("should return namespace requests successfully", func() {
			req := httptest.NewRequest("GET", "/api/diagnosis/v1alpha1/namespaces/default/requests", nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{"namespace": "default"})
			rec := httptest.NewRecorder()

			handler.HandleRequests(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveKey("items"))
			items, ok := response["items"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(items).To(HaveLen(3))
		})
	})

	Context("When getting specific request", func() {
		It("should return specific request successfully", func() {
			req := httptest.NewRequest("GET", "/api/diagnosis/v1alpha1/namespaces/default/requests/test-request-1", nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{
				"namespace": "default",
				"name":      "test-request-1",
			})
			rec := httptest.NewRecorder()

			handler.HandleGetRequest(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response diagnosisv1alpha1.DiagnosisRequest
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Name).To(Equal("test-request-1"))
			Expect(response.Spec.Type).To(Equal(diagnosisv1alpha1.DiagnosisRequestTypeAutomatic))
		})
	})

	Context("When creating request", func() {
		It("should create request successfully", func() {
			newRequest := diagnosisv1alpha1.DiagnosisRequest{
				Spec: diagnosisv1alpha1.DiagnosisRequestSpec{
					Type: diagnosisv1alpha1.DiagnosisRequestTypeManual,
					TargetPod: diagnosisv1alpha1.TargetPodReference{
						Name:      "new-test-pod",
						Namespace: "default",
					},
					PolicyRef: diagnosisv1alpha1.PolicyReference{
						Name:      "new-test-policy",
						Namespace: "default",
					},
				},
			}

			bodyBytes, err := json.Marshal(newRequest)
			Expect(err).NotTo(HaveOccurred())

			req := httptest.NewRequest("POST", "/api/diagnosis/v1alpha1/namespaces/default/requests", bytes.NewReader(bodyBytes))
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{"namespace": "default"})
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleCreateRequest(rec, req)

			Expect(rec.Code).To(Equal(http.StatusCreated))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response diagnosisv1alpha1.DiagnosisRequest
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Name).NotTo(BeEmpty()) // Auto-generated name
			Expect(response.Namespace).To(Equal("default"))
			Expect(response.Spec.Type).To(Equal(diagnosisv1alpha1.DiagnosisRequestTypeManual))
		})
	})

	Context("When deleting request", func() {
		It("should delete request successfully", func() {
			req := httptest.NewRequest("DELETE", "/api/diagnosis/v1alpha1/namespaces/default/requests/test-request-1", nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{
				"namespace": "default",
				"name":      "test-request-1",
			})
			rec := httptest.NewRecorder()

			handler.HandleDeleteRequest(rec, req)

			Expect(rec.Code).To(Equal(http.StatusNoContent))

			// Verify request is deleted
			var requests diagnosisv1alpha1.DiagnosisRequestList
			err := k8sClient.List(ctx, &requests)
			Expect(err).NotTo(HaveOccurred())
			Expect(requests.Items).To(HaveLen(2)) // Only 2 requests should remain
		})
	})

	Context("When deleting request by legacy method", func() {
		It("should delete request successfully", func() {
			req := httptest.NewRequest("DELETE", "/api/diagnosis/v1alpha1/requests/test-request-2", nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{"id": "test-request-2"})
			rec := httptest.NewRecorder()

			handler.HandleDeleteRequestLegacy(rec, req)

			Expect(rec.Code).To(Equal(http.StatusNoContent))

			// Verify request is deleted
			var requests diagnosisv1alpha1.DiagnosisRequestList
			err := k8sClient.List(ctx, &requests)
			Expect(err).NotTo(HaveOccurred())
			Expect(requests.Items).To(HaveLen(2)) // Only 2 requests should remain
		})
	})
})
