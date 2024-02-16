package handler

import (
	"errors"
	"fmt"
	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	promMetrics "github.com/zerok-ai/zk-operator/internal/metrics"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/scenario/model"
	zkredis "github.com/zerok-ai/zk-utils-go/storage/redis"
	dbNames "github.com/zerok-ai/zk-utils-go/storage/redis/clientDBNames"
	"strconv"
	"strings"
	"time"
)

var zkCRDProbeLog = "ZkCrdProbeHandler"

type ZkCRDProbeHandler struct {
	VersionedStore   *zkredis.VersionedStore[model.Scenario]
	latestUpdateTime string
}

func (h *ZkCRDProbeHandler) Init(cfg config.AppConfig) error {
	store, err := zkredis.GetVersionedStore[model.Scenario](&cfg.Redis, dbNames.ScenariosDBName, common.RedisSyncInterval)
	if err != nil {
		return err
	}
	h.VersionedStore = store
	h.latestUpdateTime = "0"

	return nil
}

func (h *ZkCRDProbeHandler) CreateCRDProbe(zerokProbe *operatorv1alpha1.ZerokProbe) (string, error) {
	logger.Debug(zkCRDProbeLog, "New CRD created")
	zkProbe := constructRedisProbeStructureFromCRD(zerokProbe)
	//check if zkProbe is enabled to false delete from redis
	if !zkProbe.Enabled {
		logger.Debug(zkCRDProbeLog, "Probe is Created with enable false, not processing and storing in redis")
	} else {
		err := h.VersionedStore.SetValue(zkProbe.Id, zkProbe)
		if err != nil {
			if errors.Is(err, zkredis.LATEST) {
				logger.Info(zkCRDProbeLog, "Latest value is already present in redis for crd probe Id ", zkProbe.Id)
			} else {
				logger.Error(zkCRDProbeLog, "Error while storing crd probe in redis ", err)
				return "", err
			}
		}
	}
	promMetrics.TotalProbesCreated.Inc()
	logger.Info(zkCRDProbeLog, "Successfully created new Probe with title ", zkProbe.Title)
	return "", nil
}

func (h *ZkCRDProbeHandler) DeleteCRDProbe(zkCRDProbeId string) (string, error) {
	err := h.VersionedStore.Delete(zkCRDProbeId)
	if err != nil {
		logger.Error(zkCRDProbeLog, "Error while deleting crd probe id ", zkCRDProbeId, " from redis ", err)
		return "", err
	}
	promMetrics.TotalProbesDeleted.Inc()
	logger.Info(zkCRDProbeLog, "Successfully Deleted Probe with id ", zkCRDProbeId, " from redis.")
	return "", nil
}

func (h *ZkCRDProbeHandler) UpdateCRDProbe(zerokProbe *operatorv1alpha1.ZerokProbe) (string, error) {
	logger.Debug(zkCRDProbeLog, "CRD updated")
	zkProbe := constructRedisProbeStructureFromCRD(zerokProbe)
	//check if zkProbe is enabled to false delete from redis
	if !zkProbe.Enabled {
		logger.Debug(zkCRDProbeLog, "Probe is disabled, deleting from redis")
		_, err := h.DeleteCRDProbe(zkProbe.Id)
		if err != nil {
			logger.Error(zkCRDProbeLog, "Error while deleting crd probe id ", zkProbe.Id, " from redis ", err)
			return "", err
		}
		logger.Info(zkCRDProbeLog, "Successfully Deleted Probe with id ", zkProbe.Id, " from redis.")
		return "", nil
	}
	err := h.VersionedStore.SetValue(zkProbe.Id, zkProbe)
	if err != nil {
		if errors.Is(err, zkredis.LATEST) {
			logger.Info(zkCRDProbeLog, "Latest value is already present in redis for crd probe Id ", zkProbe.Id)
		} else {
			logger.Error(zkCRDProbeLog, "Error while storing crd probe in redis ", err)
			return "", err
		}
	}
	promMetrics.TotalProbesUpdated.Inc()
	logger.Info(zkCRDProbeLog, "Successfully updated Probe with title ", zkProbe.Title, " from redis.")
	return "", nil
}

func (h *ZkCRDProbeHandler) CleanUpOnKill() error {
	logger.Debug(zkCRDProbeLog, "Kill method in scenario rules.")
	return nil
}

func (h *ZkCRDProbeHandler) IsHealthy() bool {
	return true
}

func constructRedisProbeStructureFromCRD(zerokProbe *operatorv1alpha1.ZerokProbe) model.Scenario {

	zkProbeScenario := model.Scenario{}
	defer func() {
		if r := recover(); r != nil {
			logger.Error(zkCRDProbeLog, "Error in constructing probe from CRD", r)
			zkProbeScenario = model.Scenario{}
		}
	}()

	var zerokProbeWorkloadsMap map[string]model.Workload
	var zerokServiceWorkloadMap map[string]string
	zkProbeScenario.Enabled = zerokProbe.Spec.Enabled
	zkProbeScenario.Version = strconv.FormatInt(time.Now().Unix(), 10)
	zkProbeScenario.Id = string(zerokProbe.GetUID())
	zkProbeScenario.Title = zerokProbe.Spec.Title
	zkProbeScenario.Type = "SYSTEM"
	zerokProbeWorkloadsMap, zerokServiceWorkloadMap = getZerokProbeWorkloadsFromCrd(zerokProbe.Spec.Workloads)
	zkProbeScenario.Workloads = &zerokProbeWorkloadsMap
	zkProbeScenario.RateLimit = getZerokProbeRateLimitFromCrd(zerokProbe.Spec.RateLimit)
	zkProbeScenario.Filter = getZerokProbeFiltersFromCrdFilters(zerokProbe.Spec.Filter, zerokServiceWorkloadMap)
	zkProbeScenario.GroupBy = getZerokProbeGroupByFromCrd(&zerokProbe.Spec.GroupBy, zerokServiceWorkloadMap)
	return zkProbeScenario
}

func getZerokProbeWorkloadsFromCrd(crdWorkloadsMap map[string]operatorv1alpha1.Workload) (map[string]model.Workload, map[string]string) {
	zerokProbeWorkloadsMap := make(map[string]model.Workload)
	zerokServiceWorkloadMap := make(map[string]string)
	for key, value := range crdWorkloadsMap {
		probeZerokWorkload := model.Workload{}
		executor, serviceName, err := getExecutorAndServiceNameFromKey(key)
		if err != nil {
			return nil, nil
		}
		probeZerokWorkload.Service = serviceName
		probeZerokWorkload.Rule = value.Rule
		probeZerokWorkload.TraceRole = "server"
		probeZerokWorkload.Protocol = "HTTP"
		probeZerokWorkload.Executor = model.ExecutorName(executor)
		workloadId := model.WorkLoadUUID(probeZerokWorkload).String()
		zerokProbeWorkloadsMap[workloadId] = probeZerokWorkload
		zerokServiceWorkloadMap[serviceName] = workloadId
	}
	return zerokProbeWorkloadsMap, zerokServiceWorkloadMap
}

func getExecutorAndServiceNameFromKey(workloadKey string) (string, string, error) {
	parts := strings.Split(workloadKey, "/")

	if len(parts) != 2 {
		return "", "", errors.New("invalid input format")
	}

	executor := operatorv1alpha1.ExecutorType(parts[0])
	switch executor {
	case operatorv1alpha1.OTEL:
		// Valid executor
	default:
		return "", "", errors.New(fmt.Sprintf("invalid executor:%s provided in workload key", executor))
	}

	serviceName := parts[1]
	return string(executor), serviceName, nil
}

func getZerokProbeRateLimitFromCrd(crdRateLimitList []operatorv1alpha1.RateLimit) []model.RateLimit {
	if crdRateLimitList == nil {
		return []model.RateLimit{
			{BucketMaxSize: 5, BucketRefillSize: 5, TickDuration: "1m"},
		}
	}
	probeZerokRateLimitList := make([]model.RateLimit, 0)
	for _, crdRateLimit := range crdRateLimitList {
		probeZerokRateLimitList = append(probeZerokRateLimitList, model.RateLimit{TickDuration: crdRateLimit.TickDuration, BucketMaxSize: crdRateLimit.BucketMaxSize, BucketRefillSize: crdRateLimit.BucketRefillSize})
	}
	return probeZerokRateLimitList
}

func getZerokProbeGroupByFromCrd(crdGroupByList *[]operatorv1alpha1.GroupBy, zerokServiceWorkloadMap map[string]string) []model.GroupBy {
	if crdGroupByList == nil {
		return nil
	}
	var probeZerokGroupByList []model.GroupBy
	for _, crdGroupBy := range *crdGroupByList {
		groupBy := &crdGroupBy
		workloadKey := zerokServiceWorkloadMap[groupBy.WorkloadKey]
		probeZerokGroupByList = append(probeZerokGroupByList, model.GroupBy{WorkloadId: workloadKey, Title: groupBy.Title, Hash: groupBy.Hash})
	}
	return probeZerokGroupByList
}

func getZerokProbeFiltersFromCrdFilters(crdFilter operatorv1alpha1.Filter, zerokServiceWorkloadMap map[string]string) model.Filter {
	var workloadIdList model.WorkloadIds
	var probeZerokFilter model.Filter
	if crdFilter.WorkloadKeys == nil || crdFilter.Filters == nil {
		for _, value := range zerokServiceWorkloadMap {
			workloadIdList = append(workloadIdList, value)
		}
		return model.Filter{
			Type:        "workload",
			Condition:   "AND",
			Filters:     nil,
			WorkloadIds: &workloadIdList,
		}
	}
	//iterate over the services in filter and update them with workload id
	// Check if WorkloadIds is not nil before iterating
	if crdFilter.WorkloadKeys != nil {
		// Iterate over WorkloadIds
		for _, serviceId := range *crdFilter.WorkloadKeys {
			workloadIdList = append(workloadIdList, zerokServiceWorkloadMap[serviceId])
		}
		probeZerokFilter.WorkloadIds = &workloadIdList
	}
	if crdFilter.Filters != nil {
		var newFilters model.Filters
		for _, filter := range *crdFilter.Filters {
			newFilters = append(newFilters, getZerokProbeFiltersFromCrdFilters(filter, zerokServiceWorkloadMap))
		}
		probeZerokFilter.Filters = &newFilters
	}
	//TODO:: throw error if both workloadIds and filters are nil
	if crdFilter.Type != "" {
		probeZerokFilter.Type = crdFilter.Type
	} else {
		probeZerokFilter.Type = "workload"
	}

	//TODO:: throw error if condition is NULL or empty
	if probeZerokFilter.Condition != "" {
		probeZerokFilter.Condition = model.Condition(crdFilter.Condition)
	} else {
		probeZerokFilter.Condition = "AND"
	}
	return probeZerokFilter
}
