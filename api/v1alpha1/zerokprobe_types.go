package v1alpha1

import (
	"github.com/zerok-ai/zk-utils-go/scenario/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DataType string
type InputTypes string
type OperatorTypes string
type ValueTypes string
type ProtocolName string
type ExecutorName string

type ExecutorTypeEnum struct {
	OTEL ExecutorType
	EBPF ExecutorType
}

const (
	OTEL ExecutorType = "OTEL"
)

type ExecutorType string

// +k8s:deepcopy-gen=true
type Workloads map[string]Workload

// +k8s:deepcopy-gen=true
type ZerokProbeSpec struct {
	Title     string      `json:"title" yaml:"title"`
	Enabled   bool        `json:"enabled" yaml:"enabled"`
	Workloads Workloads   `json:"workloads,omitempty" yaml:"workloads"`
	Filter    Filter      `json:"filter,omitempty" yaml:"filter"`
	GroupBy   []GroupBy   `json:"group_by,omitempty"`
	RateLimit []RateLimit `json:"rate_limit,omitempty"`
}

// +k8s:deepcopy-gen=true
type Workload struct {
	Rule model.Rule `json:"rule,omitempty"`
}

// +k8s:deepcopy-gen=true
type RateLimit struct {
	BucketMaxSize    int    `json:"bucket_max_size"`
	BucketRefillSize int    `json:"bucket_refill_size"`
	TickDuration     string `json:"tick_duration"`
}

// +k8s:deepcopy-gen=true
type Rule struct {
	Type       string `json:"type"`
	*RuleGroup `json:""`
	*RuleLeaf  `json:""`
}

// +k8s:deepcopy-gen=true
type RuleLeaf struct {
	ID       *string        `json:"id,omitempty"`
	Datatype *DataType      `json:"datatype,omitempty"`
	Operator *OperatorTypes `json:"operator,omitempty"`
	Value    *ValueTypes    `json:"value,omitempty"`
	JsonPath *[]string      `json:"json_path,omitempty"`
}

// +k8s:deepcopy-gen=true
type Rules []Rule

// +k8s:deepcopy-gen=true
type RuleGroup struct {
	Condition *Condition `json:"condition,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Rules Rules `json:"rules,omitempty"`
}

// +k8s:deepcopy-gen=true
type GroupBy struct {
	WorkloadKey string `json:"workload_key"`
	Title       string `json:"title"`
	Hash        string `json:"hash"`
}

// +k8s:deepcopy-gen=false
type Filters []Filter

const (
	AND Condition = "AND"
	OR  Condition = "OR"
)

// +k8s:deepcopy-gen=true
type Condition string

// +k8s:deepcopy-gen=true
type WorkloadKeys []string

// +k8s:deepcopy-gen=false
type Filter struct {
	Type      string    `json:"type"`
	Condition Condition `json:"condition"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Filters      *Filters      `json:"filters,omitempty"`
	WorkloadKeys *WorkloadKeys `json:"workload_keys,omitempty"`
}

// ZerokPronePhase is a label for the condition of a Probe at the current time.
// +enum
type ZerokProbePhase string

// These are the valid statuses of pods.
const (
	// ProbePending means the Probe has been accepted by the system,
	// has not been started. This includes time before being bound to a node,
	ProbePending ZerokProbePhase = "Pending"
	// ProbeRunning means the probe has
	// validating the probe and storing in DB
	ProbeRunning ZerokProbePhase = "Running"
	// ProbeSucceeded means that all validations are done stored in DB
	ProbeSucceeded ZerokProbePhase = "Succeeded"
	// ProbeFailed means that all validations are failing or redis storage is failing
	ProbeFailed ZerokProbePhase = "Failed"
	// ProbeUnknown means that for some reason the state of the probe could not be obtained, typically due
	// to an error in communicating with redis
	ProbeUnknown ZerokProbePhase = "Unknown"
	// ProbeDeleting means that the probe is being deleted
	ProbeDeleting ZerokProbePhase = "Deleting"
)

// ZerokProbeStatus defines the observed state of Probe
type ZerokProbeStatus struct {

	// +optional
	Phase ZerokProbePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=ZerokProbePhase"`

	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// ZerokProbe is used to specify rules to filter the spans generated by the services
// ZerokProbe is the CRD schema for crating probe
type ZerokProbe struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ZerokProbeSpec   `json:"spec,omitempty"`
	Status            ZerokProbeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZerokProbeList contains a list of ZerokProbe
type ZerokProbeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZerokProbe `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ZerokProbe{}, &ZerokProbeList{})
}

//deep copy methods

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Filter.
func (in *Filter) DeepCopy() *Filter {
	if in == nil {
		return nil
	}
	out := new(Filter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Filter) DeepCopyInto(out *Filter) {
	*out = *in
	if in.Filters != nil {
		in, out := &in.Filters, &out.Filters
		*out = (*Filters)(new([]Filter))
		if **in != nil {
			in, out := *in, *out
			*out = make([]Filter, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
	if in.WorkloadKeys != nil {
		in, out := &in.WorkloadKeys, &out.WorkloadKeys
		*out = (*WorkloadKeys)(new([]string))
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Filters) DeepCopyInto(out *Filters) {
	{
		in := &in
		*out = make(Filters, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Filters.
func (in Filters) DeepCopy() Filters {
	if in == nil {
		return nil
	}
	out := new(Filters)
	in.DeepCopyInto(out)
	return *out
}
