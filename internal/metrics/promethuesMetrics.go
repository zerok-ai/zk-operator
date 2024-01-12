package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	//total number of CRD's created
	TotalProbesCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zerok_crd_created_total",
		Help: "total number of CRD's created.",
	})

	//total number of CRD's updated
	TotalProbesUpdated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zerok_crd_updated_total",
		Help: "total number of CRD's updated.",
	})

	//total number of CRD's deleted
	TotalProbesDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zerok_crd_deleted_total",
		Help: "total number of CRD's deleted.",
	})
)
