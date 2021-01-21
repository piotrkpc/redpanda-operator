/*


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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedPandaClusterSpec defines the desired state of RedPandaCluster
type RedPandaClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Image string `json:"image"`
	Name  string `json:"name"`

	// +kubebuilder:validation:Minimum=1
	Size int32 `json:"size"`
}
type ClusterConditionType string

const (
	ClusterReady          ClusterConditionType = "Ready"
	ClusterInitialized    ClusterConditionType = "Initialized"
	ClusterReplacingNodes ClusterConditionType = "ReplacingNodes"
	ClusterScalingUp      ClusterConditionType = "ScalingUp"
	ClusterScalingDown    ClusterConditionType = "ScalingDown"
	ClusterUpdating       ClusterConditionType = "Updating"
	ClusterStopped        ClusterConditionType = "Stopped"
	ClusterResuming       ClusterConditionType = "Resuming"
	ClusterRollingRestart ClusterConditionType = "RollingRestart"
	ClusterValid          ClusterConditionType = "Valid"
)

type ClusterCondition struct {
	Type               ClusterConditionType   `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	Reason             string                 `json:"reason"`
	Message            string                 `json:"message"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
}

type ProgressState string

type RedPandaNodeStatus string

type RedPandaStatusMap map[string]RedPandaNodeStatus

// RedPandaClusterStatus defines the observed state of RedPandaCluster
type RedPandaClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []ClusterCondition `json:"conditions,omitempty"`

	// Last known progress state of the RedPanda Operator
	// +optional
	RedPandaOperatorProgress ProgressState `json:"redPandaOperatorProgress,omitempty"`

	// +optional
	NodeStatuses RedPandaStatusMap `json:"nodeStatuses"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rpkc

// RedPandaCluster is the Schema for the redpandaclusters API
type RedPandaCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedPandaClusterSpec   `json:"spec,omitempty"`
	Status RedPandaClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedPandaClusterList contains a list of RedPandaCluster
type RedPandaClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedPandaCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedPandaCluster{}, &RedPandaClusterList{})
}
