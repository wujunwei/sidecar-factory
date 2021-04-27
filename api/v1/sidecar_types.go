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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SideCarSpec defines the desired state of SideCar
type SideCarSpec struct {

	// Retry limit is the ceil of the count of retry.
	// +optional
	RetryLimit int `json:"retry_limit,omitempty"`
	// TolerationDuration is the longest duration when sidecar can not find the pod match the selector can wait
	// +optional
	TolerationDuration metav1.Duration `json:"toleration_duration,omitempty"`
	// List of containers belonging to the sidecar.
	// Containers cannot currently be added or removed.
	// There must be at least one container in a sidecar.
	// Cannot be updated.
	// +patchMergeKey=name
	// +patchStrategy=merge
	Containers []corev1.Container `json:"Containers,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// Selector is used to find matching pods.
	// Pods that match this label selector are counted to determine the number of pods
	// in their corresponding topology domain.
	// +optional
	Selector *metav1.LabelSelector `json:"Selector,omitempty" protobuf:"bytes,4,opt,name=labelSelector"`
}

// SideCarStatus defines the observed state of SideCar
type SideCarStatus struct {
	// RetryCount is the count of the sidecars had tried when the relative pod's status was not ready.
	// +optional
	RetryCount int `json:"retry_count"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// A  pointers to Relative Pod.
	// +optional
	RelativePod corev1.ObjectReference `json:"relative_pod,omitempty"`
	// The create time of this sidecar
	// +optional
	CreateTime *metav1.Time `json:"create_time,omitempty"`
}

// +kubebuilder:object:root=true

// SideCar is the Schema for the sidecars API
type SideCar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SideCarSpec   `json:"spec,omitempty"`
	Status SideCarStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SideCarList contains a list of SideCar
type SideCarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SideCar `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SideCar{}, &SideCarList{})
}
