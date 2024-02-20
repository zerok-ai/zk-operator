package response

type CRDListResponse struct {
	CRDList []ProbeResponse `json:"crdList"`
}

type ProbeResponse struct {
	Title     string    `json:"title" yaml:"title"`
	Enabled   bool      `json:"enabled" yaml:"enabled"`
	Workloads Workloads `json:"workloads,omitempty" yaml:"workloads"`
	Filter    Filter    `json:"filter,omitempty" yaml:"filter"`
}

type Workloads map[string]Workload

type Workload struct {
	Rule Rule `json:"rule,omitempty" yaml:"rule"`
}

type Filter struct {
	Type        string       `json:"type" yaml:"type"`
	Condition   Condition    `json:"condition" yaml:"condition"`
	Filters     *Filters     `json:"filters,omitempty" yaml:"filters"`
	WorkloadIds *WorkloadIds `json:"workload_keys,omitempty" yaml:"workload_keys"`
}

type Filters []Filter

type WorkloadIds []string

type Rule struct {
	Type       string `json:"type"`
	*RuleGroup `yaml:",inline,omitempty"`
	*RuleLeaf  `yaml:",inline,omitempty"`
}

type RuleGroup struct {
	Condition *Condition `json:"condition,omitempty" yaml:"condition"`
	Rules     Rules      `json:"rules,omitempty" yaml:"rules"`
}

type Rules []Rule

type RuleLeaf struct {
	ID       *string        `json:"id,omitempty" yaml:"id,omitempty"`
	Field    *string        `json:"field,omitempty" yaml:"field,omitempty"`
	Datatype *DataType      `json:"datatype,omitempty" yaml:"datatype,omitempty"`
	Input    *InputTypes    `json:"input,omitempty" yaml:"input"`
	Operator *OperatorTypes `json:"operator,omitempty" yaml:"operator"`
	Value    *ValueTypes    `json:"value,omitempty" yaml:"value"`
	JsonPath *[]string      `json:"json_path,omitempty" yaml:"json_path"`
}

type Condition string
type DataType string
type InputTypes string
type OperatorTypes string
type ValueTypes string
