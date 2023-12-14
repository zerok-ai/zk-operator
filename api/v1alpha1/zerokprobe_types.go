package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+k8s:deepcopy-gen=true

type Workloads map[string]Workload

type ZerokProbeSpec struct {
	Title     string      `json:"title"`
	Enabled   bool        `json:"enabled"`
	Workloads Workloads   `json:"workloads"`
	Filter    Filter      `json:"filter,omitempty"`
	GroupBy   []GroupBy   `json:"group_by,omitempty"`
	RateLimit []RateLimit `json:"rate_limit,omitempty"`
}

type ZerokProbeStatus struct {
	IsCreated bool `json:"is_created,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// above line acts a code marker recognized by Kube builder and root=true mean this is the root object
// ZerokProbe is the CRD schema for crating probe
type ZerokProbe struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZerokProbeSpec   `json:"spec,omitempty"`
	Status ZerokProbeStatus `json:"status,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ZerokProbe{})
}
