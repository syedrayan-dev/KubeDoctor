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
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
	"github.com/yth01/apollo/internal/llm"
)

// DiagnosisRequestReconciler reconciles a DiagnosisRequest object
type DiagnosisRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *rest.Config
}

// +kubebuilder:rbac:groups=diagnosis.apollo.dev,resources=diagnosisrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=diagnosis.apollo.dev,resources=diagnosisrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=diagnosis.apollo.dev,resources=diagnosisrequests/finalizers,verbs=update
// +kubebuilder:rbac:groups=diagnosis.apollo.dev,resources=diagnosispolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups=diagnosis.apollo.dev,resources=diagnosisreports,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/log,verbs=get
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DiagnosisRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Step 1: Fetch the DiagnosisRequest
	var diagnosisRequest diagnosisv1alpha1.DiagnosisRequest
	if err := r.Get(ctx, req.NamespacedName, &diagnosisRequest); err != nil {
		if errors.IsNotFound(err) {
			log.V(1).Info("DiagnosisRequest not found, might have been deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get DiagnosisRequest")
		return ctrl.Result{}, err
	}

	log.Info("Processing DiagnosisRequest",
		"name", diagnosisRequest.Name,
		"namespace", diagnosisRequest.Namespace,
		"phase", diagnosisRequest.Status.Phase,
		"targetPod", diagnosisRequest.Spec.TargetPod.Name)

	// Step 2: Handle different phases
	switch diagnosisRequest.Status.Phase {
	case "": // New request (empty phase)
		return r.handleNewRequest(ctx, &diagnosisRequest)
	case diagnosisv1alpha1.DiagnosisRequestPhasePending:
		return r.handlePendingRequest(ctx, &diagnosisRequest)
	case diagnosisv1alpha1.DiagnosisRequestPhaseInProgress:
		return r.handleInProgressRequest(ctx, &diagnosisRequest)
	case diagnosisv1alpha1.DiagnosisRequestPhaseCompleted, diagnosisv1alpha1.DiagnosisRequestPhaseFailed:
		// Nothing to do for completed/failed requests
		return ctrl.Result{}, nil
	default:
		log.Info("Unknown phase, setting to pending", "phase", diagnosisRequest.Status.Phase)
		return r.updateRequestStatus(ctx, &diagnosisRequest, diagnosisv1alpha1.DiagnosisRequestPhasePending, "Unknown phase, resetting to pending")
	}
}

// handleNewRequest handles newly created DiagnosisRequest
func (r *DiagnosisRequestReconciler) handleNewRequest(ctx context.Context, diagnosisRequest *diagnosisv1alpha1.DiagnosisRequest) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Handling new DiagnosisRequest", "name", diagnosisRequest.Name)

	// Set initial status
	return r.updateRequestStatus(ctx, diagnosisRequest, diagnosisv1alpha1.DiagnosisRequestPhasePending, "DiagnosisRequest created")
}

// handlePendingRequest handles pending DiagnosisRequest by starting the diagnosis process
func (r *DiagnosisRequestReconciler) handlePendingRequest(ctx context.Context, diagnosisRequest *diagnosisv1alpha1.DiagnosisRequest) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting diagnosis for pending request", "name", diagnosisRequest.Name)

	// Update status to InProgress
	if result, err := r.updateRequestStatus(ctx, diagnosisRequest, diagnosisv1alpha1.DiagnosisRequestPhaseInProgress, "Starting diagnosis"); err != nil {
		return result, err
	}

	// Requeue immediately to process the InProgress state
	return ctrl.Result{Requeue: true}, nil
}

// handleInProgressRequest handles DiagnosisRequest in progress by performing the actual diagnosis
func (r *DiagnosisRequestReconciler) handleInProgressRequest(ctx context.Context, diagnosisRequest *diagnosisv1alpha1.DiagnosisRequest) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Processing in-progress DiagnosisRequest", "name", diagnosisRequest.Name)

	// Step 1: Get the DiagnosisPolicy
	policy, err := r.getDiagnosisPolicy(ctx, diagnosisRequest)
	if err != nil {
		log.Error(err, "Failed to get DiagnosisPolicy")
		_, _ = r.updateRequestStatus(ctx, diagnosisRequest, diagnosisv1alpha1.DiagnosisRequestPhaseFailed, fmt.Sprintf("Failed to get policy: %v", err))
		return ctrl.Result{}, err
	}

	// Step 2: Get target pod
	pod, err := r.getTargetPod(ctx, diagnosisRequest)
	if err != nil {
		log.Error(err, "Failed to get target pod")
		_, _ = r.updateRequestStatus(ctx, diagnosisRequest, diagnosisv1alpha1.DiagnosisRequestPhaseFailed, fmt.Sprintf("Failed to get target pod: %v", err))
		return ctrl.Result{}, err
	}

	// Step 3: Collect pod data
	collectedData, err := r.collectPodData(ctx, pod, policy)
	if err != nil {
		log.Error(err, "Failed to collect pod data")
		_, _ = r.updateRequestStatus(ctx, diagnosisRequest, diagnosisv1alpha1.DiagnosisRequestPhaseFailed, fmt.Sprintf("Failed to collect data: %v", err))
		return ctrl.Result{}, err
	}

	// Step 4: Perform LLM analysis
	analysis, err := r.performLLMAnalysis(ctx, policy, collectedData, diagnosisRequest.Spec.TriggerCondition)
	if err != nil {
		log.Error(err, "Failed to perform LLM analysis")
		_, _ = r.updateRequestStatus(ctx, diagnosisRequest, diagnosisv1alpha1.DiagnosisRequestPhaseFailed, fmt.Sprintf("Failed to perform LLM analysis: %v", err))
		return ctrl.Result{}, err
	}

	// Step 5: Create DiagnosisReport
	if err := r.createDiagnosisReport(ctx, diagnosisRequest, policy, pod, collectedData, analysis); err != nil {
		log.Error(err, "Failed to create DiagnosisReport")
		_, _ = r.updateRequestStatus(ctx, diagnosisRequest, diagnosisv1alpha1.DiagnosisRequestPhaseFailed, fmt.Sprintf("Failed to create report: %v", err))
		return ctrl.Result{}, err
	}

	// Step 6: Update status to completed
	return r.updateRequestStatus(ctx, diagnosisRequest, diagnosisv1alpha1.DiagnosisRequestPhaseCompleted, "Diagnosis completed successfully")
}

// updateRequestStatus updates the DiagnosisRequest status
func (r *DiagnosisRequestReconciler) updateRequestStatus(ctx context.Context, diagnosisRequest *diagnosisv1alpha1.DiagnosisRequest, phase diagnosisv1alpha1.DiagnosisRequestPhase, message string) (ctrl.Result, error) {
	diagnosisRequest.Status.Phase = phase
	diagnosisRequest.Status.Message = message
	diagnosisRequest.Status.LastUpdateTime = &metav1.Time{Time: time.Now()}

	if phase == diagnosisv1alpha1.DiagnosisRequestPhaseCompleted || phase == diagnosisv1alpha1.DiagnosisRequestPhaseFailed {
		diagnosisRequest.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	}

	if err := r.Status().Update(ctx, diagnosisRequest); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
	}

	return ctrl.Result{}, nil
}

// getDiagnosisPolicy retrieves the DiagnosisPolicy referenced by the request
func (r *DiagnosisRequestReconciler) getDiagnosisPolicy(ctx context.Context, diagnosisRequest *diagnosisv1alpha1.DiagnosisRequest) (*diagnosisv1alpha1.DiagnosisPolicy, error) {
	var policy diagnosisv1alpha1.DiagnosisPolicy
	policyKey := types.NamespacedName{
		Name:      diagnosisRequest.Spec.PolicyRef.Name,
		Namespace: diagnosisRequest.Spec.PolicyRef.Namespace,
	}

	if err := r.Get(ctx, policyKey, &policy); err != nil {
		return nil, fmt.Errorf("failed to get DiagnosisPolicy %s/%s: %w",
			diagnosisRequest.Spec.PolicyRef.Namespace,
			diagnosisRequest.Spec.PolicyRef.Name, err)
	}

	return &policy, nil
}

// getTargetPod retrieves the target pod for diagnosis
func (r *DiagnosisRequestReconciler) getTargetPod(ctx context.Context, diagnosisRequest *diagnosisv1alpha1.DiagnosisRequest) (*corev1.Pod, error) {
	var pod corev1.Pod
	podKey := types.NamespacedName{
		Name:      diagnosisRequest.Spec.TargetPod.Name,
		Namespace: diagnosisRequest.Spec.TargetPod.Namespace,
	}

	if err := r.Get(ctx, podKey, &pod); err != nil {
		return nil, fmt.Errorf("failed to get target pod %s/%s: %w",
			diagnosisRequest.Spec.TargetPod.Namespace,
			diagnosisRequest.Spec.TargetPod.Name, err)
	}

	return &pod, nil
}

// collectPodData collects comprehensive pod data for LLM analysis
func (r *DiagnosisRequestReconciler) collectPodData(ctx context.Context, pod *corev1.Pod, policy *diagnosisv1alpha1.DiagnosisPolicy) (diagnosisv1alpha1.CollectedData, error) {
	// Generate full pod description including all relevant information
	podDescription, err := r.generateFullPodDescription(ctx, pod)
	if err != nil {
		return diagnosisv1alpha1.CollectedData{}, fmt.Errorf("failed to generate pod description: %w", err)
	}

	// Collect pod logs
	logs, logLines, err := r.collectPodLogs(ctx, pod, policy)
	if err != nil {
		// Log collection is not critical, continue with empty logs
		logf.FromContext(ctx).Error(err, "Failed to collect pod logs, continuing without logs")
		logs = fmt.Sprintf("Failed to collect logs: %v", err)
		logLines = 0
	}

	// Build collected data
	collectedData := diagnosisv1alpha1.CollectedData{
		PodDescription: podDescription,
		Logs:           logs,
		LogLines:       logLines,
		CollectionTime: metav1.Time{Time: time.Now()},
	}

	return collectedData, nil
}

// collectPodLogs collects logs from all containers in the pod
func (r *DiagnosisRequestReconciler) collectPodLogs(ctx context.Context, pod *corev1.Pod, policy *diagnosisv1alpha1.DiagnosisPolicy) (string, int, error) {
	if r.Config == nil {
		return "", 0, fmt.Errorf("kubernetes config not available")
	}

	// Create kubernetes clientset
	clientset, err := kubernetes.NewForConfig(r.Config)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Determine max log lines from policy settings, use Kubernetes default if not specified
	var maxLogLines *int64
	if policy.Spec.DiagnosisSettings != nil && policy.Spec.DiagnosisSettings.MaxLogLines != nil {
		lines := int64(*policy.Spec.DiagnosisSettings.MaxLogLines)
		maxLogLines = &lines
	}
	// If not specified in policy, let Kubernetes use its default (no TailLines specified)

	var allLogs []string
	totalLines := 0

	// Collect logs from all containers
	for _, container := range pod.Spec.Containers {
		containerLogs, containerLines, err := r.getContainerLogs(ctx, clientset, pod, container.Name, maxLogLines)
		if err != nil {
			allLogs = append(allLogs, fmt.Sprintf("Failed to collect logs from %s: %v\n", container.Name, err))
		} else {
			allLogs = append(allLogs, containerLogs)
			totalLines += containerLines
		}
	}

	// Also collect logs from init containers if they exist
	for _, container := range pod.Spec.InitContainers {
		containerLogs, containerLines, err := r.getContainerLogs(ctx, clientset, pod, container.Name, maxLogLines)
		if err != nil {
			allLogs = append(allLogs, fmt.Sprintf("Failed to collect logs from init container %s: %v\n", container.Name, err))
		} else {
			allLogs = append(allLogs, containerLogs)
			totalLines += containerLines
		}
	}

	return strings.Join(allLogs, "\n"), totalLines, nil
}

// getContainerLogs gets logs for a specific container
func (r *DiagnosisRequestReconciler) getContainerLogs(ctx context.Context, clientset *kubernetes.Clientset, pod *corev1.Pod, containerName string, maxLines *int64) (string, int, error) {
	// Pod log options
	podLogOpts := &corev1.PodLogOptions{
		Container: containerName,
	}

	// Only set TailLines if specified in policy
	if maxLines != nil {
		podLogOpts.TailLines = maxLines
	}

	// Get logs request
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get logs stream: %w", err)
	}
	defer func() { _ = podLogs.Close() }()

	// Read all logs
	logs, err := io.ReadAll(podLogs)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read logs: %w", err)
	}

	logString := string(logs)
	lineCount := strings.Count(logString, "\n")

	return logString, lineCount, nil
}

// generateFullPodDescription creates a comprehensive pod description similar to kubectl describe
func (r *DiagnosisRequestReconciler) generateFullPodDescription(ctx context.Context, pod *corev1.Pod) (string, error) {
	var sections []string

	// Basic Pod Information
	podInfo := map[string]interface{}{
		"Name":        pod.Name,
		"Namespace":   pod.Namespace,
		"Phase":       string(pod.Status.Phase),
		"Node":        pod.Spec.NodeName,
		"StartTime":   pod.Status.StartTime,
		"Labels":      pod.Labels,
		"Annotations": pod.Annotations,
	}

	podInfoJSON, err := json.MarshalIndent(podInfo, "", "  ")
	if err != nil {
		return "", err
	}
	sections = append(sections, "=== POD INFORMATION ===\n"+string(podInfoJSON))

	// Container Information
	containerInfo := map[string]interface{}{
		"Containers":        pod.Spec.Containers,
		"ContainerStatuses": pod.Status.ContainerStatuses,
	}

	containerInfoJSON, err := json.MarshalIndent(containerInfo, "", "  ")
	if err != nil {
		return "", err
	}
	sections = append(sections, "=== CONTAINER INFORMATION ===\n"+string(containerInfoJSON))

	// Pod Conditions
	if len(pod.Status.Conditions) > 0 {
		conditionsJSON, err := json.MarshalIndent(pod.Status.Conditions, "", "  ")
		if err != nil {
			return "", err
		}
		sections = append(sections, "=== POD CONDITIONS ===\n"+string(conditionsJSON))
	}

	// Events related to this pod
	events, err := r.getPodEvents(ctx, pod)
	if err == nil && len(events) > 0 {
		eventsJSON, err := json.MarshalIndent(events, "", "  ")
		if err == nil {
			sections = append(sections, "=== RELATED EVENTS ===\n"+string(eventsJSON))
		}
	}

	// Volume Information
	if len(pod.Spec.Volumes) > 0 {
		volumesJSON, err := json.MarshalIndent(pod.Spec.Volumes, "", "  ")
		if err == nil {
			sections = append(sections, "=== VOLUMES ===\n"+string(volumesJSON))
		}
	}

	return strings.Join(sections, "\n\n"), nil
}

// getPodEvents retrieves events related to the pod
func (r *DiagnosisRequestReconciler) getPodEvents(ctx context.Context, pod *corev1.Pod) ([]map[string]interface{}, error) {
	var eventList corev1.EventList
	if err := r.List(ctx, &eventList, client.InNamespace(pod.Namespace)); err != nil {
		return nil, err
	}

	var events []map[string]interface{}
	for _, event := range eventList.Items {
		// Filter events related to this pod
		if event.InvolvedObject.Name == pod.Name && event.InvolvedObject.Kind == "Pod" {
			eventInfo := map[string]interface{}{
				"Type":      event.Type,
				"Reason":    event.Reason,
				"Message":   event.Message,
				"FirstTime": event.FirstTimestamp,
				"LastTime":  event.LastTimestamp,
				"Count":     event.Count,
			}
			events = append(events, eventInfo)
		}
	}

	return events, nil
}

// performLLMAnalysis performs LLM analysis using configured provider
func (r *DiagnosisRequestReconciler) performLLMAnalysis(ctx context.Context, policy *diagnosisv1alpha1.DiagnosisPolicy, collectedData diagnosisv1alpha1.CollectedData, triggerCondition *diagnosisv1alpha1.TriggerCondition) (diagnosisv1alpha1.DiagnosisAnalysis, error) {
	// Read API key from Secret if specified (required for providers like OpenAI, optional for Ollama)
	var apiKey string
	if policy.Spec.LLMConfig.APIKeySecretRef != nil && policy.Spec.LLMConfig.APIKeySecretRef.Name != "" {
		var err error
		apiKey, err = r.getAPIKeyFromSecret(ctx, *policy.Spec.LLMConfig.APIKeySecretRef)
		if err != nil {
			return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("failed to get API key from secret: %w", err)
		}
	}

	// Create analyzer based on policy LLM configuration
	analyzer, err := llm.NewAnalyzer(policy.Spec.LLMConfig, apiKey)
	if err != nil {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("failed to create LLM analyzer: %w", err)
	}

	// Perform analysis using the analyzer
	analysis, err := analyzer.AnalyzePodDiagnosis(ctx, collectedData, triggerCondition, policy)
	if err != nil {
		return diagnosisv1alpha1.DiagnosisAnalysis{}, fmt.Errorf("LLM analysis failed: %w", err)
	}

	return analysis, nil
}

// createDiagnosisReport creates a DiagnosisReport with the analysis results
func (r *DiagnosisRequestReconciler) createDiagnosisReport(ctx context.Context,
	diagnosisRequest *diagnosisv1alpha1.DiagnosisRequest,
	policy *diagnosisv1alpha1.DiagnosisPolicy,
	pod *corev1.Pod,
	collectedData diagnosisv1alpha1.CollectedData,
	analysis diagnosisv1alpha1.DiagnosisAnalysis) error {

	// Generate report name
	reportName := fmt.Sprintf("%s-report-%d", diagnosisRequest.Name, time.Now().Unix())
	if len(reportName) > 253 {
		reportName = reportName[:253]
	}

	// Get main container image
	mainImage := ""
	if len(pod.Spec.Containers) > 0 {
		mainImage = pod.Spec.Containers[0].Image
	}

	report := &diagnosisv1alpha1.DiagnosisReport{
		ObjectMeta: metav1.ObjectMeta{
			Name:      reportName,
			Namespace: diagnosisRequest.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: diagnosisRequest.APIVersion,
					Kind:       diagnosisRequest.Kind,
					Name:       diagnosisRequest.Name,
					UID:        diagnosisRequest.UID,
				},
			},
		},
		Spec: diagnosisv1alpha1.DiagnosisReportSpec{
			TargetPod: diagnosisv1alpha1.PodReference{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				UID:       string(pod.UID),
				Image:     mainImage,
			},
			PolicyRef: diagnosisv1alpha1.PolicyReference{
				Name:      policy.Name,
				Namespace: policy.Namespace,
			},
			TriggerCondition: *diagnosisRequest.Spec.TriggerCondition,
			CollectedData:    collectedData,
			Analysis:         analysis,
		},
	}

	if err := r.Create(ctx, report); err != nil {
		return fmt.Errorf("failed to create DiagnosisReport: %w", err)
	}

	return nil
}

// getAPIKeyFromSecret retrieves the API key from the specified Secret
func (r *DiagnosisRequestReconciler) getAPIKeyFromSecret(ctx context.Context, secretRef diagnosisv1alpha1.SecretReference) (string, error) {
	// Determine secret namespace
	namespace := secretRef.Namespace
	if namespace == "" {
		// If not specified, use the same namespace as the DiagnosisPolicy
		// Note: This requires passing the policy namespace, but for now we'll require explicit namespace
		return "", fmt.Errorf("secret namespace must be specified")
	}

	// Get the secret
	var secret corev1.Secret
	secretKey := types.NamespacedName{
		Name:      secretRef.Name,
		Namespace: namespace,
	}

	if err := r.Get(ctx, secretKey, &secret); err != nil {
		return "", fmt.Errorf("failed to get secret %s/%s: %w", namespace, secretRef.Name, err)
	}

	// Extract the API key
	apiKeyBytes, exists := secret.Data[secretRef.Key]
	if !exists {
		return "", fmt.Errorf("key '%s' not found in secret %s/%s", secretRef.Key, namespace, secretRef.Name)
	}

	return string(apiKeyBytes), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DiagnosisRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&diagnosisv1alpha1.DiagnosisRequest{}).
		Named("diagnosisrequest").
		Complete(r)
}
