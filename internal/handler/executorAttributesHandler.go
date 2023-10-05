package handler

import (
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/models"
	"github.com/zerok-ai/zk-operator/internal/storage"
	"github.com/zerok-ai/zk-operator/internal/utils"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	zktick "github.com/zerok-ai/zk-utils-go/ticker"
	"io"
	"net/http"
	"strconv"
	"time"
)

var LOG_TAG = "ExecutorAttributesHandler"

type ExecutorAttributesHandler struct {
	executorAttributesStore *storage.ExecutorAttributesStore
	OpLogin                 *auth.OperatorLogin
	ticker                  *zktick.TickerTask
	config                  config.ZkOperatorConfig
}

func (h *ExecutorAttributesHandler) Init(executorAttributesStore *storage.ExecutorAttributesStore, OpLogin *auth.OperatorLogin, cfg config.ZkOperatorConfig) {
	h.executorAttributesStore = executorAttributesStore
	h.OpLogin = OpLogin
	h.config = cfg

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

func (h *ExecutorAttributesHandler) updateLastSyncTime(latestVersion string) {
	err := h.executorAttributesStore.UpdateLastSyncTime(latestVersion)
	if err != nil {
		logger.Error(LOG_TAG, "Error in updating latest version in redis ", err)
	}
}

func (h *ExecutorAttributesHandler) getExecutorAttributesPayloadFromZkCloud() (*models.ExecutorAttributesPayload, error) {
	lastSyncVersion := h.executorAttributesStore.GetLastSyncVersion()
	urlPath := "/v1/o/cluster/attribute?version=" + lastSyncVersion
	port := h.config.ZkCloud.Port
	protocol := "http"
	if port == "443" {
		protocol = "https"
	}

	baseURL := protocol + "://" + h.config.ZkCloud.Host + ":" + h.config.ZkCloud.Port + urlPath

	url := baseURL

	logger.Debug(LOG_TAG, "Url is ", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(LOG_TAG, "Error creating request:", err)
		return nil, err
	}

	if h.OpLogin.GetOperatorToken() == "" {
		logger.Debug(LOG_TAG, "Operator auth token is not present. Getting the auth token.")
		err := h.refreshAuthToken(h.periodicSync)
		if err != nil {
			return nil, err
		}
		return nil, RefreshAuthTokenError
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(common.OperatorTokenHeaderKey, h.OpLogin.GetOperatorToken())

	resp, err := utils.RouteRequestFromWspClient(req, h.config)
	if err != nil {
		logger.Error(LOG_TAG, "Error sending request for cloud sync ", err)
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	if statusCode == authTokenExpiredCode {
		logger.Error(LOG_TAG, "Operator auth token has expired. Refreshing the auth token")
		err := h.refreshAuthToken(h.periodicSync)
		if err != nil {
			return nil, err
		}
		return nil, RefreshAuthTokenError
	}

	if !utils.RespCodeIsOk(statusCode) {
		message := "response code is not ok for get sync api - " + strconv.Itoa(resp.StatusCode)
		logger.Error(LOG_TAG, message)
		return nil, fmt.Errorf(message)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(LOG_TAG, "Error reading response from sync api :", err)
		return nil, err
	}

	var apiResponse models.ExecutorAttributesAPIResponse

	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		logger.Error(LOG_TAG, "Error while unmarshalling sync api response :", err)
		return nil, err
	}

	responseError := apiResponse.Error
	if responseError != nil {
		message := "found error in response " + responseError.Message
		logger.Error(LOG_TAG, message)
		return nil, fmt.Errorf(message)
	}

	respStr, err := json.Marshal(apiResponse)
	logger.Debug(LOG_TAG, "Api response is ", string(respStr))

	return &apiResponse.Data, nil
}

func (h *ExecutorAttributesHandler) refreshAuthToken(callback auth.RefreshTokenCallback) error {
	err := h.OpLogin.RefreshOperatorToken(callback)
	if err != nil {
		logger.Error(cloudSyncLogTag, "Error while refreshing auth token ", err)
	}
	return err
}

func (h *ExecutorAttributesHandler) updateExecutorAttributes(cfg config.ZkOperatorConfig, forceUpdate bool) {
	logger.Debug(LOG_TAG, "In executor attributes update method")
	var executorAttributesPayload, err = h.getExecutorAttributesPayloadFromZkCloud()
	if err != nil {
		logger.Error(LOG_TAG, "Error in getting executor attributes from zk cloud ", err)
		return
	}

	if !(executorAttributesPayload.Update || forceUpdate) {
		return
	}

	logger.Debug(LOG_TAG, "Updating executor attributes.")
	for _, executorAttributes := range executorAttributesPayload.ExecutorAttributes {
		executorVersionKey := executorAttributes.Executor + "_" + executorAttributes.Version + "_" + executorAttributes.Protocol
		err := h.executorAttributesStore.UploadExecutorAttributes(executorVersionKey, executorAttributes.Attributes)
		if err != nil {
			logger.Error(LOG_TAG, "Error in updating executor attributes in redis ", err)
			return
		}
	}
	h.updateLastSyncTime(strconv.FormatInt(executorAttributesPayload.Version, 10))
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
