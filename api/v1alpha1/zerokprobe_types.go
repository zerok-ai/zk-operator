package v1alpha1

import (
	"github.com/zerok-ai/zk-utils-go/scenario/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen=true
type Workloads map[string]model.Workload

// +k8s:deepcopy-gen=true
type ZerokCrdSpec struct {
	Title     string            `json:"title"`
	Enabled   bool              `json:"enabled"`
	Workloads Workloads         `json:"workloads,omitempty"`
	Filter    model.Filter      `json:"filter,omitempty"`
	GroupBy   []model.GroupBy   `json:"group_by,omitempty"`
	RateLimit []model.RateLimit `json:"rate_limit,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// above line acts a code marker recognized by Kube builder and root=true mean this is the root object
// ZerokCrd is the CRD schema for crating probe
type ZerokCrd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ZerokCrdSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZerokCrdList contains a list of ZerokCrd
type ZerokCrdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZerokCrd `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ZerokCrd{}, &ZerokCrdList{})
}
