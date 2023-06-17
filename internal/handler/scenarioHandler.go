package handler

import (
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/utils"
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

var authTokenExpiredCode = 401

type ScenarioHandler struct {
	VersionedStore *zkredis.VersionedStore[model.Scenario]
	OpLogin        *auth.OperatorLogin
	ticker         *zktick.TickerTask
	config         config.ZkOperatorConfig
	rulesVersion   string
}

type ScenariosApiResponse struct {
	Payload ScenariosObj        `json:"payload"`
	Error   *zkhttp.ZkHttpError `json:"error,omitempty"`
}

type ScenariosObj struct {
	Scenarios []model.Scenario `json:"scenarios"`
	Deleted   []string         `json:"deleted,omitempty"`
}

func (h *ScenarioHandler) Init(VersionedStore *zkredis.VersionedStore[model.Scenario], OpLogin *auth.OperatorLogin, cfg config.ZkOperatorConfig) {
	h.VersionedStore = VersionedStore
	h.OpLogin = OpLogin
	h.config = cfg
	h.rulesVersion = "0"

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
	rules, err := h.getScenariosFromZkCloud(cfg, refreshAuthToken)
	if err != nil {
		logger.Error(LOG_TAG, "Error while getting rules from zkcloud ", err)
		return
	}
	latestVersion, err := h.processScenarios(rules)
	if err != nil {
		logger.Error(LOG_TAG, "Error while savign rules to redis ", err)
	} else {
		h.rulesVersion = latestVersion
	}
}

func (h *ScenarioHandler) getScenariosFromZkCloud(cfg config.ZkOperatorConfig, refreshAuthToken bool) (*ScenariosApiResponse, error) {

	logger.Debug(LOG_TAG, "Get rules from zk cloud.")

	baseURL := "http://" + cfg.ZkCloud.Host + ":" + cfg.ZkCloud.Port + cfg.ScenarioSync.Path

	//Adding query params
	url := fmt.Sprintf("%s?%s=%s", baseURL, "version", h.rulesVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(LOG_TAG, "Error creating request:", err)
		return nil, err
	}

	if h.OpLogin.GetOperatorToken() == "" {
		if refreshAuthToken {
			logger.Debug(LOG_TAG, "Operator auth token is expired. Refreshing the auth token.")
			return nil, h.refreshAuthToken(cfg)
		} else {
			logger.Debug(LOG_TAG, "Operator auth token is empty. Refresh auth token is false.")
			return nil, fmt.Errorf("operator token is empty")
		}

	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(common.OperatorTokenHeaderKey, h.OpLogin.GetOperatorToken())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(LOG_TAG, "Error sending request for rules api :", err)
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	if statusCode == authTokenExpiredCode {
		if refreshAuthToken {
			logger.Error(LOG_TAG, "Operator auth token has expired. Refreshing the auth token.")
			return nil, h.refreshAuthToken(cfg)
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

	logger.Debug(LOG_TAG, "Scenario response body ", body)

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
	latestVersion := "0"
	for _, scenario := range payload.Scenarios {
		ver1, err1 := strconv.ParseInt(latestVersion, 10, 64)
		ver2, err2 := strconv.ParseInt(scenario.Version, 10, 64)
		if err1 != nil || err2 != nil {
			logger.Error(LOG_TAG, "Error while converting versions to int64 for scenario ", scenario.ScenarioId)
			continue
		}

		if ver2 > ver1 {
			latestVersion = scenario.Version
		}

		logger.Debug(LOG_TAG, "Scenario string ", scenario)

		scenarioId := scenario.ScenarioId

		err := h.VersionedStore.SetValue(scenarioId, scenario)
		if err != nil {
			logger.Error(LOG_TAG, "Error while setting filter rule to redis ", err)
			return "", err
		}
	}

	for _, scenarioId := range payload.Deleted {
		err := h.VersionedStore.Delete(scenarioId)
		if err != nil {
			logger.Error(LOG_TAG, "Error while deleting filter id ", scenarioId, " from redis ", err)
			return "", err
		}
	}
	return latestVersion, nil
}

func (h *ScenarioHandler) CleanUpOnkill() error {
	logger.Debug(LOG_TAG, "Kill method in scenario rules.")
	h.ticker.Stop()
	return nil
}
