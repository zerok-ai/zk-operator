package handler

import (
	"errors"
	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/scenario/model"
	zkredis "github.com/zerok-ai/zk-utils-go/storage/redis"
	dbNames "github.com/zerok-ai/zk-utils-go/storage/redis/clientDBNames"
	"strconv"
	"time"
)

var zkCRDProbeLog = "ZkCrdProbeHandler"

type ZkCRDProbeHandler struct {
	VersionedStore   *zkredis.VersionedStore[model.Scenario]
	latestUpdateTime string
}

func (h *ZkCRDProbeHandler) Init(cfg config.ZkOperatorConfig) error {
	store, err := zkredis.GetVersionedStore[model.Scenario](&cfg.Redis, dbNames.ScenariosDBName, common.RedisSyncInterval)
	if err != nil {
		return err
	}
	h.VersionedStore = store
	h.latestUpdateTime = "0"

	return nil
}

func (h *ZkCRDProbeHandler) CreateCRDProbe(zerokProbe *operatorv1alpha1.ZerokCrd) (string, error) {
	logger.Debug(zkCRDProbeLog, "New CRD created")
	zkProbe := constructRedisProbeStructureFromCRD(zerokProbe)
	err := h.VersionedStore.SetValue(zkProbe.Id, zkProbe)
	if err != nil {
		if errors.Is(err, zkredis.LATEST) {
			logger.Info(zkCRDProbeLog, "Latest value is already present in redis for crd probe Id ", zkProbe.Id)
		} else {
			logger.Error(zkCRDProbeLog, "Error while storing crd probe in redis ", err)
			return "", err
		}
	}
	logger.Info(zkCRDProbeLog, "Successfully created new Probe with title ", zkProbe.Title, " from redis.")
	return "", nil
}

func (h *ZkCRDProbeHandler) DeleteCRDProbe(zkCRDProbeId string) (string, error) {
	err := h.VersionedStore.Delete(zkCRDProbeId)
	if err != nil {
		logger.Error(zkCRDProbeLog, "Error while deleting crd probe id ", zkCRDProbeId, " from redis ", err)
		return "", err
	}
	logger.Info(zkCRDProbeLog, "Successfully Deleted Probe with id ", zkCRDProbeId, " from redis.")
	return "", nil
}

func (h *ZkCRDProbeHandler) UpdateCRDProbe(zerokProbe *operatorv1alpha1.ZerokCrd) (string, error) {
	logger.Debug(zkCRDProbeLog, "CRD updated")
	zkProbe := constructRedisProbeStructureFromCRD(zerokProbe)
	err := h.VersionedStore.SetValue(zkProbe.Id, zkProbe)
	if err != nil {
		if errors.Is(err, zkredis.LATEST) {
			logger.Info(zkCRDProbeLog, "Latest value is already present in redis for crd probe Id ", zkProbe.Id)
		} else {
			logger.Error(zkCRDProbeLog, "Error while storing crd probe in redis ", err)
			return "", err
		}
	}
	logger.Info(zkCRDProbeLog, "Successfully updated Probe with title ", zkProbe.Title, " from redis.")
	return "", nil
}

func (h *ZkCRDProbeHandler) CleanUpOnKill() error {
	logger.Debug(scenarioLogTag, "Kill method in scenario rules.")
	return nil
}

func (h *ZkCRDProbeHandler) IsHealthy() bool {
	return true
}

func constructRedisProbeStructureFromCRD(zerokProbe *operatorv1alpha1.ZerokCrd) model.Scenario {
	var zerokProbeWorkloadsMap map[string]model.Workload
	var zerokServiceWorkloadMap map[string]string

	zkProbeScenario := model.Scenario{}

	zkProbeScenario.Enabled = zerokProbe.Spec.Enabled
	zkProbeScenario.Version = strconv.FormatInt(time.Now().Unix(), 10)
	zkProbeScenario.Id = string(zerokProbe.GetUID())
	zkProbeScenario.Title = zerokProbe.Spec.Title
	zkProbeScenario.Type = "SYSTEM"
	zerokProbeWorkloadsMap, zerokServiceWorkloadMap = getZerokProbeWorkloadsFromCrd(zerokProbe.Spec.Workloads)
	zkProbeScenario.Workloads = &zerokProbeWorkloadsMap
	zkProbeScenario.RateLimit = getZerokProbeRateLimitFromCrd(zerokProbe.Spec.RateLimit)
	zkProbeScenario.Filter = getZerokProbeFiltersFromCrdFilters(zerokProbe.Spec.Filter, zerokServiceWorkloadMap)
	zkProbeScenario.GroupBy = zerokProbe.Spec.GroupBy
	return zkProbeScenario
}

func getZerokProbeWorkloadsFromCrd(crdWorkloadsMap map[string]model.Workload) (map[string]model.Workload, map[string]string) {
	zerokProbeWorkloadsMap := make(map[string]model.Workload)
	zerokServiceWorkloadMap := make(map[string]string)
	for key, value := range crdWorkloadsMap {
		workloadId := model.WorkLoadUUID(value).String()
		value.Service = key
		zerokProbeWorkloadsMap[workloadId] = value
		zerokServiceWorkloadMap[value.Service] = workloadId
	}
	return zerokProbeWorkloadsMap, zerokServiceWorkloadMap
}

func getZerokProbeRateLimitFromCrd(crdRateLimitList []model.RateLimit) []model.RateLimit {
	if crdRateLimitList == nil {
		return []model.RateLimit{
			{BucketMaxSize: 5, BucketRefillSize: 5, TickDuration: "1m"},
		}
	}
	return crdRateLimitList
}

func getZerokProbeFiltersFromCrdFilters(crdFilter model.Filter, zerokServiceWorkloadMap map[string]string) model.Filter {
	var workloadIdList model.WorkloadIds
	if &crdFilter == nil || crdFilter.Type == "" {
		for _, value := range zerokServiceWorkloadMap {
			workloadIdList = append(workloadIdList, value)
		}
		return model.Filter{
			Type:        "defaultType",
			Condition:   "AND",
			Filters:     nil,
			WorkloadIds: &workloadIdList,
		}
	}
	//iterate over the services in filter and update them with workload id
	// Check if WorkloadIds is not nil before iterating
	if crdFilter.WorkloadIds != nil {
		// Iterate over WorkloadIds
		for _, serviceId := range *crdFilter.WorkloadIds {
			workloadIdList = append(workloadIdList, zerokServiceWorkloadMap[serviceId])
		}
		crdFilter.WorkloadIds = &workloadIdList
	}
	if crdFilter.Filters != nil {
		var newFilters model.Filters
		for _, filter := range *crdFilter.Filters {
			newFilters = append(newFilters, getZerokProbeFiltersFromCrdFilters(filter, zerokServiceWorkloadMap))
		}
		crdFilter.Filters = &newFilters
	}
	return crdFilter
}
