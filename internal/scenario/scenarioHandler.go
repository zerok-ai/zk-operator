package scenario

import (
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/storage"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-utils-go/rules/model"
)

var authTokenExpiredCode = 401

type ScenarioHandler struct {
	VersionedStore *storage.VersionedStore
	OpLogin        *auth.OperatorLogin
	ticker         *time.Ticker
	config         config.ZkInjectorConfig
	rulesVersion   string
}

type ScenariosApiResponse struct {
	Payload ScenariosObj `json:"payload"`
}

type ScenariosObj struct {
	Scenarios []model.Scenario `json:"scenarios"`
	Deleted   []string         `json:"deleted,omitempty"`
}

func (h *ScenarioHandler) Init(VersionedStore *storage.VersionedStore, OpLogin *auth.OperatorLogin, cfg config.ZkInjectorConfig) {
	h.VersionedStore = VersionedStore
	h.OpLogin = OpLogin
	h.config = cfg
	h.rulesVersion = "0"

	//Creating a timer for periodic scenario
	var duration = time.Duration(cfg.ScenarioSync.PollingInterval) * time.Second
	h.ticker = time.NewTicker(duration)
}

func (h *ScenarioHandler) StartPeriodicSync() {
	h.updateScenarios(h.config)

	for range h.ticker.C {
		fmt.Println("Sync scenarios triggered.")
		h.updateScenarios(h.config)
	}
}

func (h *ScenarioHandler) updateScenarios(cfg config.ZkInjectorConfig) {
	rules, err := h.getScenariosFromZkCloud(cfg)
	if err != nil {
		fmt.Printf("Error while getting rules from zkcloud %v.\n", err)
		return
	}
	latestVersion, err := h.processScenarios(rules)
	if err != nil {
		fmt.Printf("Error while savign rules to redis %v.\n", err)
	} else {
		h.rulesVersion = latestVersion
	}
}

func (h *ScenarioHandler) getScenariosFromZkCloud(cfg config.ZkInjectorConfig) (*ScenariosApiResponse, error) {

	fmt.Println("Get rules from zk cloud.")

	baseURL := "http://" + cfg.ScenarioSync.Host + cfg.ScenarioSync.Path

	//Adding query params
	url := fmt.Sprintf("%s?%s=%s", baseURL, "version", h.rulesVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	fmt.Printf("Current operator token is %v.\n", h.OpLogin.GetOperatorToken())

	if h.OpLogin.GetOperatorToken() == "" {
		return nil, h.refreshAuthToken(cfg)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(common.OperatorTokenHeaderKey, h.OpLogin.GetOperatorToken())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request for rules api :", err)
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	if statusCode == authTokenExpiredCode {
		return nil, h.refreshAuthToken(cfg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response from rules api :", err)
		return nil, err
	}

	var apiResponse ScenariosApiResponse

	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		fmt.Println("Error while unmarshalling rules api response :", err)
		return nil, err
	}

	return &apiResponse, nil
}

func (h *ScenarioHandler) refreshAuthToken(cfg config.ZkInjectorConfig) error {
	err := h.OpLogin.RefreshOperatorToken(func() {
		h.updateScenarios(cfg)
	})
	if err != nil {
		fmt.Printf("Error while refreshing auth token %v.\n", err)
	}
	return err
}

// This method will parse rules and return the largest version found and any error caught.
func (h *ScenarioHandler) processScenarios(rulesApiResponse *ScenariosApiResponse) (string, error) {
	payload := rulesApiResponse.Payload
	latestVersion := "0"
	for _, scenario := range payload.Scenarios {
		ver1, err1 := strconv.ParseInt(latestVersion, 10, 64)
		ver2, err2 := strconv.ParseInt(scenario.Version, 10, 64)
		if err1 != nil || err2 != nil {
			fmt.Printf("Error while converting versions to int64 for scenario %v.\n", scenario.ScenarioId)
			continue
		}

		if ver2 > ver1 {
			latestVersion = scenario.Version
		}

		scenarioString, err := json.Marshal(scenario)
		if err != nil {
			fmt.Printf("Error while converting filter rule to string %v.\n", err)
			return "", err
		}
		scenarioId := scenario.ScenarioId
		err = h.VersionedStore.SetValue(scenarioId, string(scenarioString))
		if err != nil {
			fmt.Printf("Error while setting filter rule to redis %v.\n", err)
			return "", err
		}
	}

	for _, scenarioId := range payload.Deleted {
		err := h.VersionedStore.Delete(scenarioId)
		if err != nil {
			fmt.Printf("Error while deleting filter id %v from redis %v.\n", scenarioId, err)
			return "", err
		}
	}
	return latestVersion, nil
}

func (h *ScenarioHandler) CleanUpOnkill() error {
	fmt.Printf("Kill method in scenario rules.\n")
	h.ticker.Stop()
	return nil
}
