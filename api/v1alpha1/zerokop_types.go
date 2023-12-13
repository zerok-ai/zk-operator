package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ZerokinstrumentationSpec defines the desired state of Zerokinstrumentation
type ZerokinstrumentationSpec struct {
}

// ZerokinstrumentationStatus defines the observed state of ZerokInstrumentation
type ZerokinstrumentationStatus struct {
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true

// Zerokinstrumentation is the Schema for the Zerokinstrumentations API
// +kubebuilder:subresource:status
type Zerokinstrumentation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZerokinstrumentationSpec   `json:"spec,omitempty"`
	Status ZerokinstrumentationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ZerokinstrumentationList contains a list of ZerokInstrumentation
type ZerokinstrumentationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zerokinstrumentation `json:"items"`
}

//func init() {
//	SchemeBuilder.Register(&Zerokinstrumentation{}, &ZerokinstrumentationList{})
//}
