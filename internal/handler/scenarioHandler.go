package handler

import (
	"errors"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	common2 "github.com/zerok-ai/zk-utils-go/common"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/scenario/model"
	zkredis "github.com/zerok-ai/zk-utils-go/storage/redis"
)

var scenarioLogTag = "ScenarioHandler"

var RefreshAuthTokenError = fmt.Errorf("refreshing auth token")

var authTokenExpiredCode = 401

type ScenarioHandler struct {
	VersionedStore     *zkredis.VersionedStore[model.Scenario]
	config             config.ZkOperatorConfig
	latestUpdateTime   string
	zkCloudSyncHandler *ZkCloudSyncHandler[ScenariosApiResponse]
}

type ScenariosApiResponse struct {
	Payload ScenariosObj        `json:"payload"`
	Error   *zkhttp.ZkHttpError `json:"error,omitempty"`
}

func (s ScenariosApiResponse) GetError() *zkhttp.ZkHttpError {
	return s.Error
}

type ScenariosObj struct {
	Scenarios []ScenarioModelResponse `json:"scenarios"`
	Deleted   []string                `json:"deleted_scenario_id,omitempty"`
	Disabled  []string                `json:"disabled_scenario_id,omitempty"`
}

type ScenarioModelResponse struct {
	Scenario   model.Scenario `json:"scenario"`
	CreatedAt  int64          `json:"created_at"`
	DisabledAt *int64         `json:"disabled_at,omitempty"`
	UpdatedAt  int64          `json:"updated_at"`
}

func (h *ScenarioHandler) Init(OpLogin *auth.OperatorLogin, cfg config.ZkOperatorConfig) error {
	store, err := zkredis.GetVersionedStore[model.Scenario](&cfg.Redis, common.RedisScenarioDbName, common.RedisSyncInterval)
	if err != nil {
		return err
	}
	h.VersionedStore = store
	h.config = cfg
	h.latestUpdateTime = "0"

	syncHandler := ZkCloudSyncHandler[ScenariosApiResponse]{}
	syncHandler.Init(OpLogin, cfg, cfg.ScenarioSync.PollingInterval, "scenario_sync", h.periodicSync)
	h.zkCloudSyncHandler = &syncHandler

	return nil
}

func (h *ScenarioHandler) StartPeriodicSync() {
	h.updateScenarios(h.config, true)
	h.zkCloudSyncHandler.StartSync()
}

func (h *ScenarioHandler) periodicSync() {
	logger.Debug(scenarioLogTag, "Sync scenarios triggered.")
	h.updateScenarios(h.config, true)
}

func (h *ScenarioHandler) updateScenarios(cfg config.ZkOperatorConfig, refreshAuthToken bool) {
	logger.Debug(scenarioLogTag, "Update scenarios method called.", refreshAuthToken)
	callback := func() {
		h.updateScenarios(cfg, false)
	}
	scenarioResponse, err := h.zkCloudSyncHandler.GetDataFromZkCloud(h.config.ScenarioSync.Path, callback, h.latestUpdateTime, refreshAuthToken)
	if err != nil {
		if errors.Is(err, RefreshAuthTokenError) {
			logger.Debug(scenarioLogTag, "Ignore this, since we are making another call after refreshing auth token.")
			return
		}
		logger.Error(scenarioLogTag, "Error while getting scenarioResponse from zkcloud ", err)
		return
	}
	latestUpdateTime, err := h.processScenarios(scenarioResponse)
	if err != nil {
		logger.Error(scenarioLogTag, "Error while saving scenarioResponse to redis ", err)
	} else {
		h.latestUpdateTime = latestUpdateTime
	}
}

// This method will parse rules and return the largest version found and any error caught.
func (h *ScenarioHandler) processScenarios(rulesApiResponse *ScenariosApiResponse) (string, error) {
	if rulesApiResponse == nil {
		logger.Error(scenarioLogTag, "Rules Api response is nil.")
		return "", fmt.Errorf("rules Api response is nil")
	}
	payload := rulesApiResponse.Payload
	var latestUpdateTime int64
	for _, scenarioResp := range payload.Scenarios {
		updatedAt := scenarioResp.UpdatedAt

		if updatedAt > latestUpdateTime {
			latestUpdateTime = updatedAt
		}

		logger.Debug(scenarioLogTag, "Scenario string ", scenarioResp)

		var scenarioId string

		if common2.IsEmpty(scenarioResp.Scenario.Id) {
			logger.Error(scenarioLogTag, "Scenario id is empty. Ignoring this scenario.", scenarioResp.Scenario)
			continue
		} else {
			scenarioId = scenarioResp.Scenario.Id
		}

		err := h.VersionedStore.SetValue(scenarioId, scenarioResp.Scenario)
		if err != nil {
			if errors.Is(err, zkredis.LATEST) {
				logger.Info(scenarioLogTag, "Latest value is already present in redis for scenario Id ", scenarioId)
			} else {
				logger.Error(scenarioLogTag, "Error while setting filter rule to redis ", err)
				return "", err
			}
		}
	}

	for _, scenarioId := range payload.Deleted {
		err := h.VersionedStore.Delete(scenarioId)
		if err != nil {
			logger.Error(scenarioLogTag, "Error while deleting filter id ", scenarioId, " from redis ", err)
			return "", err
		}
	}

	for _, scenarioId := range payload.Disabled {
		err := h.VersionedStore.Delete(scenarioId)
		if err != nil {
			logger.Error(LOG_TAG, "Error while deleting disabled filter id ", scenarioId, " from redis ", err)
			return "", err
		}
	}

	latestUpdateTimeStr := fmt.Sprintf("%v", latestUpdateTime)

	return latestUpdateTimeStr, nil
}

func (h *ScenarioHandler) CleanUpOnKill() error {
	logger.Debug(scenarioLogTag, "Kill method in scenario rules.")
	h.zkCloudSyncHandler.StopSync()
	return nil
}

func (h *ScenarioHandler) IsHealthy() bool {
	return true
}
