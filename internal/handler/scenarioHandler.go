package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/utils"
	common2 "github.com/zerok-ai/zk-utils-go/common"
	zkhttp "github.com/zerok-ai/zk-utils-go/http"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"github.com/zerok-ai/zk-utils-go/scenario/model"
	zktick "github.com/zerok-ai/zk-utils-go/ticker"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/zerok-ai/zk-operator/internal/config"
	zkredis "github.com/zerok-ai/zk-utils-go/storage/redis"
)

var LOG_TAG = "ScenarioHandler"

var RefreshAuthTokenError = fmt.Errorf("refreshing auth token")

var authTokenExpiredCode = 401

type ScenarioHandler struct {
	VersionedStore   *zkredis.VersionedStore[model.Scenario]
	OpLogin          *auth.OperatorLogin
	ticker           *zktick.TickerTask
	config           config.ZkOperatorConfig
	latestUpdateTime string
}

type ScenariosApiResponse struct {
	Payload ScenariosObj        `json:"payload"`
	Error   *zkhttp.ZkHttpError `json:"error,omitempty"`
}

type ScenariosObj struct {
	Scenarios []ScenarioModelResponse `json:"scenarios"`
	Deleted   []string                `json:"deleted_scenario_id,omitempty"`
}

type ScenarioModelResponse struct {
	Scenario   model.Scenario `json:"scenario"`
	CreatedAt  int64          `json:"created_at"`
	DisabledAt *int64         `json:"disabled_at,omitempty"`
	UpdatedAt  int64          `json:"updated_at"`
}

func (h *ScenarioHandler) Init(VersionedStore *zkredis.VersionedStore[model.Scenario], OpLogin *auth.OperatorLogin, cfg config.ZkOperatorConfig) {
	h.VersionedStore = VersionedStore
	h.OpLogin = OpLogin
	h.config = cfg
	h.latestUpdateTime = "0"

	//Creating a timer for periodic scenario
	var duration = time.Duration(cfg.ScenarioSync.PollingInterval) * time.Second
	h.ticker = zktick.GetNewTickerTask("scenario_sync", duration, h.periodicSync)
}

func (h *ScenarioHandler) StartPeriodicSync() {
	h.updateScenarios(h.config, true)
	h.ticker.Start()
}

func (h *ScenarioHandler) periodicSync() {
	logger.Debug(LOG_TAG, "Sync scenarios triggered.")
	h.updateScenarios(h.config, true)
}

func (h *ScenarioHandler) updateScenarios(cfg config.ZkOperatorConfig, refreshAuthToken bool) {
	logger.Debug(LOG_TAG, "Update scenarios method called.", refreshAuthToken)
	rules, err := h.getScenariosFromZkCloud(cfg, refreshAuthToken)
	if err != nil {
		if errors.Is(err, RefreshAuthTokenError) {
			logger.Debug(LOG_TAG, "Ignore this, since we are making another call after refreshing auth token.")
			return
		}
		logger.Error(LOG_TAG, "Error while getting rules from zkcloud ", err)
		return
	}
	latestUpdateTime, err := h.processScenarios(rules)
	if err != nil {
		logger.Error(LOG_TAG, "Error while savign rules to redis ", err)
	} else {
		h.latestUpdateTime = latestUpdateTime
	}
}

func (h *ScenarioHandler) getScenariosFromZkCloud(cfg config.ZkOperatorConfig, refreshAuthToken bool) (*ScenariosApiResponse, error) {

	port := cfg.ZkCloud.Port
	protocol := "http"
	if port == "443" {
		protocol = "https"
	}

	logger.Debug(LOG_TAG, "Get rules from zk cloud.")

	baseURL := protocol + "://" + cfg.ZkCloud.Host + ":" + cfg.ZkCloud.Port + cfg.ScenarioSync.Path

	//Adding query params
	url := fmt.Sprintf("%s?%s=%s", baseURL, "last_sync_ts", h.latestUpdateTime)

	logger.Debug(LOG_TAG, "Url for scenario sync ", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(LOG_TAG, "Error creating request:", err)
		return nil, err
	}

	if h.OpLogin.GetOperatorToken() == "" {
		if refreshAuthToken {
			logger.Debug(LOG_TAG, "Operator auth token is not present. Getting the auth token.")
			err := h.refreshAuthToken(cfg)
			if err != nil {
				return nil, err
			}
			return nil, RefreshAuthTokenError
		} else {
			logger.Debug(LOG_TAG, "Operator auth token is empty. Refresh auth token is false.")
			return nil, fmt.Errorf("operator token is empty")
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(common.OperatorTokenHeaderKey, h.OpLogin.GetOperatorToken())

	resp, err := utils.RouteRequestFromWspClient(req, h.config)
	if err != nil {
		logger.Error(LOG_TAG, "Error sending request for rules api :", err)
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	if statusCode == authTokenExpiredCode {
		if refreshAuthToken {
			logger.Error(LOG_TAG, "Operator auth token has expired. Refreshing the auth token.")
			err := h.refreshAuthToken(cfg)
			if err != nil {
				return nil, err
			}
			return nil, RefreshAuthTokenError
		} else {
			logger.Error(LOG_TAG, "Operator auth token has expired. Refresh auth token is false.")
			return nil, fmt.Errorf("operator auth token has expired")
		}
	}

	if !utils.RespCodeIsOk(statusCode) {
		message := "response code is not ok for get scenario api - " + strconv.Itoa(resp.StatusCode)
		logger.Error(LOG_TAG, message)
		return nil, fmt.Errorf(message)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(LOG_TAG, "Error reading response from rules api :", err)
		return nil, err
	}

	//logger.Debug(LOG_TAG, "Scenario response body ", body)

	var apiResponse ScenariosApiResponse

	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		logger.Error(LOG_TAG, "Error while unmarshalling rules api response :", err)
		return nil, err
	}

	if apiResponse.Error != nil {
		message := "found error in get scenario api response " + apiResponse.Error.Message
		logger.Error(LOG_TAG, message)
		return nil, fmt.Errorf(message)
	}

	return &apiResponse, nil
}

func (h *ScenarioHandler) refreshAuthToken(cfg config.ZkOperatorConfig) error {
	err := h.OpLogin.RefreshOperatorToken(func() {
		h.updateScenarios(cfg, false)
	})
	if err != nil {
		logger.Error(LOG_TAG, "Error while refreshing auth token ", err)
	}
	return err
}

// This method will parse rules and return the largest version found and any error caught.
func (h *ScenarioHandler) processScenarios(rulesApiResponse *ScenariosApiResponse) (string, error) {
	if rulesApiResponse == nil {
		logger.Error(LOG_TAG, "Rules Api response is nil.")
		return "", fmt.Errorf("rules Api response is nil")
	}
	payload := rulesApiResponse.Payload
	var latestUpdateTime int64
	for _, scenarioResp := range payload.Scenarios {
		updatedAt := scenarioResp.UpdatedAt

		if updatedAt > latestUpdateTime {
			latestUpdateTime = updatedAt
		}

		logger.Debug(LOG_TAG, "Scenario string ", scenarioResp)

		var scenarioId string

		if common2.IsEmpty(scenarioResp.Scenario.Id) {
			logger.Error(LOG_TAG, "Scenario id is empty. Ignoring this scenario.", scenarioResp.Scenario)
			continue
		} else {
			scenarioId = scenarioResp.Scenario.Id
		}

		err := h.VersionedStore.SetValue(scenarioId, scenarioResp.Scenario)
		if err != nil {
			if errors.Is(err, zkredis.LATEST) {
				logger.Info(LOG_TAG, "Latest value is already present in redis for scenario Id ", scenarioId)
			} else {
				logger.Error(LOG_TAG, "Error while setting filter rule to redis ", err)
				return "", err
			}
		}
	}

	for _, scenarioId := range payload.Deleted {
		err := h.VersionedStore.Delete(scenarioId)
		if err != nil {
			logger.Error(LOG_TAG, "Error while deleting filter id ", scenarioId, " from redis ", err)
			return "", err
		}
	}

	latestUpdateTimeStr := fmt.Sprintf("%v", latestUpdateTime)

	return latestUpdateTimeStr, nil
}

func (h *ScenarioHandler) CleanUpOnkill() error {
	logger.Debug(LOG_TAG, "Kill method in scenario rules.")
	h.ticker.Stop()
	return nil
}
