package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ZerokopSpec defines the desired state of Zerokop
type ZerokopSpec struct {
	Images []ImageOverride `json:"images"`
}

// ImageOverride defines overrides for a specific image.
type ImageOverride struct {
	ImageID     string   `json:"imageID"`
	Env         []EnvVar `json:"env,omitempty"`
	CmdOverride []string `json:"cmd_override,omitempty"`
}

// EnvVar defines environment variables to override.
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ZerokopStatus defines the observed state of Zerokop
type ZerokopStatus struct {
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true

// Zerokop is the Schema for the zerokops API
// +kubebuilder:subresource:status
type Zerokop struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZerokopSpec   `json:"spec,omitempty"`
	Status ZerokopStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ZerokopList contains a list of Zerokop
type ZerokopList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zerokop `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Zerokop{}, &ZerokopList{})
}
