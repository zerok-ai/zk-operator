package handler

import (
	"encoding/json"
	"fmt"
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
	config   config.ZkOperatorConfig
	ticker   *zktick.TickerTask
	TaskName string
}

func (h *ZkCloudSyncHandler[T]) Init(cfg config.ZkOperatorConfig, pollingInterval int, taskName string, task func()) {
	h.config = cfg
	h.TaskName = taskName
	//Creating a timer for periodic sync
	var duration = time.Duration(pollingInterval) * time.Second
	h.ticker = zktick.GetNewTickerTask(taskName, duration, task)
}

// TODO: Breakdown this method.
func (h *ZkCloudSyncHandler[T]) GetDataFromZkCloud(urlPath string, latestUpdateTime string) (*T, error) {
	port := h.config.ZkCloud.Port
	protocol := "http"
	if port == "443" {
		protocol = "https"
	}

	logger.Debug(cloudSyncLogTag, h.TaskName, " from zk cloud.")

	baseURL := protocol + "://" + h.config.ZkCloud.Host + ":" + h.config.ZkCloud.Port + urlPath

	url := baseURL

	if len(latestUpdateTime) > 0 {
		//Adding query params
		url = fmt.Sprintf("%s?%s=%s", baseURL, "last_sync_ts", latestUpdateTime)
	}

	logger.Debug(cloudSyncLogTag, "Url for ", h.TaskName, " is ", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(cloudSyncLogTag, "Error creating request:", err, " for task ", h.TaskName)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := utils.RouteRequestFromWspClient(req, h.config)
	if err != nil {
		logger.Error(cloudSyncLogTag, "Error sending request for cloud sync ", err, " for task ", h.TaskName)
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

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

	logger.Debug(cloudSyncLogTag, "Api response for ", h.TaskName, " is ", apiResponse)

	return &apiResponse, nil
}

func (h *ZkCloudSyncHandler[T]) StartSync() {
	h.ticker.Start()
}

func (h *ZkCloudSyncHandler[T]) StopSync() {
	h.ticker.Stop()
}
