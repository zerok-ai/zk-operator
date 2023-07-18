package utils

const (
	DEPLYOMENT = iota
	STATEFULSET
	DAEMONSET
	OTHERS
)

type WorkLoad struct {
	workLoadType int
	name         string
}

func (w *WorkLoad) Equals(other *WorkLoad) bool {
	if other == nil {
		return false
	}
	return w.name == other.name && w.workLoadType == other.workLoadType
}

func (w *WorkLoad) NotEquals(other *WorkLoad) bool {
	return !w.Equals(other)
}
