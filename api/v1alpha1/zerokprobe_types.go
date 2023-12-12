package v1alpha1

import (
	"github.com/zerok-ai/zk-operator/api/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Workloads map[string]model.Workload

type ZerokProbeSpec struct {
	Title     string            `json:"title"`
	Enabled   bool              `json:"enabled"`
	Workloads Workloads         `json:"workloads"`
	Filter    model.Filter      `json:"filter,omitempty"`
	GroupBy   []model.GroupBy   `json:"group_by,omitempty"`
	RateLimit []model.RateLimit `json:"rate_limit,omitempty"`
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

//+kubebuilder:object:root=true

type ZerokProbeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZerokProbe `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ZerokProbe{}, &ZerokProbeList{})
}
