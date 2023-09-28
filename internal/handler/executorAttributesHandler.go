package handler

import (
	"encoding/json"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/models"
	"github.com/zerok-ai/zk-operator/internal/storage"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	zktick "github.com/zerok-ai/zk-utils-go/ticker"
	"strconv"
	"time"
)

var LOG_TAG = "ExecutorAttributesHandler"

type ExecutorAttributesHandler struct {
	executorAttributesStore *storage.ExecutorAttributesStore
	OpLogin                 *auth.OperatorLogin
	ticker                  *zktick.TickerTask
	config                  config.ZkOperatorConfig
	latestVersion           string
}

func (h *ExecutorAttributesHandler) Init(executorAttributesStore *storage.ExecutorAttributesStore, OpLogin *auth.OperatorLogin, cfg config.ZkOperatorConfig) {
	h.executorAttributesStore = executorAttributesStore
	h.OpLogin = OpLogin
	h.config = cfg
	h.latestVersion = "0"

	//Creating a timer for periodic scenario
	var duration = time.Duration(cfg.ExecutorAttributesSync.PollingInterval) * time.Second
	h.ticker = zktick.GetNewTickerTask("executor_attributes_sync", duration, h.periodicSync)
}

func (h *ExecutorAttributesHandler) StartPeriodicSync() {
	h.updateExecutorAttributes(h.config, true)
	h.ticker.Start()
}

func (h *ExecutorAttributesHandler) periodicSync() {
	h.updateExecutorAttributes(h.config, false)
}

func (h *ExecutorAttributesHandler) getExecutorAttributesFromZkCloud() (models.ExecutorAttributesResponse, error) {
	responseStr := `
		"executor_attributes": [
			{
				executor: "EBPF",
				version: "1.2",
				protocol: "HTTP",
				attributes: {"status_code": "\"attributes\".\"status_code\""}
			}
		],
		"version": 12356645343,
		"update": true,
	}`
	var response models.ExecutorAttributesResponse
	err := json.Unmarshal([]byte(responseStr), &response)
	return response, err
}

func (h *ExecutorAttributesHandler) updateExecutorAttributes(cfg config.ZkOperatorConfig, forceUpdate bool) {
	var executorAttributesResponse, err = h.getExecutorAttributesFromZkCloud()
	if err != nil {
		logger.Error(LOG_TAG, "Error in getting executor attributes from zk cloud ", err)
		return
	}

	if !(executorAttributesResponse.Update || forceUpdate) {
		return
	}

	logger.Debug(LOG_TAG, "Updating executor attributes.")
	for _, executorAttributes := range executorAttributesResponse.ExecutorAttributes {
		executorVersionKey := executorAttributes.Executor + "_" + executorAttributes.Version + "_" + executorAttributes.Protocol
		err := h.executorAttributesStore.UploadExecutorAttributes(executorVersionKey, executorAttributes.Attributes)
		if err != nil {
			logger.Error(LOG_TAG, "Error in updating executor attributes in redis ", err)
			return
		}

	}
	h.latestVersion = strconv.FormatInt(executorAttributesResponse.Version, 10)
}

func (h *ExecutorAttributesHandler) CleanUpOnKill() error {
	logger.Debug(LOG_TAG, "Kill method in scenario rules.")
	h.executorAttributesStore.Close()
	h.ticker.Stop()
	return nil
}

func (h *ExecutorAttributesHandler) IsHealthy() bool {
	return true
}
