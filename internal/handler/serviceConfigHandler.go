package handler

import (
	"encoding/json"
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/config"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"strings"
)

var serviceConfigTag = "ServiceConfigHandler"

type ServiceConfigHandler struct {
	ZkConfig           *config.ZkOperatorConfig
	ConfigData         map[string]json.RawMessage
	config             config.ZkOperatorConfig
	zkCloudSyncHandler *ZkCloudSyncHandler[ConfigApiResponse]
	clusterId          string
}

type ConfigApiResponse struct {
	Payload map[string]json.RawMessage `json:"payload"`
	Error   *zkhttp.ZkHttpError        `json:"error,omitempty"`
}

func (c ConfigApiResponse) GetError() *zkhttp.ZkHttpError {
	return c.Error
}

func (h *ServiceConfigHandler) Init(cfg config.ZkOperatorConfig, clusterId string) error {
	h.config = cfg
	syncHandler := ZkCloudSyncHandler[ConfigApiResponse]{}
	syncHandler.Init(cfg, cfg.ConfigurationSync.PollingInterval, "configuration_sync", h.periodicSync)
	h.zkCloudSyncHandler = &syncHandler
	h.clusterId = clusterId
	return nil
}

func (h *ServiceConfigHandler) StartPeriodicSync() {
	h.updateServiceConfig(true)
	h.zkCloudSyncHandler.StartSync()
}

func (h *ServiceConfigHandler) periodicSync() {
	logger.Debug(serviceConfigTag, "Sync configuration triggered.")
	h.updateServiceConfig(true)
}

func (h *ServiceConfigHandler) updateServiceConfig(refreshAuthToken bool) {
	logger.Debug(serviceConfigTag, "Update configurations method called.", refreshAuthToken)
	path := h.config.ConfigurationSync.CloudPath
	path = strings.ReplaceAll(path, "<clusterid>", h.clusterId)
	serviceConfigResponse, err := h.zkCloudSyncHandler.GetDataFromZkCloud(path, "")
	if err != nil {
		if errors.Is(err, RefreshAuthTokenError) {
			logger.Debug(serviceConfigTag, "Ignore this, since we are making another call after refreshing auth token.")
			return
		}
		logger.Error(serviceConfigTag, "Error while getting serviceConfigResponse from zkcloud ", err)
		return
	} else if serviceConfigResponse != nil {
		h.ConfigData = serviceConfigResponse.Payload
	}

}

// Handler to give config data to other services in the client cluster.
func (h *ServiceConfigHandler) Handler(ctx iris.Context) {
	var response json.RawMessage
	svcName := ctx.URLParam("svc")
	response, _ = h.ConfigData[svcName]
	zkHttpResponse := zkhttp.ToZkResponse[json.RawMessage](200, response, nil, nil)
	ctx.StatusCode(zkHttpResponse.Status)
	ctx.JSON(zkHttpResponse)
}

func (h *ServiceConfigHandler) CleanUpOnKill() error {
	logger.Debug(serviceConfigTag, "Nothing to clean here.")
	h.zkCloudSyncHandler.StopSync()
	return nil
}

func (h *ServiceConfigHandler) IsHealthy() bool {
	return len(h.ConfigData) > 0
}
