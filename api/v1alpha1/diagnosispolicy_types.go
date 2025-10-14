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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiagnosisPolicySpec defines the desired state of DiagnosisPolicy.
type DiagnosisPolicySpec struct {
	// TargetNamespaces specifies the namespaces this policy applies to.
	// If not specified, applies to the same namespace as the policy.
	// +optional
	TargetNamespaces []string `json:"targetNamespaces,omitempty"`

	// PodSelector selects the pods this policy applies to.
	// If not specified, applies to all pods in the target namespaces.
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`

	// TriggerConditions defines which pod conditions trigger diagnosis
	TriggerConditions []PodConditionTrigger `json:"triggerConditions"`

	// LLMConfig specifies the LLM configuration for analysis
	LLMConfig LLMConfiguration `json:"llmConfig"`

	// DiagnosisSettings controls diagnosis behavior
	// +optional
	DiagnosisSettings *DiagnosisSettings `json:"diagnosisSettings,omitempty"`
}

// PodConditionTrigger defines conditions that trigger diagnosis based on Pod Phase and Conditions
type PodConditionTrigger struct {
	// Type specifies the pod phase that triggers diagnosis
	// Valid values: "Failed", "Pending", "Unknown", "Running"
	// +kubebuilder:validation:Enum=Failed;Pending;Unknown;Running
	Type string `json:"type"`

	// MinDuration specifies minimum duration the condition must persist before triggering
	// +optional
	MinDuration *metav1.Duration `json:"minDuration,omitempty"`

	// Conditions specifies additional pod conditions to check (required when Type is "Running")
	// This allows fine-grained detection based on Pod Conditions like Ready, ContainersReady, etc.
	// +optional
	Conditions []PodConditionCheck `json:"conditions,omitempty"`
}

// PodConditionCheck defines a specific pod condition to verify
type PodConditionCheck struct {
	// Name of the pod condition
	// +kubebuilder:validation:Enum=Ready;ContainersReady;PodScheduled;Initialized;PodReadyToStartContainers;DisruptionTarget
	Name string `json:"name"`

	// Status expected for the condition to trigger diagnosis
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status string `json:"status"`

	// MinDuration for this specific condition (overrides parent minDuration if specified)
	// +optional
	MinDuration *metav1.Duration `json:"minDuration,omitempty"`
}

// LLMConfiguration defines LLM settings
type LLMConfiguration struct {
	// Provider specifies the LLM provider (e.g., "openai", "ollama", "azure-openai")
	Provider string `json:"provider"`

	// Model specifies the LLM model to use (e.g., "gpt-4", "llama3.2", "mistral")
	Model string `json:"model"`

	// APIKeySecretRef references a secret containing the API key
	// Required for providers like OpenAI, optional for self-hosted providers like Ollama
	// +optional
	APIKeySecretRef *SecretReference `json:"apiKeySecretRef,omitempty"`

	// BaseURL specifies custom API base URL (optional for custom deployments)
	// For Ollama, this would be the Ollama server URL (e.g., "http://ollama-service:11434")
	// +optional
	BaseURL string `json:"baseURL,omitempty"`
}

// SecretReference represents a reference to a secret
type SecretReference struct {
	// Name of the secret
	Name string `json:"name"`

	// Key in the secret containing the API key
	Key string `json:"key"`

	// Namespace of the secret (if different from policy namespace)
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// DiagnosisSettings controls diagnosis behavior
type DiagnosisSettings struct {
	// MaxLogLines specifies maximum number of log lines to collect for analysis
	// +optional
	// +kubebuilder:default=100
	MaxLogLines *int `json:"maxLogLines,omitempty"`

	// RetryInterval specifies how long to wait before retrying diagnosis for the same pod
	// +optional
	// +kubebuilder:default="5m"
	RetryInterval *metav1.Duration `json:"retryInterval,omitempty"`
}

// DiagnosisPolicyStatus defines the observed state of DiagnosisPolicy.
type DiagnosisPolicyStatus struct {
	// Conditions represent the latest available observations of the policy's current state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed DiagnosisPolicy.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// LastDiagnosisTime is the last time a diagnosis was triggered by this policy
	// +optional
	LastDiagnosisTime *metav1.Time `json:"lastDiagnosisTime,omitempty"`

	// TotalDiagnosisCount is the total number of diagnoses triggered by this policy
	// +optional
	TotalDiagnosisCount int64 `json:"totalDiagnosisCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DiagnosisPolicy is the Schema for the diagnosispolicies API.
type DiagnosisPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiagnosisPolicySpec   `json:"spec,omitempty"`
	Status DiagnosisPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DiagnosisPolicyList contains a list of DiagnosisPolicy.
type DiagnosisPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiagnosisPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiagnosisPolicy{}, &DiagnosisPolicyList{})
}
