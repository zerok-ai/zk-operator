package model

import (
	"github.com/zerok-ai/zk-utils-go/interfaces"
	"sort"
)

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

func (f Filter) Equals(otherInterface interfaces.ZKComparable) bool {
	other, ok := otherInterface.(Filter)
	if !ok {
		return false
	}

	if f.Type != other.Type || f.Condition != other.Condition {
		return false
	}

	// check nil mismatch
	if (f.Filters == nil && other.Filters != nil) || (f.Filters != nil && other.Filters == nil) || (f.WorkloadIds == nil && other.WorkloadIds != nil) || (f.WorkloadIds != nil && other.WorkloadIds == nil) {
		return false
	}

	// match filters
	if f.Filters != nil && !(*f.Filters).Equals(*other.Filters) {
		return false
	}

	// match workloads
	if f.WorkloadIds != nil && !(*f.WorkloadIds).Equals(*other.WorkloadIds) {
		return false
	}

	return true
}

func (f Filter) LessThan(other Filter) bool {

	if f.Type < other.Type || f.Condition < other.Condition {
		return true
	}

	if f.Type == other.Type {
		if f.Type == WORKLOAD {
			// do null checks
			if f.WorkloadIds == nil && other.WorkloadIds != nil {
				return true
			} else if f.WorkloadIds != nil && other.WorkloadIds == nil {
				return false
			} else if f.WorkloadIds == nil && other.WorkloadIds == nil {
				return false
			}
			return (*f.WorkloadIds).LessThan(*other.WorkloadIds)
		} else if f.Type == FILTER {
			// do null checks
			if f.Filters == nil && other.Filters != nil {
				return true
			} else if f.Filters != nil && other.Filters == nil {
				return false
			} else if f.Filters == nil && other.Filters == nil {
				return false
			}
			return (*f.Filters).LessThan(*other.Filters)
		}
	}

	return false
}

func (f Filter) sort() {
	if f.Type == WORKLOAD {
		sort.Strings(*f.WorkloadIds)
	} else if f.Type == FILTER {
		(*f.Filters).sort()
	}
}

type Filters []Filter

func (f Filters) Len() int           { return len(f) }
func (f Filters) Less(i, j int) bool { return f[i].LessThan(f[j]) }
func (f Filters) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f Filters) Equals(other Filters) bool {
	f.sort()
	other.sort()

	if len(f) != len(other) {
		return false
	}
	for i := 0; i < len(f) && i < len(other); i++ {
		if !f[i].Equals(other[i]) {
			return false
		}
	}

	return true
}

func (f Filters) LessThan(other Filters) bool {

	f.sort()
	other.sort()

	if len(f) != len(other) {
		return len(f) < len(other)
	}

	for i := 0; i < len(f) && i < len(other); i++ {
		if f[i].LessThan(other[i]) {
			return true
		}
	}

	return false
}

func (f Filters) sort() {
	for i := 0; i < len(f); i++ {
		f[i].sort()
	}
	sort.Sort(f)
}

type WorkloadIds []string

func (s WorkloadIds) Len() int           { return len(s) }
func (s WorkloadIds) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s WorkloadIds) Less(i, j int) bool { return s[i] < (s[j]) }
func (s WorkloadIds) Equals(other WorkloadIds) bool {
	if len(s) != len(other) {
		return false
	}

	// sort and check equality
	sort.Strings(s)
	sort.Strings(other)
	for i := 0; i < len(s) && i < len(other); i++ {
		if s[i] != other[i] {
			return false
		}
	}
	return true
}
func (s WorkloadIds) LessThan(other WorkloadIds) bool {

	// sort and check equality
	sort.Strings(s)
	sort.Strings(other)

	for i := 0; i < len(s) && i < len(other); i++ {
		if s[i] < other[i] {
			return true
		}
	}

	return len(s) < len(other)
}
