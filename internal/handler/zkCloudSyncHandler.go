package handler

import (
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	zktick "github.com/zerok-ai/zk-utils-go/ticker"
	"io"
	"net/http"
	"strconv"
	"time"
)

var cloudSyncLogTag = "zkCloudSyncHandler"

type ApiResponse interface {
	GetError() *zkhttp.ZkHttpError
}

type ZkCloudSyncHandler[T ApiResponse] struct {
	OpLogin  *auth.OperatorLogin
	config   config.ZkOperatorConfig
	ticker   *zktick.TickerTask
	TaskName string
}

func (h *ZkCloudSyncHandler[T]) Init(OpLogin *auth.OperatorLogin, cfg config.ZkOperatorConfig, pollingInterval int, taskName string, task func()) {
	h.OpLogin = OpLogin
	h.config = cfg
	h.TaskName = taskName
	//Creating a timer for periodic sync
	var duration = time.Duration(pollingInterval) * time.Second
	h.ticker = zktick.GetNewTickerTask(taskName, duration, task)
}

// TODO: Breakdown this method.
func (h *ZkCloudSyncHandler[T]) GetDataFromZkCloud(urlPath string, callback auth.RefreshTokenCallback, latestUpdateTime string, refreshAuthToken bool) (*T, error) {
	port := h.config.ZkCloud.Port
	protocol := "http"
	if port == "443" {
		protocol = "https"
	}

	logger.Debug(cloudSyncLogTag, h.TaskName, " from zk cloud.")

	baseURL := protocol + "://" + h.config.ZkCloud.Host + ":" + h.config.ZkCloud.Port + urlPath

	//Adding query params
	url := fmt.Sprintf("%s?%s=%s", baseURL, "last_sync_ts", latestUpdateTime)

	logger.Debug(cloudSyncLogTag, "Url for ", h.TaskName, " is ", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(cloudSyncLogTag, "Error creating request:", err, " for task ", h.TaskName)
		return nil, err
	}

	if h.OpLogin.GetOperatorToken() == "" {
		if refreshAuthToken {
			logger.Debug(cloudSyncLogTag, "Operator auth token is not present. Getting the auth token.", " for task ", h.TaskName)
			err := h.refreshAuthToken(callback)
			if err != nil {
				return nil, err
			}
			return nil, RefreshAuthTokenError
		} else {
			logger.Debug(cloudSyncLogTag, "Operator auth token is empty. Refresh auth token is false.", " for task ", h.TaskName)
			return nil, fmt.Errorf("operator token is empty")
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(common.OperatorTokenHeaderKey, h.OpLogin.GetOperatorToken())

	resp, err := utils.RouteRequestFromWspClient(req, h.config)
	if err != nil {
		logger.Error(cloudSyncLogTag, "Error sending request for cloud sync ", err, " for task ", h.TaskName)
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	if statusCode == authTokenExpiredCode {
		if refreshAuthToken {
			logger.Error(cloudSyncLogTag, "Operator auth token has expired. Refreshing the auth token for task ", h.TaskName)
			err := h.refreshAuthToken(callback)
			if err != nil {
				return nil, err
			}
			return nil, RefreshAuthTokenError
		} else {
			logger.Error(cloudSyncLogTag, "Operator auth token has expired. Refresh auth token is false for task ", h.TaskName)
			message := "oerator auth token expired for task " + h.TaskName
			return nil, fmt.Errorf(message)
		}
	}

	if !utils.RespCodeIsOk(statusCode) {
		message := "response code is not ok for get sync api - " + strconv.Itoa(resp.StatusCode) + " for task " + h.TaskName
		logger.Error(cloudSyncLogTag, message)
		return nil, fmt.Errorf(message)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(cloudSyncLogTag, "Error reading response from sync api :", err, " for task ", h.TaskName)
		return nil, err
	}

	var apiResponse T

	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		logger.Error(cloudSyncLogTag, "Error while unmarshalling sync api response :", err, " for task ", h.TaskName)
		return nil, err
	}

	responseError := apiResponse.GetError()
	if responseError != nil {
		message := "found error in response " + responseError.Message + " for task " + h.TaskName
		logger.Error(cloudSyncLogTag, message)
		return nil, fmt.Errorf(message)
	}

	return &apiResponse, nil
}

func (h *ZkCloudSyncHandler[T]) refreshAuthToken(callback auth.RefreshTokenCallback) error {
	err := h.OpLogin.RefreshOperatorToken(callback)
	if err != nil {
		logger.Error(cloudSyncLogTag, "Error while refreshing auth token ", err)
	}
	return err
}

func (h *ZkCloudSyncHandler[T]) StartSync() {
	h.ticker.Start()
}

func (h *ZkCloudSyncHandler[T]) StopSync() {
	h.ticker.Stop()
}
