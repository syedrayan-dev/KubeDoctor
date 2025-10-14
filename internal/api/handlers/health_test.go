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
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Health Handler", func() {
	var (
		handler *HealthHandler
	)

	BeforeEach(func() {
		handler = NewHealthHandler(k8sClient)
	})

	Context("When checking health status", func() {
		It("should return successful health response", func() {
			req := httptest.NewRequest("GET", "/health", nil)
			rec := httptest.NewRecorder()

			handler.HandleHealth(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Content-Type")).To(Equal("application/json"))

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveKey("status"))
			Expect(response["status"]).NotTo(BeEmpty())
			Expect(response).To(HaveKey("timestamp"))
			Expect(response["timestamp"]).NotTo(BeEmpty())
		})
	})
})
