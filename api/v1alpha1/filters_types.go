package v1alpha1

const (
	FILTER   = "filter"
	WORKLOAD = "workload"

	CONDITION_AND = "AND"
	CONDITION_OR  = "OR"
)

type Filter struct {
	Type        string       `json:"type"`
	Condition   Condition    `json:"condition"`
	Filters     *Filters     `json:"filters,omitempty"`
	WorkloadIds *WorkloadIds `json:"workload_ids,omitempty"`
}

type Filters []Filter

type WorkloadIds []string
