package v1alpha1

var LogTag = "scenario_model"

type Scenario struct {
	Version   string               `json:"version"`
	Id        string               `json:"scenario_id"`
	Title     string               `json:"scenario_title"`
	Type      string               `json:"scenario_type"`
	Enabled   bool                 `json:"enabled"`
	Workloads *map[string]Workload `json:"workloads"`
	Filter    Filter               `json:"filter"`
	GroupBy   []GroupBy            `json:"group_by"`
	RateLimit []RateLimit          `json:"rate_limit"`
}

type Workload struct {
	Executor  ExecutorName `json:"executor"`
	Service   string       `json:"service,omitempty"`
	TraceRole TraceRole    `json:"trace_role,omitempty"`
	Protocol  ProtocolName `json:"protocol,omitempty"`
	Rule      Rule         `json:"rule,omitempty"`
}

type Rule struct {
	Type       string `json:"type"`
	*RuleGroup `json:""`
	*RuleLeaf  `json:""`
}

type RuleGroup struct {
	Condition *Condition `json:"condition,omitempty"`
	Rules     Rules      `json:"rules,omitempty"`
}

type RuleLeaf struct {
	ID       *string        `json:"id,omitempty"`
	Field    *string        `json:"field,omitempty"`
	Datatype *DataType      `json:"datatype,omitempty"`
	Input    *InputTypes    `json:"input,omitempty"`
	Operator *OperatorTypes `json:"operator,omitempty"`
	Value    *ValueTypes    `json:"value,omitempty"`
	JsonPath *[]string      `json:"json_path,omitempty"`
}

type DataType string
type InputTypes string
type OperatorTypes string
type ValueTypes string
type ProtocolName string
type ExecutorName string

const (
	ExecutorEbpf ExecutorName = "EBPF"
	ExecutorOTel ExecutorName = "OTEL"

	ProtocolHTTP       ProtocolName = "HTTP"
	ProtocolGRPC       ProtocolName = "GRPC"
	ProtocolGeneral    ProtocolName = "GENERAL"
	ProtocolIdentifier ProtocolName = "IDENTIFIER"
)

type Rules []Rule

const (
	MYSQL      ProtocolName = "MYSQL"
	HTTP       ProtocolName = "HTTP"
	RULE       string       = "rule"
	RULE_GROUP string       = "rule_group"
)

const (
	server TraceRole = "server"
	client TraceRole = "client"
)

type TraceRole string

const (
	AND Condition = "AND"
	OR  Condition = "OR"
)

type Condition string

type GroupBy struct {
	WorkloadId string `json:"workload_id"`
	Title      string `json:"title"`
	Hash       string `json:"hash"`
}

type RateLimit struct {
	BucketMaxSize    int    `json:"bucket_max_size"`
	BucketRefillSize int    `json:"bucket_refill_size"`
	TickDuration     string `json:"tick_duration"`
}
