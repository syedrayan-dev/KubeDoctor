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

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

var _ = Describe("DiagnosisReport Controller", func() {
	Context("When reconciling a DiagnosisReport", func() {
		var (
			report *diagnosisv1alpha1.DiagnosisReport
			ctx    context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()

			// Create a valid DiagnosisReport with all required fields
			report = &diagnosisv1alpha1.DiagnosisReport{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-report",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisReportSpec{
					TargetPod: diagnosisv1alpha1.PodReference{
						Name:      "test-pod",
						Namespace: "default",
						UID:       "test-uid-123",
						Image:     "nginx:latest",
					},
					PolicyRef: diagnosisv1alpha1.PolicyReference{
						Name:      "test-policy",
						Namespace: "default",
					},
					TriggerCondition: diagnosisv1alpha1.TriggerCondition{
						Type:       "Failed",
						DetectedAt: metav1.Now(),
					},
					CollectedData: diagnosisv1alpha1.CollectedData{
						PodDescription: "Test pod description with detailed information",
						Logs:           "Test logs from container\nError: Something went wrong",
						LogLines:       2,
						CollectionTime: metav1.Now(),
					},
					Analysis: diagnosisv1alpha1.DiagnosisAnalysis{
						Provider:  "ollama",
						Model:     "test-model",
						Summary:   "Pod failed due to configuration error",
						RootCause: "Invalid environment variable configuration",
						Recommendations: []string{
							"Check environment variable settings",
							"Verify configuration file syntax",
						},
						ProcessingTime: "2.5s",
					},
				},
			}
			Expect(k8sClient.Create(ctx, report)).To(Succeed())

			// Wait for report to be available
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(report), &diagnosisv1alpha1.DiagnosisReport{})
			}, time.Second*5, time.Millisecond*100).Should(Succeed())
		})

		AfterEach(func() {
			// Clean up resources
			Expect(k8sClient.Delete(ctx, report)).To(Succeed())
		})

		It("should successfully reconcile without errors", func() {
			reconciler := &DiagnosisReportReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-report",
					Namespace: "default",
				},
			}

			// Reconcile should not return errors
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			// Report should still exist and be unchanged
			var updatedReport diagnosisv1alpha1.DiagnosisReport
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(report), &updatedReport)).To(Succeed())
			Expect(updatedReport.Name).To(Equal("test-report"))
			Expect(updatedReport.Spec.Analysis.Summary).To(Equal("Pod failed due to configuration error"))
		})
	})

	Context("When handling non-existent report", func() {
		It("should handle non-existent report gracefully", func() {
			reconciler := &DiagnosisReportReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "non-existent-report",
					Namespace: "default",
				},
			}

			// Should not return error for non-existent report
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))
		})
	})

	Context("When report is deleted", func() {
		It("should handle report deletion gracefully", func() {
			ctx := context.Background()
			reconciler := &DiagnosisReportReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Create a report
			report := &diagnosisv1alpha1.DiagnosisReport{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-report-delete",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisReportSpec{
					TargetPod: diagnosisv1alpha1.PodReference{
						Name:      "test-pod-delete",
						Namespace: "default",
						UID:       "test-uid-789",
						Image:     "redis:latest",
					},
					PolicyRef: diagnosisv1alpha1.PolicyReference{
						Name:      "test-policy-delete",
						Namespace: "default",
					},
					TriggerCondition: diagnosisv1alpha1.TriggerCondition{
						Type:       "Failed",
						DetectedAt: metav1.Now(),
					},
					CollectedData: diagnosisv1alpha1.CollectedData{
						PodDescription: "Deleted pod description",
						Logs:           "Deleted logs",
						LogLines:       1,
						CollectionTime: metav1.Now(),
					},
					Analysis: diagnosisv1alpha1.DiagnosisAnalysis{
						Provider:        "openai",
						Model:           "gpt-4",
						Summary:         "Deleted analysis summary",
						RootCause:       "Deleted root cause",
						Recommendations: []string{"Deleted recommendation"},
						ProcessingTime:  "4.1s",
					},
				},
			}
			Expect(k8sClient.Create(ctx, report)).To(Succeed())

			// Wait for report to be available
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(report), &diagnosisv1alpha1.DiagnosisReport{})
			}, time.Second*5, time.Millisecond*100).Should(Succeed())

			// Delete the report
			Expect(k8sClient.Delete(ctx, report)).To(Succeed())

			// Wait for deletion to complete
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(report), &diagnosisv1alpha1.DiagnosisReport{})
				return err != nil
			}, time.Second*5, time.Millisecond*100).Should(BeTrue())

			// Reconcile should handle deleted report gracefully
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-report-delete",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))
		})
	})

	Context("When reconciling multiple reports", func() {
		var (
			report1 *diagnosisv1alpha1.DiagnosisReport
			report2 *diagnosisv1alpha1.DiagnosisReport
			ctx     context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()

			// Create first report
			report1 = &diagnosisv1alpha1.DiagnosisReport{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-report-1",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisReportSpec{
					TargetPod: diagnosisv1alpha1.PodReference{
						Name:      "test-pod-1",
						Namespace: "default",
						UID:       "test-uid-111",
						Image:     "nginx:1.20",
					},
					PolicyRef: diagnosisv1alpha1.PolicyReference{
						Name:      "test-policy-1",
						Namespace: "default",
					},
					TriggerCondition: diagnosisv1alpha1.TriggerCondition{
						Type:       "Failed",
						DetectedAt: metav1.Now(),
					},
					CollectedData: diagnosisv1alpha1.CollectedData{
						PodDescription: "First pod description",
						Logs:           "First pod logs",
						LogLines:       3,
						CollectionTime: metav1.Now(),
					},
					Analysis: diagnosisv1alpha1.DiagnosisAnalysis{
						Provider:        "ollama",
						Model:           "llama2",
						Summary:         "First analysis summary",
						RootCause:       "First root cause",
						Recommendations: []string{"First recommendation"},
						ProcessingTime:  "1.2s",
					},
				},
			}
			Expect(k8sClient.Create(ctx, report1)).To(Succeed())

			// Create second report
			report2 = &diagnosisv1alpha1.DiagnosisReport{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-report-2",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisReportSpec{
					TargetPod: diagnosisv1alpha1.PodReference{
						Name:      "test-pod-2",
						Namespace: "default",
						UID:       "test-uid-222",
						Image:     "postgres:13",
					},
					PolicyRef: diagnosisv1alpha1.PolicyReference{
						Name:      "test-policy-2",
						Namespace: "default",
					},
					TriggerCondition: diagnosisv1alpha1.TriggerCondition{
						Type:       "Pending",
						DetectedAt: metav1.Now(),
					},
					CollectedData: diagnosisv1alpha1.CollectedData{
						PodDescription: "Second pod description",
						Logs:           "Second pod logs",
						LogLines:       5,
						CollectionTime: metav1.Now(),
					},
					Analysis: diagnosisv1alpha1.DiagnosisAnalysis{
						Provider:        "openai",
						Model:           "gpt-3.5-turbo",
						Summary:         "Second analysis summary",
						RootCause:       "Second root cause",
						Recommendations: []string{"Second recommendation", "Additional fix"},
						ProcessingTime:  "2.7s",
					},
				},
			}
			Expect(k8sClient.Create(ctx, report2)).To(Succeed())

			// Wait for both reports to be available
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(report1), &diagnosisv1alpha1.DiagnosisReport{})
			}, time.Second*5, time.Millisecond*100).Should(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(report2), &diagnosisv1alpha1.DiagnosisReport{})
			}, time.Second*5, time.Millisecond*100).Should(Succeed())
		})

		AfterEach(func() {
			// Clean up resources
			Expect(k8sClient.Delete(ctx, report1)).To(Succeed())
			Expect(k8sClient.Delete(ctx, report2)).To(Succeed())
		})

		It("should handle multiple reports independently", func() {
			reconciler := &DiagnosisReportReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Reconcile first report
			req1 := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-report-1",
					Namespace: "default",
				},
			}

			result1, err1 := reconciler.Reconcile(ctx, req1)
			Expect(err1).NotTo(HaveOccurred())
			Expect(result1).To(Equal(reconcile.Result{}))

			// Reconcile second report
			req2 := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-report-2",
					Namespace: "default",
				},
			}

			result2, err2 := reconciler.Reconcile(ctx, req2)
			Expect(err2).NotTo(HaveOccurred())
			Expect(result2).To(Equal(reconcile.Result{}))

			// Both reports should still exist independently
			var updatedReport1 diagnosisv1alpha1.DiagnosisReport
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(report1), &updatedReport1)).To(Succeed())
			Expect(updatedReport1.Spec.Analysis.Summary).To(Equal("First analysis summary"))
			Expect(updatedReport1.Spec.TargetPod.Image).To(Equal("nginx:1.20"))

			var updatedReport2 diagnosisv1alpha1.DiagnosisReport
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(report2), &updatedReport2)).To(Succeed())
			Expect(updatedReport2.Spec.Analysis.Summary).To(Equal("Second analysis summary"))
			Expect(updatedReport2.Spec.TargetPod.Image).To(Equal("postgres:13"))
		})
	})
})
