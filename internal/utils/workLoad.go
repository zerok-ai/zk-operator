package utils

const (
	DEPLYOMENT = iota
	STATEFULSET
	DAEMONSET
	OTHERS
)

type WorkLoad struct {
	WorkLoadType int
	Name         string
}

func (w *WorkLoad) Equals(other *WorkLoad) bool {
	if other == nil {
		return false
	}
	return w.Name == other.Name && w.WorkLoadType == other.WorkLoadType
}

func (w *WorkLoad) NotEquals(other *WorkLoad) bool {
	return !w.Equals(other)
}
