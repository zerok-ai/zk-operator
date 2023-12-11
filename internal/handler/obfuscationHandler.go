package handler

import (
	"errors"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	common2 "github.com/zerok-ai/zk-utils-go/common"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/obfuscation/model"
	zkredis "github.com/zerok-ai/zk-utils-go/storage/redis"
	dbNames "github.com/zerok-ai/zk-utils-go/storage/redis/clientDBNames"
)

var obfuscationLogTag = "ObfuscationHandler"

type ObfuscationObj struct {
	Obfuscations []model.RuleOperator `json:"obfuscations"`
	Deleted      []string             `json:"deleted_obfuscations"`
	Disabled     []string             `json:"disabled_obfuscations"`
}

type ObfuscationApiResponse struct {
	Payload ObfuscationObj      `json:"payload"`
	Error   *zkhttp.ZkHttpError `json:"error,omitempty"`
}

func (o ObfuscationApiResponse) GetError() *zkhttp.ZkHttpError {
	return o.Error
}

type ObfuscationHandler struct {
	VersionedStore     *zkredis.VersionedStore[model.RuleOperator]
	config             config.ZkOperatorConfig
	latestUpdateTime   string
	zkCloudSyncHandler *ZkCloudSyncHandler[ObfuscationApiResponse]
}

func (h *ObfuscationHandler) Init(cfg config.ZkOperatorConfig) error {
	store, err := zkredis.GetVersionedStore[model.RuleOperator](&cfg.Redis, dbNames.ObfuscationRulesDBName, common.RedisSyncInterval)
	if err != nil {
		return err
	}
	h.VersionedStore = store
	h.config = cfg
	h.latestUpdateTime = "0"

	syncHandler := ZkCloudSyncHandler[ObfuscationApiResponse]{}
	syncHandler.Init(cfg, cfg.ObfuscationSync.PollingInterval, "obfuscation_sync", h.periodicSync)
	h.zkCloudSyncHandler = &syncHandler

	return nil
}

func (h *ObfuscationHandler) periodicSync() {
	logger.Debug(integrationLogTag, "Sync obfuscations triggered.")
	h.updateObfuscations(h.config, true)
}

func (h *ObfuscationHandler) CleanUpOnKill() error {
	logger.Debug(scenarioLogTag, "Kill method in scenario rules.")
	h.zkCloudSyncHandler.StopSync()
	return nil
}

func (h *ObfuscationHandler) IsHealthy() bool {
	return true
}

func (h *ObfuscationHandler) StartPeriodicSync() {
	h.updateObfuscations(h.config, true)
	h.zkCloudSyncHandler.StartSync()
}

func (h *ObfuscationHandler) updateObfuscations(operatorConfig config.ZkOperatorConfig, refreshAuthToken bool) {
	logger.Debug(obfuscationLogTag, "Update obfuscations method called.", refreshAuthToken)
	path := h.config.ObfuscationSync.Path
	obfuscationResponse, err := h.zkCloudSyncHandler.GetDataFromZkCloud(path, h.latestUpdateTime)
	if err != nil {
		if errors.Is(err, RefreshAuthTokenError) {
			logger.Debug(obfuscationLogTag, "Ignore this, since we are making another call after refreshing auth token.")
			return
		}
		logger.Error(obfuscationLogTag, "Error while getting obfuscationResponse from zkcloud ", err)
		return
	}
	latestUpdateTime, err := h.processObfuscations(obfuscationResponse)
	if err != nil {
		logger.Error(obfuscationLogTag, "Error while saving obfuscationResponse to redis ", err)
	} else {
		h.latestUpdateTime = latestUpdateTime
	}
}

func (h *ObfuscationHandler) processObfuscations(rulesApiResponse *ObfuscationApiResponse) (string, error) {
	if rulesApiResponse == nil {
		logger.Error(obfuscationLogTag, "Rules Api response is nil.")
		return "", fmt.Errorf("rules Api response is nil")
	}
	payload := rulesApiResponse.Payload
	var latestUpdateTime int64
	for _, obfuscation := range payload.Obfuscations {
		updatedAt := obfuscation.UpdatedAt

		if updatedAt > latestUpdateTime {
			latestUpdateTime = updatedAt
		}

		logger.Debug(obfuscationLogTag, "obfuscation string ", obfuscation)

		var obfuscationId string

		if common2.IsEmpty(obfuscation.Id) {
			logger.Error(obfuscationLogTag, "id is empty. Ignoring this obfuscation.", obfuscation)
			continue
		} else {
			obfuscationId = obfuscation.Id
		}

		err := h.VersionedStore.SetValue(obfuscationId, obfuscation)
		if err != nil {
			if errors.Is(err, zkredis.LATEST) {
				logger.Info(obfuscationLogTag, "Latest value is already present in redis for obfuscation Id ", obfuscationId)
			} else {
				logger.Error(obfuscationLogTag, "Error while setting obfuscation rule to redis ", err)
				return "", err
			}
		}
	}

	for _, obfuscationId := range payload.Deleted {
		err := h.VersionedStore.Delete(obfuscationId)
		if err != nil {
			logger.Error(obfuscationLogTag, "Error while deleting filter id ", obfuscationId, " from redis ", err)
			return "", err
		}
	}

	for _, obfuscationId := range payload.Disabled {
		err := h.VersionedStore.Delete(obfuscationId)
		if err != nil {
			logger.Error(obfuscationLogTag, "Error while deleting disabled filter id ", obfuscationId, " from redis ", err)
			return "", err
		}
	}

	latestUpdateTimeStr := fmt.Sprintf("%v", latestUpdateTime)

	return latestUpdateTimeStr, nil
}
