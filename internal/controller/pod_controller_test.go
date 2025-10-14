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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

var _ = Describe("Pod Controller", func() {
	Context("When reconciling a healthy pod", func() {
		var (
			pod    *corev1.Pod
			policy *diagnosisv1alpha1.DiagnosisPolicy
			ctx    context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()

			// Create a healthy running pod
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "healthy-pod",
					Namespace: "default",
					Labels:    map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodReady, Status: corev1.ConditionTrue},
					},
				},
			}

			// Create policy that triggers on Failed pods
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

			Expect(k8sClient.Create(ctx, pod)).To(Succeed())
			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
		})

		AfterEach(func() {
			// Clean up resources
			Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())

			// Clean up any DiagnosisRequests
			var requests diagnosisv1alpha1.DiagnosisRequestList
			Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
			for _, req := range requests.Items {
				Expect(k8sClient.Delete(ctx, &req)).To(Succeed())
			}
		})

		It("should not create a DiagnosisRequest for healthy pod", func() {
			reconciler := &PodReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "healthy-pod",
					Namespace: "default",
				},
			}

			// Reconcile
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			// Verify no DiagnosisRequest was created
			var requests diagnosisv1alpha1.DiagnosisRequestList
			Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
			Expect(requests.Items).To(BeEmpty(), "No DiagnosisRequest should be created for healthy pod")
		})
	})

	Context("When reconciling a failed pod", func() {
		var (
			pod    *corev1.Pod
			policy *diagnosisv1alpha1.DiagnosisPolicy
			ctx    context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()

			// Create policy first and wait for it to be available
			policy = &diagnosisv1alpha1.DiagnosisPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "failed-policy",
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

			// Create a failed pod
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "failed-pod",
					Namespace: "default",
					Labels:    map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			// Wait for pod to be available
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), &corev1.Pod{})
			}, time.Second*5, time.Millisecond*100).Should(Succeed())

			// Update pod status to Failed
			pod.Status.Phase = corev1.PodFailed
			Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())

			// Wait for status update to propagate
			Eventually(func() corev1.PodPhase {
				var updatedPod corev1.Pod
				Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), &updatedPod)).To(Succeed())
				return updatedPod.Status.Phase
			}, time.Second*5, time.Millisecond*100).Should(Equal(corev1.PodFailed))
		})

		AfterEach(func() {
			// Clean up resources
			Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())

			// Clean up any DiagnosisRequests
			var requests diagnosisv1alpha1.DiagnosisRequestList
			Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
			for _, req := range requests.Items {
				Expect(k8sClient.Delete(ctx, &req)).To(Succeed())
			}
		})

		It("should create a DiagnosisRequest for failed pod", func() {
			reconciler := &PodReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "failed-pod",
					Namespace: "default",
				},
			}

			// Reconcile
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			// Verify DiagnosisRequest was created
			Eventually(func() []diagnosisv1alpha1.DiagnosisRequest {
				var requests diagnosisv1alpha1.DiagnosisRequestList
				Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
				return requests.Items
			}, time.Second*5, time.Millisecond*100).Should(HaveLen(1), "One DiagnosisRequest should be created")

			var requests diagnosisv1alpha1.DiagnosisRequestList
			Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
			request := requests.Items[0]

			Expect(request.Spec.TargetPod.Name).To(Equal("failed-pod"))
			Expect(request.Spec.PolicyRef.Name).To(Equal("failed-policy"))
			Expect(request.Spec.TriggerCondition.Type).To(Equal("Failed"))
		})
	})

	Context("When reconciling a pending pod with minimum duration", func() {
		var (
			pod    *corev1.Pod
			policy *diagnosisv1alpha1.DiagnosisPolicy
			ctx    context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()

			// Clean up any existing policies first to avoid interference
			var existingPolicies diagnosisv1alpha1.DiagnosisPolicyList
			Expect(k8sClient.List(ctx, &existingPolicies)).To(Succeed())
			for _, p := range existingPolicies.Items {
				if p.Namespace == "default" {
					_ = k8sClient.Delete(ctx, &p) // Ignore errors during cleanup
				}
			}

			// Clean up any existing pods
			var existingPods corev1.PodList
			Expect(k8sClient.List(ctx, &existingPods, client.InNamespace("default"))).To(Succeed())
			for _, p := range existingPods.Items {
				_ = k8sClient.Delete(ctx, &p) // Ignore errors during cleanup
			}

			// Wait a moment for cleanup to complete
			time.Sleep(100 * time.Millisecond)

			// Create policy first
			policy = &diagnosisv1alpha1.DiagnosisPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pending-policy",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisPolicySpec{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					TriggerConditions: []diagnosisv1alpha1.PodConditionTrigger{
						{
							Type:        "Pending",
							MinDuration: &metav1.Duration{Duration: 100 * time.Millisecond},
						},
					},
					LLMConfig: diagnosisv1alpha1.LLMConfiguration{
						Provider: "ollama",
						Model:    "test-model",
					},
				},
			}
			Expect(k8sClient.Create(ctx, policy)).To(Succeed())

			// Wait for policy to be available and verify it exists
			Eventually(func() error {
				var savedPolicy diagnosisv1alpha1.DiagnosisPolicy
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(policy), &savedPolicy)
			}, time.Second*5, time.Millisecond*100).Should(Succeed())

			// Create a pending pod
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pending-pod",
					Namespace: "default",
					Labels:    map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			// Wait for pod to be available
			Eventually(func() error {
				var savedPod corev1.Pod
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), &savedPod)
			}, time.Second*5, time.Millisecond*100).Should(Succeed())

			// Get the updated pod and set status
			var savedPod corev1.Pod
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), &savedPod)).To(Succeed())
			savedPod.Status.Phase = corev1.PodPending
			Expect(k8sClient.Status().Update(ctx, &savedPod)).To(Succeed())

			// Update local pod reference and wait for minimum duration
			pod = &savedPod
			time.Sleep(200 * time.Millisecond) // Wait longer than MinDuration

			// Wait for status update to propagate
			Eventually(func() corev1.PodPhase {
				var updatedPod corev1.Pod
				Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), &updatedPod)).To(Succeed())
				return updatedPod.Status.Phase
			}, time.Second*5, time.Millisecond*100).Should(Equal(corev1.PodPending))

		})

		AfterEach(func() {
			// Clean up resources
			Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())

			// Clean up any DiagnosisRequests
			var requests diagnosisv1alpha1.DiagnosisRequestList
			Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
			for _, req := range requests.Items {
				Expect(k8sClient.Delete(ctx, &req)).To(Succeed())
			}
		})

		It("should create DiagnosisRequest after minimum duration", func() {
			reconciler := &PodReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "pending-pod",
					Namespace: "default",
				},
			}

			// Reconcile
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			// Verify DiagnosisRequest was created
			Eventually(func() []diagnosisv1alpha1.DiagnosisRequest {
				var requests diagnosisv1alpha1.DiagnosisRequestList
				Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
				return requests.Items
			}, time.Second*5, time.Millisecond*100).Should(HaveLen(1), "One DiagnosisRequest should be created")

			var requests diagnosisv1alpha1.DiagnosisRequestList
			Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
			request := requests.Items[0]

			Expect(request.Spec.TargetPod.Name).To(Equal("pending-pod"))
			Expect(request.Spec.TriggerCondition.Type).To(Equal("Pending"))
		})
	})

	Context("When no matching policy exists", func() {
		var (
			pod *corev1.Pod
			ctx context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()

			// Create a failed pod
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "failed-pod-no-policy",
					Namespace: "default",
					Labels:    map[string]string{"app": "different"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
				},
			}

			Expect(k8sClient.Create(ctx, pod)).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
		})

		It("should not create DiagnosisRequest when no policy matches", func() {
			reconciler := &PodReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "failed-pod-no-policy",
					Namespace: "default",
				},
			}

			// Reconcile
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			// Verify no DiagnosisRequest was created
			Consistently(func() []diagnosisv1alpha1.DiagnosisRequest {
				var requests diagnosisv1alpha1.DiagnosisRequestList
				Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
				return requests.Items
			}, time.Second*2, time.Millisecond*100).Should(BeEmpty(), "No DiagnosisRequest should be created when no policy matches")
		})
	})

	Context("When pod doesn't exist", func() {
		It("should handle non-existent pod gracefully", func() {
			reconciler := &PodReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "non-existent-pod",
					Namespace: "default",
				},
			}

			// Should not return error for non-existent pod
			result, err := reconciler.Reconcile(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))
		})
	})

	Context("When preventing duplicate requests", func() {
		var (
			pod             *corev1.Pod
			policy          *diagnosisv1alpha1.DiagnosisPolicy
			existingRequest *diagnosisv1alpha1.DiagnosisRequest
			ctx             context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()

			// Create a failed pod
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "failed-pod-duplicate",
					Namespace: "default",
					Labels:    map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
				},
			}

			// Create policy
			policy = &diagnosisv1alpha1.DiagnosisPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-policy",
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

			// Create existing active DiagnosisRequest
			existingRequest = &diagnosisv1alpha1.DiagnosisRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "existing-request",
					Namespace: "default",
				},
				Spec: diagnosisv1alpha1.DiagnosisRequestSpec{
					Type: diagnosisv1alpha1.DiagnosisRequestTypeAutomatic,
					TargetPod: diagnosisv1alpha1.TargetPodReference{
						Name:      "failed-pod-duplicate",
						Namespace: "default",
					},
					PolicyRef: diagnosisv1alpha1.PolicyReference{
						Name:      "duplicate-policy",
						Namespace: "default",
					},
					TriggerCondition: &diagnosisv1alpha1.TriggerCondition{
						Type:       "Failed",
						DetectedAt: metav1.Now(),
					},
				},
				Status: diagnosisv1alpha1.DiagnosisRequestStatus{
					Phase: diagnosisv1alpha1.DiagnosisRequestPhasePending,
				},
			}

			Expect(k8sClient.Create(ctx, pod)).To(Succeed())
			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
			Expect(k8sClient.Create(ctx, existingRequest)).To(Succeed())
		})

		AfterEach(func() {
			// Clean up resources
			Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())
			Expect(k8sClient.Delete(ctx, existingRequest)).To(Succeed())

			// Clean up any additional DiagnosisRequests
			var requests diagnosisv1alpha1.DiagnosisRequestList
			Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
			for _, req := range requests.Items {
				if req.Name != "existing-request" {
					Expect(k8sClient.Delete(ctx, &req)).To(Succeed())
				}
			}
		})

		It("should not create duplicate DiagnosisRequest", func() {
			reconciler := &PodReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "failed-pod-duplicate",
					Namespace: "default",
				},
			}

			// Reconcile
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			// Verify only one DiagnosisRequest exists (the original one)
			Consistently(func() int {
				var requests diagnosisv1alpha1.DiagnosisRequestList
				Expect(k8sClient.List(ctx, &requests, client.InNamespace("default"))).To(Succeed())
				return len(requests.Items)
			}, time.Second*2, time.Millisecond*100).Should(Equal(1), "Should not create duplicate DiagnosisRequest")
		})
	})
})
