package handler

import (
	"errors"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	model "github.com/zerok-ai/zk-utils-go/integration/model"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	zkredis "github.com/zerok-ai/zk-utils-go/storage/redis"
	dbNames "github.com/zerok-ai/zk-utils-go/storage/redis/clientDBNames"
	"strings"
)

var integrationLogTag = "IntegrationLogTag"

type IntegrationApiResponse struct {
	Payload IntegrationResponse `json:"payload"`
	Error   *zkhttp.ZkHttpError `json:"error,omitempty"`
}

func (r IntegrationApiResponse) GetError() *zkhttp.ZkHttpError {
	return r.Error
}

type IntegrationResponse struct {
	Response []model.IntegrationResponseObj `json:"integrations"`
}

type IntegrationsHandler struct {
	VersionedStore     *zkredis.VersionedStore[model.IntegrationResponseObj]
	config             config.ZkOperatorConfig
	latestUpdateTime   string
	zkCloudSyncHandler *ZkCloudSyncHandler[IntegrationApiResponse]
	clusterId          string
}

func (h *IntegrationsHandler) Init(cfg config.ZkOperatorConfig, clusterId string) error {
	store, err := zkredis.GetVersionedStore[model.IntegrationResponseObj](&cfg.Redis, dbNames.IntegrationDetailsDBName, common.RedisSyncInterval)
	if err != nil {
		return err
	}
	h.VersionedStore = store
	h.config = cfg
	h.latestUpdateTime = ""
	h.clusterId = clusterId

	syncHandler := ZkCloudSyncHandler[IntegrationApiResponse]{}
	syncHandler.Init(cfg, cfg.IntegrationSync.PollingInterval, "integration_sync", h.periodicSync)
	h.zkCloudSyncHandler = &syncHandler
	return nil
}

func (h *IntegrationsHandler) StartPeriodicSync() {
	h.updateIntegrations(h.config)
	h.zkCloudSyncHandler.StartSync()
}

func (h *IntegrationsHandler) periodicSync() {
	logger.Debug(integrationLogTag, "Sync integrations triggered.")
	h.updateIntegrations(h.config)
}

func (h *IntegrationsHandler) updateIntegrations(cfg config.ZkOperatorConfig) {
	logger.Debug(integrationLogTag, "Update integrations method called.")
	path := h.config.IntegrationSync.Path
	path = strings.ReplaceAll(path, "<clusterid>", h.clusterId)
	integrationResponse, err := h.zkCloudSyncHandler.GetDataFromZkCloud(path, h.latestUpdateTime)
	if err != nil {
		if errors.Is(err, RefreshAuthTokenError) {
			logger.Debug(integrationLogTag, "Ignore this, since we are making another call after refreshing auth token.")
			return
		}
		logger.Error(integrationLogTag, "Error while getting integrationResponse from zkcloud ", err)
		return
	}
	latestUpdateTime, err := h.processIntegrations(integrationResponse)
	if err != nil {
		logger.Error(integrationLogTag, "Error while saving integrationResponse to redis ", err)
	} else {
		h.latestUpdateTime = latestUpdateTime
	}
}

func (h *IntegrationsHandler) processIntegrations(response *IntegrationApiResponse) (string, error) {
	if response == nil {
		logger.Error(integrationLogTag, "integrations Api response is nil.")
		return "", fmt.Errorf("integrations Api response is nil")
	}
	//Deleting existing integrations.
	err := h.VersionedStore.DeleteAllKeys()
	if err != nil {
		logger.Error(integrationLogTag, "Error while deleting all keys from redis ", err)
	}
	payload := response.Payload
	for _, integration := range payload.Response {

		err := h.VersionedStore.SetValue(integration.ID, integration)
		if err != nil {
			if errors.Is(err, zkredis.LATEST) {
				logger.Info(integrationLogTag, "Latest value is already present in redis for integration Id ", integration.ID)
			} else {
				logger.Error(integrationLogTag, "Error while setting filter integration to redis ", err)
				return "", err
			}
		}
	}

	return "", nil
}

func (h *IntegrationsHandler) CleanUpOnKill() error {
	logger.Debug(integrationLogTag, "Kill method in scenario rules.")
	h.zkCloudSyncHandler.StopSync()
	return nil
}

func (h *IntegrationsHandler) IsHealthy() bool {
	return true
}
