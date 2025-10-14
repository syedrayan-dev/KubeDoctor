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

package llm

import (
	"context"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
)

// DiagnosisAnalyzer defines the interface for LLM-based diagnosis analysis
type DiagnosisAnalyzer interface {
	// AnalyzePodDiagnosis performs diagnosis analysis on collected pod data
	AnalyzePodDiagnosis(ctx context.Context, collectedData diagnosisv1alpha1.CollectedData, triggerCondition *diagnosisv1alpha1.TriggerCondition, policy *diagnosisv1alpha1.DiagnosisPolicy) (diagnosisv1alpha1.DiagnosisAnalysis, error)
}
