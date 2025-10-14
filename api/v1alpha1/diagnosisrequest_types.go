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

// DiagnosisRequestSpec defines the desired state of DiagnosisRequest.
type DiagnosisRequestSpec struct {
	// Type indicates whether this is an automatic or manual diagnosis request
	// +kubebuilder:validation:Enum=Automatic;Manual
	Type DiagnosisRequestType `json:"type"`

	// TargetPod specifies the pod to diagnose
	TargetPod TargetPodReference `json:"targetPod"`

	// PolicyRef references the DiagnosisPolicy to use
	PolicyRef PolicyReference `json:"policyRef"`

	// TriggerCondition contains trigger information for automatic diagnosis
	// Required when Type is "Automatic"
	// +optional
	TriggerCondition *TriggerCondition `json:"triggerCondition,omitempty"`
}

// DiagnosisRequestType represents the type of diagnosis request
type DiagnosisRequestType string

const (
	// DiagnosisRequestTypeAutomatic indicates an automatic diagnosis triggered by PodWatcher
	DiagnosisRequestTypeAutomatic DiagnosisRequestType = "Automatic"

	// DiagnosisRequestTypeManual indicates a manual diagnosis requested by user
	DiagnosisRequestTypeManual DiagnosisRequestType = "Manual"
)

// TargetPodReference contains minimal information to identify a pod
type TargetPodReference struct {
	// Name of the pod
	Name string `json:"name"`

	// Namespace of the pod
	Namespace string `json:"namespace"`
}

// DiagnosisRequestStatus defines the observed state of DiagnosisRequest.
type DiagnosisRequestStatus struct {
	// Phase represents the current phase of the diagnosis request
	// +kubebuilder:validation:Enum=Pending;InProgress;Completed;Failed
	Phase DiagnosisRequestPhase `json:"phase,omitempty"`

	// Message contains a human-readable message about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// LastUpdateTime is when the status was last updated
	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// DiagnosisReportRef references the created DiagnosisReport
	// +optional
	DiagnosisReportRef *DiagnosisReportReference `json:"diagnosisReportRef,omitempty"`

	// RequestTime is when the request was created
	// +optional
	RequestTime *metav1.Time `json:"requestTime,omitempty"`

	// StartTime is when the diagnosis started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the diagnosis completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Error contains error information if the request failed
	// +optional
	Error *DiagnosisError `json:"error,omitempty"`
}

// DiagnosisRequestPhase represents the phase of a diagnosis request
type DiagnosisRequestPhase string

const (
	// DiagnosisRequestPhasePending means the request is waiting to be processed
	DiagnosisRequestPhasePending DiagnosisRequestPhase = "Pending"

	// DiagnosisRequestPhaseInProgress means the diagnosis is running
	DiagnosisRequestPhaseInProgress DiagnosisRequestPhase = "InProgress"

	// DiagnosisRequestPhaseCompleted means the diagnosis completed successfully
	DiagnosisRequestPhaseCompleted DiagnosisRequestPhase = "Completed"

	// DiagnosisRequestPhaseFailed means the diagnosis failed
	DiagnosisRequestPhaseFailed DiagnosisRequestPhase = "Failed"
)

// DiagnosisReportReference contains reference to the created report
type DiagnosisReportReference struct {
	// Name of the DiagnosisReport
	Name string `json:"name"`

	// Namespace of the DiagnosisReport
	Namespace string `json:"namespace"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DiagnosisRequest is the Schema for the diagnosisrequests API.
type DiagnosisRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiagnosisRequestSpec   `json:"spec,omitempty"`
	Status DiagnosisRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DiagnosisRequestList contains a list of DiagnosisRequest.
type DiagnosisRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiagnosisRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiagnosisRequest{}, &DiagnosisRequestList{})
}
