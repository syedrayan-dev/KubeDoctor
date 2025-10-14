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
	"time"

	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Kubernetes Handler", func() {
	var (
		handler *KubernetesHandler
		ctx     context.Context
	)

	BeforeEach(func() {
		handler = NewKubernetesHandler(k8sClient)
		ctx = context.Background()

		// Create 2 test namespaces with unique names
		timestamp := time.Now().UnixNano()
		for i := 1; i <= 2; i++ {
			testNamespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("test-ns-%d-%d", i, timestamp),
				},
			}
			Expect(k8sClient.Create(ctx, testNamespace)).To(Succeed())
		}

		// Create 3 test pods in default namespace
		for i := 1; i <= 3; i++ {
			testPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("test-pod-%d", i),
					Namespace: "default",
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
			Expect(k8sClient.Create(ctx, testPod)).To(Succeed())
		}
	})

	AfterEach(func() {
		// Clean up test namespaces (custom namespaces with timestamp pattern)
		var namespaces corev1.NamespaceList
		if err := k8sClient.List(ctx, &namespaces); err == nil {
			for _, ns := range namespaces.Items {
				if len(ns.Name) > 7 && ns.Name[:7] == "test-ns" {
					_ = k8sClient.Delete(ctx, &ns)
				}
			}
		}

		// Clean up test pods
		var pods corev1.PodList
		if err := k8sClient.List(ctx, &pods); err == nil {
			for _, pod := range pods.Items {
				_ = k8sClient.Delete(ctx, &pod)
			}
		}
	})

	Context("When listing namespaces", func() {
		It("should return namespaces list successfully", func() {
			req := httptest.NewRequest("GET", "/api/v1/namespaces", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.HandleNamespaces(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response corev1.NamespaceList
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			// Should include default namespace + 2 test namespaces (at least 3)
			Expect(len(response.Items)).To(BeNumerically(">=", 3))

			// Verify test namespaces exist (check for timestamp-based names)
			namespaceNames := make([]string, len(response.Items))
			testNamespaceCount := 0
			for i, ns := range response.Items {
				namespaceNames[i] = ns.Name
				if len(ns.Name) > 7 && ns.Name[:7] == "test-ns" {
					testNamespaceCount++
				}
			}
			Expect(testNamespaceCount).To(Equal(2))
		})
	})

	Context("When listing pods in namespace", func() {
		It("should return pods list successfully", func() {
			req := httptest.NewRequest("GET", "/api/v1/namespaces/default/pods", nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{"namespace": "default"})
			rec := httptest.NewRecorder()

			handler.HandlePods(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response corev1.PodList
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Items).To(HaveLen(3))

			// Verify test pods exist
			podNames := make([]string, len(response.Items))
			for i, pod := range response.Items {
				podNames[i] = pod.Name
			}
			Expect(podNames).To(ContainElement("test-pod-1"))
			Expect(podNames).To(ContainElement("test-pod-2"))
			Expect(podNames).To(ContainElement("test-pod-3"))
		})
	})
})
