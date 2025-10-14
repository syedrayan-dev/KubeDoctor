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

	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

var _ = Describe("Reports Handler", func() {
	var (
		handler *ReportsHandler
		ctx     context.Context
	)

	BeforeEach(func() {
		handler = NewReportsHandler(k8sClient)
		ctx = context.Background()

		// Create 3 test reports
		for i := 1; i <= 3; i++ {
			testReport := &diagnosisv1alpha1.DiagnosisReport{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("test-report-%d", i),
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisReportSpec{
					TriggerCondition: diagnosisv1alpha1.TriggerCondition{
						Type:       "Failed",
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

		// Verify all reports are actually deleted
		Eventually(func() int {
			var reports diagnosisv1alpha1.DiagnosisReportList
			_ = k8sClient.List(ctx, &reports)
			return len(reports.Items)
		}).Should(BeZero())
	})

	Context("When listing diagnosis reports", func() {
		It("should return reports list successfully", func() {
			req := httptest.NewRequest("GET", "/api/diagnosis/v1alpha1/reports", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.HandleReports(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveKey("items"))
			Expect(response).To(HaveKey("totalCount"))

			items, ok := response["items"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(items).To(HaveLen(3))

			totalCount, ok := response["totalCount"].(float64)
			Expect(ok).To(BeTrue())
			Expect(int(totalCount)).To(Equal(3))
		})
	})

	Context("When getting specific diagnosis report", func() {
		It("should return specific report successfully", func() {
			req := httptest.NewRequest("GET", "/api/diagnosis/v1alpha1/reports/test-report-1", nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{"name": "test-report-1"})
			rec := httptest.NewRecorder()

			handler.HandleGetReport(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response diagnosisv1alpha1.DiagnosisReport
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Name).To(Equal("test-report-1"))
			Expect(response.Spec.TriggerCondition.Type).NotTo(BeEmpty())
			Expect(response.Spec.CollectedData.PodDescription).NotTo(BeEmpty())
		})
	})
})
