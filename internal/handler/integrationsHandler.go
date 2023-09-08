package handler

import (
	"encoding/json"
	"errors"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	"github.com/zerok-ai/zk-utils-go/interfaces"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	zkredis "github.com/zerok-ai/zk-utils-go/storage/redis"
	"time"
)

var integrationResponseTag = "IntegrationResponseTag"

type IntegrationResponseObj struct {
	ID             int             `json:"id"`
	ClusterId      string          `json:"cluster_id"`
	Type           string          `json:"type"`
	URL            string          `json:"url"`
	Authentication json.RawMessage `json:"authentication"`
	Level          string          `json:"level"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Deleted        bool            `json:"deleted"`
	Disabled       bool            `json:"disabled"`
}

type IntegrationApiResponse struct {
	Payload IntegrationResponse `json:"payload"`
	Error   *zkhttp.ZkHttpError `json:"error,omitempty"`
}

func (r IntegrationApiResponse) GetError() *zkhttp.ZkHttpError {
	return r.Error
}

type IntegrationResponse struct {
	Response []IntegrationResponseObj `json:"integrations"`
}

func (i IntegrationResponseObj) Equals(other interfaces.ZKComparable) bool {
	return false
}

type IntegrationsHandler struct {
	VersionedStore     *zkredis.VersionedStore[IntegrationResponseObj]
	config             config.ZkOperatorConfig
	latestUpdateTime   string
	zkCloudSyncHandler *ZkCloudSyncHandler[IntegrationApiResponse]
}

func (h *IntegrationsHandler) Init(OpLogin *auth.OperatorLogin, cfg config.ZkOperatorConfig) error {
	store, err := zkredis.GetVersionedStore[IntegrationResponseObj](&cfg.Redis, common.RedisIntegrationsDbName, common.RedisSyncInterval)
	if err != nil {
		return err
	}
	h.VersionedStore = store
	h.config = cfg
	h.latestUpdateTime = "0"

	syncHandler := ZkCloudSyncHandler[IntegrationApiResponse]{}
	syncHandler.Init(OpLogin, cfg, cfg.ScenarioSync.PollingInterval, "integration_sync", h.periodicSync)
	h.zkCloudSyncHandler = &syncHandler
	return nil
}

func (h *IntegrationsHandler) StartPeriodicSync() {
	h.updateIntegrations(h.config, true)
	h.zkCloudSyncHandler.StartSync()
}

func (h *IntegrationsHandler) periodicSync() {
	logger.Debug(integrationResponseTag, "Sync integrations triggered.")
	h.updateIntegrations(h.config, true)
}

func (h *IntegrationsHandler) updateIntegrations(cfg config.ZkOperatorConfig, refreshAuthToken bool) {
	logger.Debug(integrationResponseTag, "Update integrations method called.", refreshAuthToken)
	callback := func() {
		h.updateIntegrations(cfg, false)
	}
	integrationResponse, err := h.zkCloudSyncHandler.GetDataFromZkCloud(h.config.IntegrationSync.Path, callback, h.latestUpdateTime, refreshAuthToken)
	if err != nil {
		if errors.Is(err, RefreshAuthTokenError) {
			logger.Debug(integrationResponseTag, "Ignore this, since we are making another call after refreshing auth token.")
			return
		}
		logger.Error(integrationResponseTag, "Error while getting integrationResponse from zkcloud ", err)
		return
	}
	latestUpdateTime, err := h.processIntegrations(integrationResponse)
	if err != nil {
		logger.Error(integrationResponseTag, "Error while saving integrationResponse to redis ", err)
	} else {
		h.latestUpdateTime = latestUpdateTime
	}
}

func (h *IntegrationsHandler) processIntegrations(response *IntegrationApiResponse) (string, error) {
	return "", nil
}

func (h *IntegrationsHandler) CleanUpOnKill() error {
	logger.Debug(integrationResponseTag, "Kill method in scenario rules.")
	h.zkCloudSyncHandler.StopSync()
	return nil
}

func (h *IntegrationsHandler) IsHealthy() bool {
	return true
}
