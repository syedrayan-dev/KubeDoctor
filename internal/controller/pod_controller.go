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
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update
// +kubebuilder:rbac:groups=diagnosis.apollo.dev,resources=diagnosispolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups=diagnosis.apollo.dev,resources=diagnosisrequests,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Step 1: Fetch the Pod
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		// Pod might have been deleted, ignore
		log.V(1).Info("unable to fetch Pod", "error", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Log basic pod information
	log.Info("Processing pod",
		"namespace", pod.Namespace,
		"name", pod.Name,
		"phase", pod.Status.Phase,
		"conditions", len(pod.Status.Conditions))

	// Step 2: Check if pod is in a problematic state
	problemType := r.getPodProblemType(&pod)
	if problemType == "" {
		// Pod is healthy, nothing to do
		return ctrl.Result{}, nil
	}

	log.Info("Pod is in problematic state",
		"namespace", pod.Namespace,
		"name", pod.Name,
		"problemType", problemType)

	// Step 3: Find matching DiagnosisPolicy
	policy, err := r.findMatchingPolicy(ctx, &pod)
	if err != nil {
		log.Error(err, "Failed to find matching policy")
		return ctrl.Result{}, err
	}
	if policy == nil {
		log.V(1).Info("No matching DiagnosisPolicy found for pod",
			"namespace", pod.Namespace,
			"name", pod.Name)
		return ctrl.Result{}, nil
	}

	log.Info("Found matching DiagnosisPolicy",
		"policyName", policy.Name,
		"policyNamespace", policy.Namespace)

	// Step 4: Check if diagnosis already exists or in progress
	exists, err := r.diagnosisRequestExists(ctx, &pod, policy)
	if err != nil {
		log.Error(err, "Failed to check existing diagnosis request")
		return ctrl.Result{}, err
	}
	if exists {
		log.V(1).Info("Diagnosis request already exists for this pod",
			"namespace", pod.Namespace,
			"name", pod.Name)
		return ctrl.Result{}, nil
	}

	// Step 5: Create DiagnosisRequest
	if err := r.createDiagnosisRequest(ctx, &pod, policy, problemType); err != nil {
		log.Error(err, "Failed to create diagnosis request")
		return ctrl.Result{}, err
	}

	log.Info("Created DiagnosisRequest for pod",
		"namespace", pod.Namespace,
		"name", pod.Name)

	return ctrl.Result{}, nil
}

// getPodProblemType returns the Pod Phase for policy matching
func (r *PodReconciler) getPodProblemType(pod *corev1.Pod) string {
	// Return the current Pod Phase - policy will determine if it's problematic
	return string(pod.Status.Phase)
}

// findMatchingPolicy finds a DiagnosisPolicy that applies to the given pod
func (r *PodReconciler) findMatchingPolicy(ctx context.Context, pod *corev1.Pod) (*diagnosisv1alpha1.DiagnosisPolicy, error) {
	// List all DiagnosisPolicies
	var policies diagnosisv1alpha1.DiagnosisPolicyList
	if err := r.List(ctx, &policies); err != nil {
		return nil, err
	}

	// Check each policy
	for _, policy := range policies.Items {
		// Check if policy matches pod namespace
		if r.policyMatchesPod(&policy, pod) {
			return &policy, nil
		}
	}

	return nil, nil
}

// policyMatchesPod checks if a policy applies to a pod and its current state triggers diagnosis
func (r *PodReconciler) policyMatchesPod(policy *diagnosisv1alpha1.DiagnosisPolicy, pod *corev1.Pod) bool {
	// Step 1: Check basic policy scope (namespace, selector)
	if !r.policyMatchesPodScope(policy, pod) {
		return false
	}

	// Step 2: Check if current pod state matches any trigger conditions
	currentPhase := string(pod.Status.Phase)
	return r.podStateMatchesTriggers(policy, pod, currentPhase)
}

// policyMatchesPodScope checks if policy scope matches the pod (namespace, selector)
func (r *PodReconciler) policyMatchesPodScope(policy *diagnosisv1alpha1.DiagnosisPolicy, pod *corev1.Pod) bool {
	// Check target namespaces
	if len(policy.Spec.TargetNamespaces) > 0 {
		// Check if pod namespace is in target namespaces
		namespaceMatches := false
		for _, ns := range policy.Spec.TargetNamespaces {
			if pod.Namespace == ns {
				namespaceMatches = true
				break
			}
		}
		if !namespaceMatches {
			return false
		}
	} else {
		// No target namespaces specified, policy applies to its own namespace
		if policy.Namespace != pod.Namespace {
			return false
		}
	}

	// Check pod selector
	if policy.Spec.PodSelector != nil {
		selector, err := metav1.LabelSelectorAsSelector(policy.Spec.PodSelector)
		if err != nil {
			return false
		}
		if !selector.Matches(labels.Set(pod.Labels)) {
			return false
		}
	}
	// No selector means all pods in namespace

	return true
}

// podStateMatchesTriggers checks if pod's current state matches any trigger conditions (OR logic)
func (r *PodReconciler) podStateMatchesTriggers(policy *diagnosisv1alpha1.DiagnosisPolicy, pod *corev1.Pod, currentPhase string) bool {
	for _, trigger := range policy.Spec.TriggerConditions {
		if trigger.Type == currentPhase {
			// Type matches, now check conditions (AND logic within each trigger)
			if r.triggerConditionsMatch(trigger, pod) {
				// Check minimum duration requirement
				if r.checkMinDurationRequirement(trigger, pod, currentPhase) {
					return true // OR logic: one matching trigger is enough
				}
			}
		}
	}
	return false
}

// triggerConditionsMatch checks if all conditions within a trigger are satisfied (AND logic)
func (r *PodReconciler) triggerConditionsMatch(trigger diagnosisv1alpha1.PodConditionTrigger, pod *corev1.Pod) bool {
	// If no conditions specified, Type match is sufficient
	if len(trigger.Conditions) == 0 {
		return true
	}

	// All conditions must be satisfied (AND logic)
	for _, conditionCheck := range trigger.Conditions {
		if !r.podConditionMatches(pod, conditionCheck) {
			return false
		}
	}
	return true
}

// podConditionMatches checks if a specific pod condition matches the expected state
func (r *PodReconciler) podConditionMatches(pod *corev1.Pod, check diagnosisv1alpha1.PodConditionCheck) bool {
	for _, condition := range pod.Status.Conditions {
		if string(condition.Type) == check.Name {
			return string(condition.Status) == check.Status
		}
	}
	// If condition not found, consider it as not matching
	return false
}

// checkMinDurationRequirement checks if the pod has been in the current state long enough
func (r *PodReconciler) checkMinDurationRequirement(trigger diagnosisv1alpha1.PodConditionTrigger, pod *corev1.Pod, currentPhase string) bool {
	// If no minDuration specified, trigger immediately
	if trigger.MinDuration == nil {
		return true
	}

	// Calculate how long pod has been in current state
	timeInCurrentState := r.getTimeInCurrentState(pod, currentPhase, trigger)
	if timeInCurrentState == nil {
		// Conservative approach: if we can't determine the exact time, don't trigger
		return false
	}

	return *timeInCurrentState >= trigger.MinDuration.Duration
}

// getTimeInCurrentState calculates how long the pod has been in its current problematic state
// Returns nil if the exact time cannot be determined reliably
func (r *PodReconciler) getTimeInCurrentState(pod *corev1.Pod, currentPhase string, trigger diagnosisv1alpha1.PodConditionTrigger) *time.Duration {
	now := time.Now()

	switch currentPhase {
	case "Pending":
		// Pod has been Pending since creation - this is reliable
		duration := now.Sub(pod.CreationTimestamp.Time)
		return &duration
	case "Failed", "Unknown":
		// Try to find the exact transition time for Failed/Unknown phase
		if phaseTransitionTime := r.getPhaseTransitionTime(pod, currentPhase); !phaseTransitionTime.IsZero() {
			duration := now.Sub(phaseTransitionTime)
			return &duration
		}
		// Return nil if we can't determine the exact transition time
		return nil
	case "Running":
		// For Running pods with conditions, check specific condition transition time
		if len(trigger.Conditions) > 0 {
			if conditionDuration := r.getConditionDuration(pod, trigger.Conditions[0]); conditionDuration > 0 {
				return &conditionDuration
			}
			// Return nil if condition is not found or duration is 0
			return nil
		}
		// Running without conditions - since pod started running (reliable)
		duration := r.getTimeSinceRunning(pod)
		return &duration
	}

	return nil
}

// getConditionDuration gets how long a specific condition has been in its current state
func (r *PodReconciler) getConditionDuration(pod *corev1.Pod, conditionCheck diagnosisv1alpha1.PodConditionCheck) time.Duration {
	for _, condition := range pod.Status.Conditions {
		if string(condition.Type) == conditionCheck.Name && string(condition.Status) == conditionCheck.Status {
			return time.Since(condition.LastTransitionTime.Time)
		}
	}
	return 0
}

// getTimeSinceRunning calculates how long the pod has been in Running phase
func (r *PodReconciler) getTimeSinceRunning(pod *corev1.Pod) time.Duration {
	// Use the Ready condition transition time as approximation for when pod started running
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return time.Since(condition.LastTransitionTime.Time)
		}
	}
	// Fallback to creation time
	return time.Since(pod.CreationTimestamp.Time)
}

// getPhaseTransitionTime attempts to find when the pod transitioned to the current phase
func (r *PodReconciler) getPhaseTransitionTime(pod *corev1.Pod, currentPhase string) time.Time {
	// For Failed phase, look for container termination or pod failure indicators
	if currentPhase == "Failed" {
		// Check container statuses for termination time
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Terminated != nil {
				return containerStatus.State.Terminated.FinishedAt.Time
			}
		}
		// Check init container statuses
		for _, initContainerStatus := range pod.Status.InitContainerStatuses {
			if initContainerStatus.State.Terminated != nil && initContainerStatus.State.Terminated.ExitCode != 0 {
				return initContainerStatus.State.Terminated.FinishedAt.Time
			}
		}
	}

	// For Unknown phase, look for relevant condition transitions
	if currentPhase == "Unknown" {
		// Look for conditions that indicate the pod became unknown
		for _, condition := range pod.Status.Conditions {
			if condition.Status == corev1.ConditionUnknown {
				return condition.LastTransitionTime.Time
			}
		}
	}

	// For any phase, try to find the most recent condition change that might indicate phase transition
	var latestTransitionTime time.Time
	for _, condition := range pod.Status.Conditions {
		if condition.LastTransitionTime.After(latestTransitionTime) {
			latestTransitionTime = condition.LastTransitionTime.Time
		}
	}

	// Only return the latest transition time if it's not the pod creation time
	// (to avoid false positives where conditions were set at creation)
	if !latestTransitionTime.IsZero() && latestTransitionTime.After(pod.CreationTimestamp.Add(5*time.Second)) {
		return latestTransitionTime
	}

	// Return zero time if we can't determine the transition time
	return time.Time{}
}

// diagnosisRequestExists checks if a diagnosis request already exists for this pod
func (r *PodReconciler) diagnosisRequestExists(ctx context.Context, pod *corev1.Pod, policy *diagnosisv1alpha1.DiagnosisPolicy) (bool, error) {
	// List DiagnosisRequests in the policy namespace
	var requests diagnosisv1alpha1.DiagnosisRequestList
	if err := r.List(ctx, &requests, client.InNamespace(policy.Namespace)); err != nil {
		return false, err
	}

	currentProblemType := r.getPodProblemType(pod)

	// Check each request for conflicts
	for _, req := range requests.Items {
		if !r.requestMatchesPodAndPolicy(&req, pod, policy) {
			continue
		}

		if exists := r.checkRequestConflict(ctx, &req, policy, currentProblemType); exists {
			return true, nil
		}
	}

	return false, nil
}

// requestMatchesPodAndPolicy checks if a diagnosis request matches the given pod and policy
func (r *PodReconciler) requestMatchesPodAndPolicy(req *diagnosisv1alpha1.DiagnosisRequest, pod *corev1.Pod, policy *diagnosisv1alpha1.DiagnosisPolicy) bool {
	return req.Spec.TargetPod.Name == pod.Name &&
		req.Spec.TargetPod.Namespace == pod.Namespace &&
		req.Spec.PolicyRef.Name == policy.Name &&
		req.Spec.PolicyRef.Namespace == policy.Namespace
}

// checkRequestConflict checks if an existing request conflicts with creating a new one
func (r *PodReconciler) checkRequestConflict(ctx context.Context, req *diagnosisv1alpha1.DiagnosisRequest, policy *diagnosisv1alpha1.DiagnosisPolicy, currentProblemType string) bool {
	log := logf.FromContext(ctx)

	// Check if request is actively running
	if r.isRequestActive(req) {
		log.V(1).Info("Found active diagnosis request for pod",
			"requestName", req.Name,
			"phase", req.Status.Phase)
		return true
	}

	// Skip if request is not in a final state
	if !r.isRequestInFinalState(req) {
		return false
	}

	// Check general cooldown period
	if inCooldown := r.checkGeneralCooldown(req, policy); inCooldown {
		log.V(1).Info("Found recent diagnosis request within cooldown period",
			"requestName", req.Name,
			"phase", req.Status.Phase)
		return true
	}

	// Check same problem type cooldown
	if inCooldown := r.checkSameProblemTypeCooldown(req, currentProblemType); inCooldown {
		log.V(1).Info("Found recent diagnosis request for same problem type",
			"requestName", req.Name,
			"problemType", currentProblemType)
		return true
	}

	return false
}

// isRequestActive checks if a diagnosis request is currently active
func (r *PodReconciler) isRequestActive(req *diagnosisv1alpha1.DiagnosisRequest) bool {
	return req.Status.Phase == diagnosisv1alpha1.DiagnosisRequestPhasePending ||
		req.Status.Phase == diagnosisv1alpha1.DiagnosisRequestPhaseInProgress
}

// isRequestInFinalState checks if a diagnosis request is in a final state
func (r *PodReconciler) isRequestInFinalState(req *diagnosisv1alpha1.DiagnosisRequest) bool {
	return req.Status.Phase == diagnosisv1alpha1.DiagnosisRequestPhaseCompleted ||
		req.Status.Phase == diagnosisv1alpha1.DiagnosisRequestPhaseFailed
}

// checkGeneralCooldown checks if a request is within the general cooldown period
func (r *PodReconciler) checkGeneralCooldown(req *diagnosisv1alpha1.DiagnosisRequest, policy *diagnosisv1alpha1.DiagnosisPolicy) bool {
	if req.Status.CompletionTime == nil {
		return false
	}

	// Default cooldown period (can be overridden by policy settings)
	defaultCooldown := 10 * time.Minute
	cooldownPeriod := defaultCooldown
	if policy.Spec.DiagnosisSettings != nil && policy.Spec.DiagnosisSettings.RetryInterval != nil {
		cooldownPeriod = policy.Spec.DiagnosisSettings.RetryInterval.Duration
	}

	timeSinceCompletion := metav1.Now().Sub(req.Status.CompletionTime.Time)
	return timeSinceCompletion < cooldownPeriod
}

// checkSameProblemTypeCooldown checks if a request for the same problem type is within the short cooldown period
func (r *PodReconciler) checkSameProblemTypeCooldown(req *diagnosisv1alpha1.DiagnosisRequest, currentProblemType string) bool {
	if req.Spec.TriggerCondition == nil || req.Spec.TriggerCondition.Type != currentProblemType {
		return false
	}

	if req.Status.CompletionTime == nil {
		return false
	}

	shortCooldown := 5 * time.Minute
	timeSinceCompletion := metav1.Now().Sub(req.Status.CompletionTime.Time)
	return timeSinceCompletion < shortCooldown
}

// createDiagnosisRequest creates a new DiagnosisRequest
func (r *PodReconciler) createDiagnosisRequest(ctx context.Context, pod *corev1.Pod, policy *diagnosisv1alpha1.DiagnosisPolicy, problemType string) error {
	// Generate unique name for the request
	// Convert problemType to lowercase for Kubernetes naming compliance
	problemTypeLower := strings.ToLower(problemType)
	requestName := fmt.Sprintf("%s-%s-%d", pod.Name, problemTypeLower, time.Now().Unix())
	if len(requestName) > 253 { // Kubernetes name limit
		requestName = requestName[:253]
	}

	// Create DiagnosisRequest
	request := &diagnosisv1alpha1.DiagnosisRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      requestName,
			Namespace: policy.Namespace,
		},
		Spec: diagnosisv1alpha1.DiagnosisRequestSpec{
			Type: diagnosisv1alpha1.DiagnosisRequestTypeAutomatic,
			TargetPod: diagnosisv1alpha1.TargetPodReference{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			PolicyRef: diagnosisv1alpha1.PolicyReference{
				Name:      policy.Name,
				Namespace: policy.Namespace,
			},
			TriggerCondition: &diagnosisv1alpha1.TriggerCondition{
				Type:       problemType,
				DetectedAt: metav1.Now(),
			},
		},
	}

	return r.Create(ctx, request)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Named("pod").
		Complete(r)
}
