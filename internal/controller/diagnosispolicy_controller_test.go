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

var _ = Describe("DiagnosisPolicy Controller", func() {
	Context("When reconciling a DiagnosisPolicy", func() {
		var (
			policy *diagnosisv1alpha1.DiagnosisPolicy
			ctx    context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()

			// Create a valid DiagnosisPolicy
			policy = &diagnosisv1alpha1.DiagnosisPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisPolicySpec{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
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
			Expect(k8sClient.Create(ctx, policy)).To(Succeed())

			// Wait for policy to be available
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(policy), &diagnosisv1alpha1.DiagnosisPolicy{})
			}, time.Second*5, time.Millisecond*100).Should(Succeed())
		})

		AfterEach(func() {
			// Clean up resources
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())
		})

		It("should successfully reconcile without errors", func() {
			reconciler := &DiagnosisPolicyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-policy",
					Namespace: "default",
				},
			}

			// Reconcile should not return errors
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			// Policy should still exist and be unchanged
			var updatedPolicy diagnosisv1alpha1.DiagnosisPolicy
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(policy), &updatedPolicy)).To(Succeed())
			Expect(updatedPolicy.Name).To(Equal("test-policy"))
		})
	})

	Context("When reconciling policy updates", func() {
		var (
			policy *diagnosisv1alpha1.DiagnosisPolicy
			ctx    context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()

			// Create a DiagnosisPolicy
			policy = &diagnosisv1alpha1.DiagnosisPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy-update",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisPolicySpec{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
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
			Expect(k8sClient.Create(ctx, policy)).To(Succeed())

			// Wait for policy to be available
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(policy), &diagnosisv1alpha1.DiagnosisPolicy{})
			}, time.Second*5, time.Millisecond*100).Should(Succeed())

			// Update the policy
			policy.Spec.LLMConfig.Model = "updated-model"
			Expect(k8sClient.Update(ctx, policy)).To(Succeed())
		})

		AfterEach(func() {
			// Clean up resources
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())
		})

		It("should handle policy updates without errors", func() {
			reconciler := &DiagnosisPolicyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-policy-update",
					Namespace: "default",
				},
			}

			// Reconcile should not return errors even after update
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			// Policy should still exist with updated values
			var updatedPolicy diagnosisv1alpha1.DiagnosisPolicy
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(policy), &updatedPolicy)).To(Succeed())
			Expect(updatedPolicy.Spec.LLMConfig.Model).To(Equal("updated-model"))
		})
	})

	Context("When handling non-existent policy", func() {
		It("should handle non-existent policy gracefully", func() {
			reconciler := &DiagnosisPolicyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "non-existent-policy",
					Namespace: "default",
				},
			}

			// Should not return error for non-existent policy
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))
		})
	})

	Context("When policy is deleted", func() {
		It("should handle policy deletion gracefully", func() {
			ctx := context.Background()
			reconciler := &DiagnosisPolicyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Create a policy
			policy := &diagnosisv1alpha1.DiagnosisPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy-delete",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisPolicySpec{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
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
			Expect(k8sClient.Create(ctx, policy)).To(Succeed())

			// Wait for policy to be available
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(policy), &diagnosisv1alpha1.DiagnosisPolicy{})
			}, time.Second*5, time.Millisecond*100).Should(Succeed())

			// Delete the policy
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())

			// Wait for deletion to complete
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(policy), &diagnosisv1alpha1.DiagnosisPolicy{})
				return err != nil
			}, time.Second*5, time.Millisecond*100).Should(BeTrue())

			// Reconcile should handle deleted policy gracefully
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-policy-delete",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))
		})
	})

	Context("When reconciling multiple policies", func() {
		var (
			policy1 *diagnosisv1alpha1.DiagnosisPolicy
			policy2 *diagnosisv1alpha1.DiagnosisPolicy
			ctx     context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()

			// Create first policy
			policy1 = &diagnosisv1alpha1.DiagnosisPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy-1",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisPolicySpec{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test1"},
					},
					TriggerConditions: []diagnosisv1alpha1.PodConditionTrigger{
						{Type: "Failed"},
					},
					LLMConfig: diagnosisv1alpha1.LLMConfiguration{
						Provider: "ollama",
						Model:    "test-model-1",
					},
				},
			}
			Expect(k8sClient.Create(ctx, policy1)).To(Succeed())

			// Create second policy
			policy2 = &diagnosisv1alpha1.DiagnosisPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy-2",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisPolicySpec{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test2"},
					},
					TriggerConditions: []diagnosisv1alpha1.PodConditionTrigger{
						{Type: "Pending"},
					},
					LLMConfig: diagnosisv1alpha1.LLMConfiguration{
						Provider: "openai",
						Model:    "gpt-4",
					},
				},
			}
			Expect(k8sClient.Create(ctx, policy2)).To(Succeed())

			// Wait for both policies to be available
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(policy1), &diagnosisv1alpha1.DiagnosisPolicy{})
			}, time.Second*5, time.Millisecond*100).Should(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(policy2), &diagnosisv1alpha1.DiagnosisPolicy{})
			}, time.Second*5, time.Millisecond*100).Should(Succeed())
		})

		AfterEach(func() {
			// Clean up resources
			Expect(k8sClient.Delete(ctx, policy1)).To(Succeed())
			Expect(k8sClient.Delete(ctx, policy2)).To(Succeed())
		})

		It("should handle multiple policies independently", func() {
			reconciler := &DiagnosisPolicyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Reconcile first policy
			req1 := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-policy-1",
					Namespace: "default",
				},
			}

			result1, err1 := reconciler.Reconcile(ctx, req1)
			Expect(err1).NotTo(HaveOccurred())
			Expect(result1).To(Equal(reconcile.Result{}))

			// Reconcile second policy
			req2 := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-policy-2",
					Namespace: "default",
				},
			}

			result2, err2 := reconciler.Reconcile(ctx, req2)
			Expect(err2).NotTo(HaveOccurred())
			Expect(result2).To(Equal(reconcile.Result{}))

			// Both policies should still exist independently
			var updatedPolicy1 diagnosisv1alpha1.DiagnosisPolicy
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(policy1), &updatedPolicy1)).To(Succeed())
			Expect(updatedPolicy1.Spec.LLMConfig.Model).To(Equal("test-model-1"))

			var updatedPolicy2 diagnosisv1alpha1.DiagnosisPolicy
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(policy2), &updatedPolicy2)).To(Succeed())
			Expect(updatedPolicy2.Spec.LLMConfig.Model).To(Equal("gpt-4"))
		})
	})
})
