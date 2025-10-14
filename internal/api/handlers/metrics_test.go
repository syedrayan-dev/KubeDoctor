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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

var _ = Describe("Metrics Handler", func() {
	var (
		handler *MetricsHandler
		ctx     context.Context
	)

	BeforeEach(func() {
		handler = NewMetricsHandler(k8sClient)
		ctx = context.Background()

		// Create 2 test reports with different trigger types
		for i := 1; i <= 2; i++ {
			triggerType := "Failed"
			if i == 2 {
				triggerType = "CrashLoopBackOff"
			}

			testReport := &diagnosisv1alpha1.DiagnosisReport{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("test-report-%d", i),
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisReportSpec{
					TriggerCondition: diagnosisv1alpha1.TriggerCondition{
						Type:       triggerType,
						DetectedAt: metav1.Now(),
					},
					CollectedData: diagnosisv1alpha1.CollectedData{
						PodDescription: fmt.Sprintf("Test pod description %d", i),
						CollectionTime: metav1.Now(),
					},
					Analysis: diagnosisv1alpha1.DiagnosisAnalysis{
						Summary:         "Test analysis summary",
						RootCause:       "Test root cause",
						Recommendations: []string{"Test recommendation"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, testReport)).To(Succeed())
		}

		// Create 3 test requests with different statuses
		statuses := []diagnosisv1alpha1.DiagnosisRequestPhase{
			diagnosisv1alpha1.DiagnosisRequestPhasePending,
			diagnosisv1alpha1.DiagnosisRequestPhaseInProgress,
			diagnosisv1alpha1.DiagnosisRequestPhaseCompleted,
		}

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
				},
			}
			Expect(k8sClient.Create(ctx, testRequest)).To(Succeed())

			// Update status separately (status is a separate subresource)
			testRequest.Status = diagnosisv1alpha1.DiagnosisRequestStatus{
				Phase: statuses[i-1],
			}
			Expect(k8sClient.Status().Update(ctx, testRequest)).To(Succeed())
		}
	})

	AfterEach(func() {
		// Clean up test reports and wait for deletion to complete
		Eventually(func() error {
			var reports diagnosisv1alpha1.DiagnosisReportList
			if err := k8sClient.List(ctx, &reports); err != nil {
				return err
			}
			for _, report := range reports.Items {
				if err := k8sClient.Delete(ctx, &report); err != nil {
					return err
				}
			}
			return nil
		}).Should(Succeed())

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

		// Verify all resources are actually deleted
		Eventually(func() int {
			var reports diagnosisv1alpha1.DiagnosisReportList
			var requests diagnosisv1alpha1.DiagnosisRequestList
			_ = k8sClient.List(ctx, &reports)
			_ = k8sClient.List(ctx, &requests)
			return len(reports.Items) + len(requests.Items)
		}).Should(BeZero())
	})

	Context("When getting dashboard metrics", func() {
		It("should return metrics successfully", func() {
			req := httptest.NewRequest("GET", "/api/diagnosis/v1alpha1/metrics", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.HandleMetrics(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			// Verify metrics structure
			Expect(response).To(HaveKey("problemDistribution"))
			Expect(response).To(HaveKey("activeRequests"))
			Expect(response).To(HaveKey("recentReports"))

			// Verify problem distribution
			problemDistribution, ok := response["problemDistribution"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(problemDistribution).To(HaveLen(4)) // Failed, Pending, Running, Unknown

			// Verify active requests count
			activeRequests, ok := response["activeRequests"].(float64)
			Expect(ok).To(BeTrue())
			Expect(int(activeRequests)).To(Equal(2)) // Pending and InProgress requests

			// Verify recent reports
			recentReports, ok := response["recentReports"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(recentReports).To(HaveLen(2))
		})
	})
})
