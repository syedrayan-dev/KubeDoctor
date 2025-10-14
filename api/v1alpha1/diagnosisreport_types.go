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

// DiagnosisReportSpec defines the desired state of DiagnosisReport.
type DiagnosisReportSpec struct {
	// TargetPod contains information about the diagnosed pod
	TargetPod PodReference `json:"targetPod"`

	// PolicyRef references the DiagnosisPolicy that triggered this diagnosis
	PolicyRef PolicyReference `json:"policyRef"`

	// TriggerCondition describes what condition triggered the diagnosis
	TriggerCondition TriggerCondition `json:"triggerCondition"`

	// CollectedData contains the raw data collected for diagnosis
	CollectedData CollectedData `json:"collectedData"`

	// Analysis contains the LLM analysis results
	Analysis DiagnosisAnalysis `json:"analysis"`
}

// PodReference contains information about the target pod
type PodReference struct {
	// Name of the pod
	Name string `json:"name"`

	// Namespace of the pod
	Namespace string `json:"namespace"`

	// UID of the pod for unique identification
	UID string `json:"uid"`

	// Image contains the main container image information
	Image string `json:"image"`
}

// PolicyReference contains information about the diagnosis policy
type PolicyReference struct {
	// Name of the DiagnosisPolicy
	Name string `json:"name"`

	// Namespace of the DiagnosisPolicy
	Namespace string `json:"namespace"`
}

// TriggerCondition describes the condition that triggered the diagnosis
type TriggerCondition struct {
	// Type of the condition (e.g., "CrashLoopBackOff", "ImagePullBackOff")
	Type string `json:"type"`

	// DetectedAt is when the condition was detected
	DetectedAt metav1.Time `json:"detectedAt"`
}

// CollectedData contains the raw diagnostic data
type CollectedData struct {
	// PodDescription contains the output of kubectl describe pod
	PodDescription string `json:"podDescription"`

	// Logs contains the recent pod logs
	Logs string `json:"logs"`

	// LogLines is the number of log lines collected
	LogLines int `json:"logLines"`

	// CollectionTime is when the data was collected
	CollectionTime metav1.Time `json:"collectionTime"`
}

// DiagnosisAnalysis contains the LLM analysis results
type DiagnosisAnalysis struct {
	// Provider is the LLM provider used (e.g., "openai")
	Provider string `json:"provider"`

	// Model is the LLM model used (e.g., "gpt-4")
	Model string `json:"model"`

	// Summary is a brief summary of the problem
	Summary string `json:"summary"`

	// RootCause describes the identified root cause
	RootCause string `json:"rootCause"`

	// Recommendations contains suggested solutions
	Recommendations []string `json:"recommendations"`

	// ProcessingTime is how long the LLM analysis took
	ProcessingTime string `json:"processingTime"`
}

// DiagnosisReportStatus defines the observed state of DiagnosisReport.
type DiagnosisReportStatus struct {
	// Phase represents the current phase of the diagnosis
	// +kubebuilder:validation:Enum=Pending;InProgress;Completed;Failed
	Phase DiagnosisPhase `json:"phase,omitempty"`

	// Conditions represent the latest available observations of the diagnosis state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// StartTime is when the diagnosis started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the diagnosis completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Error contains error information if the diagnosis failed
	// +optional
	Error *DiagnosisError `json:"error,omitempty"`
}

// DiagnosisPhase represents the phase of a diagnosis
type DiagnosisPhase string

const (
	// DiagnosisPhasePending means the diagnosis is waiting to start
	DiagnosisPhasePending DiagnosisPhase = "Pending"

	// DiagnosisPhaseInProgress means the diagnosis is currently running
	DiagnosisPhaseInProgress DiagnosisPhase = "InProgress"

	// DiagnosisPhaseCompleted means the diagnosis completed successfully
	DiagnosisPhaseCompleted DiagnosisPhase = "Completed"

	// DiagnosisPhaseFailed means the diagnosis failed
	DiagnosisPhaseFailed DiagnosisPhase = "Failed"
)

// DiagnosisError contains error information
type DiagnosisError struct {
	// Message is a human readable error message
	Message string `json:"message"`

	// Reason is a machine readable error reason
	Reason string `json:"reason"`

	// RetryCount is the number of retry attempts made
	// +optional
	RetryCount int `json:"retryCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DiagnosisReport is the Schema for the diagnosisreports API.
type DiagnosisReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiagnosisReportSpec   `json:"spec,omitempty"`
	Status DiagnosisReportStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DiagnosisReportList contains a list of DiagnosisReport.
type DiagnosisReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiagnosisReport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiagnosisReport{}, &DiagnosisReportList{})
}
